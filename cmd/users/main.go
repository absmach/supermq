// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/mainflux/mainflux/certs/postgres"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	authGrpcClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
	pgClient "github.com/mainflux/mainflux/internal/client/postgres"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/bcrypt"
	"github.com/mainflux/mainflux/users/emailer"
	"github.com/mainflux/mainflux/users/tracing"
	"golang.org/x/sync/errgroup"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	jaegerClient "github.com/mainflux/mainflux/internal/client/jaeger"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/users/api"
	userRepo "github.com/mainflux/mainflux/users/postgres"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	svcName       = "users"
	envPrefix     = "MF_USERS_"
	envPrefixHttp = "MF_USERS_HTTP_"
)

type config struct {
	logLevel      string `env:"MF_USERS_LOG_LEVEL"               envDefault:"debug"`
	adminEmail    string `env:"MF_USERS_ADMIN_EMAIL"             envDefault:""`
	adminPassword string `env:"MF_USERS_ADMIN_PASSWORD"          envDefault:""`
	passRegexText string `env:"MF_USERS_PASS_REGEX"              envDefault:"^.{8,}$"`
	selfRegister  bool   `env:"MF_USERS_ALLOW_SELF_REGISTER"     envDefault:"true"`
	jaegerURL     string `env:"MF_JAEGER_URL"                    envDefault:""`
	resetURL      string `env:"MF_TOKEN_RESET_ENDPOINT"          envDefault:"email.tmpl"`
	passRegex     *regexp.Regexp
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load %s configuration : %s", svcName, err.Error())
	}
	passRegex, err := regexp.Compile(cfg.passRegexText)
	if err != nil {
		log.Fatalf("Invalid password validation rules %s\n", cfg.passRegexText)
	}
	cfg.passRegex = passRegex

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	ec := email.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load email configuration : %s", err.Error())
	}

	db, err := pgClient.Setup(envPrefix, *userRepo.Migration())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	auth, authGrpcClient, authGrpcTracerCloser, authGrpcSecure, err := authGrpcClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer authGrpcClient.Close()
	defer authGrpcTracerCloser.Close()
	logger.Info("Successfully connected to auth grpc server " + authGrpcSecure)

	dbTracer, dbCloser, err := jaegerClient.NewTracer("auth_db", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer dbCloser.Close()

	svc := newService(db, dbTracer, auth, cfg, ec, logger)

	tracer, closer := internalauth.Jaeger("users", cfg.jaegerURL, logger)
	defer closer.Close()

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s HTTP server configuration : %s", svcName, err.Error()))
	}

	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, tracer, logger), logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Users service terminated: %s", err))
	}
}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, tracer opentracing.Tracer, auth mainflux.AuthServiceClient, c config, ec email.Config, logger logger.Logger) users.Service {
	database := userRepo.NewDatabase(db)
	hasher := bcrypt.New()
	userRepo := tracing.UserRepositoryMiddleware(userRepo.NewUserRepo(database), tracer)

	emailer, err := emailer.New(c.resetURL, &ec)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to configure e-mailing util: %s", err.Error()))
	}

	idProvider := uuid.New()

	svc := users.New(userRepo, hasher, auth, emailer, idProvider, c.passRegex)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	if err := createAdmin(svc, userRepo, c, auth); err != nil {
		logger.Error("failed to create admin user: " + err.Error())
		os.Exit(1)
	}

	switch c.selfRegister {
	case true:
		// If MF_USERS_ALLOW_SELF_REGISTER environment variable is "true",
		// everybody can create a new user. Here, check the existence of that
		// policy. If the policy does not exist, create it; otherwise, there is
		// no need to do anything further.
		_, err := auth.Authorize(context.Background(), &mainflux.AuthorizeReq{Obj: "user", Act: "create", Sub: "*"})
		if err != nil {
			// Add a policy that allows anybody to create a user
			apr, err := auth.AddPolicy(context.Background(), &mainflux.AddPolicyReq{Obj: "user", Act: "create", Sub: "*"})
			if err != nil {
				logger.Error("failed to add the policy related to MF_USERS_ALLOW_SELF_REGISTER: " + err.Error())
				os.Exit(1)
			}
			if !apr.GetAuthorized() {
				logger.Error("failed to authorized the policy result related to MF_USERS_ALLOW_SELF_REGISTER: " + errors.ErrAuthorization.Error())
				os.Exit(1)
			}
		}
	default:
		// If MF_USERS_ALLOW_SELF_REGISTER environment variable is "false",
		// everybody cannot create a new user. Therefore, delete a policy that
		// allows everybody to create a new user.
		dpr, err := auth.DeletePolicy(context.Background(), &mainflux.DeletePolicyReq{Obj: "user", Act: "create", Sub: "*"})
		if err != nil {
			logger.Error("failed to delete a policy: " + err.Error())
			os.Exit(1)
		}
		if !dpr.GetDeleted() {
			logger.Error("deleting a policy expected to succeed.")
			os.Exit(1)
		}
	}

	return svc
}

func createAdmin(svc users.Service, userRepo users.UserRepository, c config, auth mainflux.AuthServiceClient) error {
	user := users.User{
		Email:    c.adminEmail,
		Password: c.adminPassword,
	}

	if admin, err := userRepo.RetrieveByEmail(context.Background(), user.Email); err == nil {
		// The admin is already created. Check existence of the admin policy.
		_, err := auth.Authorize(context.Background(), &mainflux.AuthorizeReq{Obj: "authorities", Act: "member", Sub: admin.ID})
		if err != nil {
			apr, err := auth.AddPolicy(context.Background(), &mainflux.AddPolicyReq{Obj: "authorities", Act: "member", Sub: admin.ID})
			if err != nil {
				return err
			}
			if !apr.GetAuthorized() {
				return errors.ErrAuthorization
			}
		}
		return nil
	}

	// Add a policy that allows anybody to create a user
	apr, err := auth.AddPolicy(context.Background(), &mainflux.AddPolicyReq{Obj: "user", Act: "create", Sub: "*"})
	if err != nil {
		return err
	}
	if !apr.GetAuthorized() {
		return errors.ErrAuthorization
	}

	// Create an admin
	uid, err := svc.Register(context.Background(), "", user)
	if err != nil {
		return err
	}

	apr, err = auth.AddPolicy(context.Background(), &mainflux.AddPolicyReq{Obj: "authorities", Act: "member", Sub: uid})
	if err != nil {
		return err
	}
	if !apr.GetAuthorized() {
		return errors.ErrAuthorization
	}

	return nil
}

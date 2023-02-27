// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/go-zoo/bone"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/internal"
	pgClient "github.com/mainflux/mainflux/internal/clients/postgres"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/internal/server"
	grpcserver "github.com/mainflux/mainflux/internal/server/grpc"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users/clients"
	capi "github.com/mainflux/mainflux/users/clients/api"
	"github.com/mainflux/mainflux/users/clients/emailer"
	cpostgres "github.com/mainflux/mainflux/users/clients/postgres"
	ctracing "github.com/mainflux/mainflux/users/clients/tracing"
	"github.com/mainflux/mainflux/users/groups"
	gapi "github.com/mainflux/mainflux/users/groups/api"
	gpostgres "github.com/mainflux/mainflux/users/groups/postgres"
	gtracing "github.com/mainflux/mainflux/users/groups/tracing"
	"github.com/mainflux/mainflux/users/hasher"
	"github.com/mainflux/mainflux/users/jwt"
	"github.com/mainflux/mainflux/users/policies"
	grpcapi "github.com/mainflux/mainflux/users/policies/api/grpc"
	papi "github.com/mainflux/mainflux/users/policies/api/http"
	ppostgres "github.com/mainflux/mainflux/users/policies/postgres"
	ptracing "github.com/mainflux/mainflux/users/policies/tracing"
	clientsPg "github.com/mainflux/mainflux/users/postgres"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	svcName        = "users"
	envPrefix      = "MF_USERS_"
	envPrefixHttp  = "MF_USERS_HTTP_"
	envPrefixGrpc  = "MF_USERS_GRPC_"
	defDB          = "users"
	defSvcHttpPort = "9191"
	defSvcGrpcPort = "9192"
)

type config struct {
	LogLevel        string `env:"MF_USERS_LOG_LEVEL"              envDefault:"info"`
	SecretKey       string `env:"MF_USERS_SECRET_KEY"             envDefault:"secret"`
	AdminEmail      string `env:"MF_USERS_ADMIN_EMAIL"            envDefault:""`
	AdminPassword   string `env:"MF_USERS_ADMIN_PASSWORD"         envDefault:""`
	PassRegexText   string `env:"MF_USERS_PASS_REGEX"             envDefault:"^.{8,}$"`
	AccessDuration  string `env:"MF_USERS_ACCESS_TOKEN_DURATION"  envDefault:"15m"`
	RefreshDuration string `env:"MF_USERS_REFRESH_TOKEN_DURATION" envDefault:"24h"`
	ResetURL        string `env:"MF_TOKEN_RESET_ENDPOINT"         envDefault:"/reset-request"`
	JaegerURL       string `env:"MF_JAEGER_URL"                   envDefault:"localhost:6831"`
	PassRegex       *regexp.Regexp
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}
	passRegex, err := regexp.Compile(cfg.PassRegexText)
	if err != nil {
		log.Fatalf("Invalid password validation rules %s\n", cfg.PassRegexText)
	}
	cfg.PassRegex = passRegex

	logger, err := mflog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to init logger: %s", err.Error()))
	}

	ec := email.Config{}
	if err := env.Parse(&ec); err != nil {
		logger.Fatal(fmt.Sprintf("failed to load email configuration : %s", err.Error()))
	}

	dbConfig := pgClient.Config{Name: defDB}
	db, err := pgClient.SetupWithConfig(envPrefix, *clientsPg.Migration(), dbConfig)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer db.Close()

	tp, err := initJaeger(svcName, cfg.JaegerURL)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to init Jaeger: %s", err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := otel.Tracer(svcName)

	csvc, gsvc, psvc := newService(db, tracer, cfg, ec, logger)

	httpServerConfig := server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		logger.Fatal(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
	}
	m := bone.New()
	hsc := httpserver.New(ctx, cancel, svcName, httpServerConfig, capi.MakeClientsHandler(csvc, m, logger), logger)
	hsg := httpserver.New(ctx, cancel, svcName, httpServerConfig, gapi.MakeGroupsHandler(gsvc, m, logger), logger)
	hsp := httpserver.New(ctx, cancel, svcName, httpServerConfig, papi.MakePolicyHandler(psvc, m, logger), logger)

	// Create new grpc server
	registerAuthServiceServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		policies.RegisterAuthServiceServer(srv, grpcapi.NewServer(csvc, psvc))

	}
	grpcServerConfig := server.Config{Port: defSvcGrpcPort}
	if err := env.Parse(&grpcServerConfig, env.Options{Prefix: envPrefixGrpc, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s gRPC server configuration : %s", svcName, err.Error())
	}
	gs := grpcserver.New(ctx, cancel, svcName, grpcServerConfig, registerAuthServiceServer, logger)

	g.Go(func() error {
		return hsc.Start()
	})
	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hsc, hsg, hsp, gs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Users service terminated: %s", err))
	}
}

func initJaeger(svcName, url string) (*tracesdk.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(exporter)),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(svcName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func newService(db *sqlx.DB, tracer trace.Tracer, c config, ec email.Config, logger mflog.Logger) (clients.Service, groups.GroupService, policies.PolicyService) {
	database := postgres.NewDatabase(db, tracer)
	cRepo := cpostgres.NewClientRepo(database)
	gRepo := gpostgres.NewGroupRepo(database)
	pRepo := ppostgres.NewPolicyRepo(database)

	idp := uuid.New()
	hsr := hasher.New()

	aDuration, err := time.ParseDuration(c.AccessDuration)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse access token duration: %s", err.Error()))
	}
	rDuration, err := time.ParseDuration(c.RefreshDuration)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to parse refresh token duration: %s", err.Error()))
	}
	tokenizer := jwt.NewTokenRepo([]byte(c.SecretKey), aDuration, rDuration)
	tokenizer = jwt.NewTokenRepoMiddleware(tokenizer, tracer)

	emailer, err := emailer.New(c.ResetURL, &ec)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to configure e-mailing util: %s", err.Error()))
	}
	csvc := clients.NewService(cRepo, pRepo, tokenizer, emailer, hsr, idp, c.PassRegex)
	gsvc := groups.NewService(gRepo, pRepo, tokenizer, idp)
	psvc := policies.NewService(pRepo, tokenizer, idp)

	csvc = ctracing.TracingMiddleware(csvc, tracer)
	csvc = capi.LoggingMiddleware(csvc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	csvc = capi.MetricsMiddleware(csvc, counter, latency)

	gsvc = gtracing.TracingMiddleware(gsvc, tracer)
	gsvc = gapi.LoggingMiddleware(gsvc, logger)
	counter, latency = internal.MakeMetrics("groups", "api")
	gsvc = gapi.MetricsMiddleware(gsvc, counter, latency)

	psvc = ptracing.TracingMiddleware(psvc, tracer)
	psvc = papi.LoggingMiddleware(psvc, logger)
	counter, latency = internal.MakeMetrics("policies", "api")
	psvc = papi.MetricsMiddleware(psvc, counter, latency)

	if err := createAdmin(c, cRepo, hsr, csvc); err != nil {
		logger.Error(fmt.Sprintf("Failed to create admin client: %s", err))
	}
	return csvc, gsvc, psvc
}

func createAdmin(c config, crepo clients.ClientRepository, hsr clients.Hasher, svc clients.Service) error {
	id, err := uuid.New().ID()
	if err != nil {
		return err
	}
	hash, err := hsr.Hash(c.AdminPassword)
	if err != nil {
		return err
	}

	client := clients.Client{
		ID:   id,
		Name: "admin",
		Credentials: clients.Credentials{
			Identity: c.AdminEmail,
			Secret:   hash,
		},
		Metadata: clients.Metadata{
			"role": "admin",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role:      clients.AdminRole,
		Status:    clients.EnabledStatus,
	}

	if _, err := crepo.RetrieveByIdentity(context.Background(), client.Credentials.Identity); err == nil {
		return nil
	}

	// Create an admin
	if _, err = crepo.Save(context.Background(), client); err != nil {
		return err
	}
	_, err = svc.IssueToken(context.Background(), c.AdminEmail, c.AdminPassword)
	if err != nil {
		return err
	}

	return nil
}

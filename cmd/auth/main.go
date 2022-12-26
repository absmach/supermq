package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth"
	api "github.com/mainflux/mainflux/auth/api"
	grpcapi "github.com/mainflux/mainflux/auth/api/grpc"
	httpapi "github.com/mainflux/mainflux/auth/api/http"
	"github.com/mainflux/mainflux/auth/jwt"
	"github.com/mainflux/mainflux/auth/keto"
	authRepo "github.com/mainflux/mainflux/auth/postgres"
	"github.com/mainflux/mainflux/auth/tracing"
	"github.com/mainflux/mainflux/internal"
	grpcClient "github.com/mainflux/mainflux/internal/client/grpc"
	jaegerClient "github.com/mainflux/mainflux/internal/client/jaeger"
	pgClient "github.com/mainflux/mainflux/internal/client/postgres"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	grpcserver "github.com/mainflux/mainflux/internal/server/grpc"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/opentracing/opentracing-go"
	acl "github.com/ory/keto/proto/ory/keto/acl/v1alpha1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	svcName          = "auth"
	defListenAddress = ""
	envPrefix        = "MF_AUTH_"
	envPrefixHttp    = "MF_AUTH_HTTP_"
	envPrefixGrpc    = "MF_AUTH_GRPC_"
)

type config struct {
	LogLevel      string        `env:"MF_AUTH_LOG_LEVEL"             envDefault:"debug"`
	Secret        string        `env:"MF_AUTH_SECRET"                envDefault:"auth"`
	JaegerURL     string        `env:"MF_JAEGER_URL"                 envDefault:""`
	KetoReadHost  string        `env:"MF_KETO_READ_REMOTE_HOST"      envDefault:"mainflux-keto"`
	KetoWriteHost string        `env:"MF_KETO_WRITE_REMOTE_HOST"     envDefault:"mainflux-keto"`
	KetoWritePort string        `env:"MF_KETO_READ_REMOTE_PORT"      envDefault:"4466"`
	KetoReadPort  string        `env:"MF_KETO_WRITE_REMOTE_PORT"     envDefault:"4467"`
	LoginDuration time.Duration `env:"MF_AUTH_LOGIN_TOKEN_DURATION"  envDefault:"10h"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// create auth service configurations
	cfg := config{}
	// load auth service configurations from environment
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s configuration : %s", svcName, err.Error()))
	}

	// create new logger
	logger, err := logger.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	///////////////// POSTGRES CLIENT /////////////////////////
	// create new postgres config
	dbConfig := pgClient.Config{}
	// load postgres config from environment
	if err := env.Parse(&dbConfig, env.Options{Prefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s database configuration : %s", svcName, err.Error()))
	}
	// create new postgres client
	db, err := pgClient.SetupDB(dbConfig, *authRepo.Migration())
	if err != nil {
		log.Fatalf("Failed to setup %s database : %s", svcName, err.Error())
	}
	defer db.Close()

	///////////////// AUTH SERVICE /////////////////////////
	// create new tracker for database
	dbTracer, dbCloser, err := jaegerClient.NewTracer("auth_db", cfg.JaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer dbCloser.Close()

	// create new jaeger tracer
	tracer, closer, err := jaegerClient.NewTracer("auth", cfg.JaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer closer.Close()

	//create keto read grpc client
	readerConn, _, err := grpcClient.Connect(grpcClient.Config{ClientTLS: false, URL: fmt.Sprintf("%s:%s", cfg.KetoReadHost, cfg.KetoReadPort)})
	if err != nil {
		log.Fatalf("Failed to connect to keto gRPC: %s", err.Error())
	}
	//create keto write grpc client
	writerConn, _, err := grpcClient.Connect(grpcClient.Config{ClientTLS: false, URL: fmt.Sprintf("%s:%s", cfg.KetoWriteHost, cfg.KetoWritePort)})
	if err != nil {
		log.Fatalf("Failed to connect to keto gRPC: %s", err.Error())
	}

	svc := newService(db, dbTracer, cfg.Secret, logger, readerConn, writerConn, cfg.LoginDuration)

	///////////////// HTTP SERVER //////////////////////////
	// create new http server config
	httpServerConfig := server.Config{}
	// load http server config from environment variables
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s HTTP server configuration : %s", svcName, err.Error()))
	}
	// create new http server
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svc, tracer, logger), logger)

	///////////////// GRPC SERVER //////////////////////////
	// create new grpc server config
	grpcServerConfig := server.Config{}
	// load grpc server config from environment variables
	if err := env.Parse(&grpcServerConfig, env.Options{Prefix: envPrefixGrpc, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s gRPC server configuration : %s", svcName, err.Error()))
	}
	registerAuthServiceServer := func(srv *grpc.Server) {
		mainflux.RegisterAuthServiceServer(srv, grpcapi.NewServer(tracer, svc))
	}
	// create new grpc server
	gs := grpcserver.New(ctx, cancel, svcName, grpcServerConfig, registerAuthServiceServer, logger)

	//Start servers
	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, gs)
	})
	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Authentication service terminated: %s", err))
	}
}

func newService(db *sqlx.DB, tracer opentracing.Tracer, secret string, logger logger.Logger, readerConn, writerConn *grpc.ClientConn, duration time.Duration) auth.Service {
	database := authRepo.NewDatabase(db)
	keysRepo := tracing.New(authRepo.New(database), tracer)

	groupsRepo := authRepo.NewGroupRepo(database)
	groupsRepo = tracing.GroupRepositoryMiddleware(tracer, groupsRepo)

	pa := keto.NewPolicyAgent(acl.NewCheckServiceClient(readerConn), acl.NewWriteServiceClient(writerConn), acl.NewReadServiceClient(readerConn))

	idProvider := uuid.New()
	t := jwt.New(secret)

	svc := auth.New(keysRepo, groupsRepo, idProvider, t, pa, duration)
	svc = api.LoggingMiddleware(svc, logger)

	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}

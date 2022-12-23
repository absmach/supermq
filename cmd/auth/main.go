package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
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
	envPreFix        = "MF_AUTH_"
)

type config struct {
	LogLevel      string        `env:"MF_AUTH_LOG_LEVEL"             default:"debug"`
	HttpPort      string        `env:"MF_AUTH_HTTP_PORT"             default:"6180"`
	GrpcPort      string        `env:"MF_AUTH_GRPC_PORT"             default:"6181"`
	Secret        string        `env:"MF_AUTH_SECRET"                default:"auth"`
	ServerCert    string        `env:"MF_AUTH_SERVER_CERT"           default:""`
	ServerKey     string        `env:"MF_AUTH_SERVER_KEY"            default:""`
	JaegerURL     string        `env:"MF_JAEGER_URL"                 default:""`
	KetoReadHost  string        `env:"MF_KETO_READ_REMOTE_HOST"      default:"mainflux-keto"`
	KetoWriteHost string        `env:"MF_KETO_WRITE_REMOTE_HOST"     default:"mainflux-keto"`
	KetoWritePort string        `env:"MF_KETO_READ_REMOTE_PORT"      default:"4466"`
	KetoReadPort  string        `env:"MF_KETO_WRITE_REMOTE_PORT"     default:"4467"`
	LoginDuration time.Duration `env:"MF_AUTH_LOGIN_TOKEN_DURATION"  default:"10h"`
}

func main() {
	cfg := config{}
	dbConfig := pgClient.Config{}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s configuration : %s", svcName, err.Error()))
	}
	if err := env.Parse(&dbConfig, env.Options{Prefix: envPreFix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s database configuration : %s", svcName, err.Error()))
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	logger, err := logger.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db, err := pgClient.SetupDB(dbConfig, *authRepo.Migration())
	if err != nil {
		log.Fatalf("Failed to setup %s database : %s", svcName, err.Error())
	}
	defer db.Close()

	tracer, closer, err := jaegerClient.NewTracer("auth", cfg.JaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer closer.Close()

	dbTracer, dbCloser, err := jaegerClient.NewTracer("auth_db", cfg.JaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer dbCloser.Close()

	readerConn, _, err := grpcClient.Connect(grpcClient.Config{ClientTLS: false, URL: fmt.Sprintf("%s:%s", cfg.KetoReadHost, cfg.KetoReadPort)})
	if err != nil {
		log.Fatalf("Failed to connect to keto gRPC: %s", err.Error())
	}
	writerConn, _, err := grpcClient.Connect(grpcClient.Config{ClientTLS: false, URL: fmt.Sprintf("%s:%s", cfg.KetoWriteHost, cfg.KetoWritePort)})
	if err != nil {
		log.Fatalf("Failed to connect to keto gRPC: %s", err.Error())
	}

	svc := newService(db, dbTracer, cfg.Secret, logger, readerConn, writerConn, cfg.LoginDuration)

	registerAuthServiceServer := func(srv *grpc.Server) {
		mainflux.RegisterAuthServiceServer(srv, grpcapi.NewServer(tracer, svc))
	}

	hs := httpserver.New(ctx, cancel, svcName, defListenAddress, cfg.HttpPort, httpapi.MakeHandler(svc, tracer, logger), cfg.ServerCert, cfg.ServerKey, logger)
	gs := grpcserver.New(ctx, cancel, svcName, defListenAddress, cfg.GrpcPort, registerAuthServiceServer, cfg.ServerCert, cfg.ServerKey, logger)
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

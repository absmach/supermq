package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-redis/redis/v8"
	"github.com/go-zoo/bone"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/clients/clients"
	capi "github.com/mainflux/mainflux/clients/clients/api"
	cpostgres "github.com/mainflux/mainflux/clients/clients/postgres"
	ctracing "github.com/mainflux/mainflux/clients/clients/tracing"
	"github.com/mainflux/mainflux/clients/groups"
	gapi "github.com/mainflux/mainflux/clients/groups/api"
	gpostgres "github.com/mainflux/mainflux/clients/groups/postgres"
	gtracing "github.com/mainflux/mainflux/clients/groups/tracing"
	"github.com/mainflux/mainflux/clients/policies"
	grpcapi "github.com/mainflux/mainflux/clients/policies/api/grpc"
	papi "github.com/mainflux/mainflux/clients/policies/api/http"
	ppostgres "github.com/mainflux/mainflux/clients/policies/postgres"
	ppracing "github.com/mainflux/mainflux/clients/policies/tracing"
	"github.com/mainflux/mainflux/clients/postgres"
	authClient "github.com/mainflux/mainflux/internal/clients/grpc/auth"
	redisClient "github.com/mainflux/mainflux/internal/clients/redis"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	rediscache "github.com/mainflux/mainflux/things/redis"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	stopWaitTime       = 5 * time.Second
	svcName            = "things"
	envPrefix          = "MF_THINGS_"
	envPrefixCache     = "MF_THINGS_CACHE_"
	envPrefixES        = "MF_THINGS_ES_"
	envPrefixHttp      = "MF_THINGS_HTTP_"
	envPrefixAuthHttp  = "MF_THINGS_AUTH_HTTP_"
	envPrefixAuthGrpc  = "MF_THINGS_AUTH_GRPC_"
	defDB              = "things"
	defSvcHttpPort     = "8182"
	defSvcAuthHttpPort = "8989"
	defSvcAuthGrpcPort = "8181"
)

const (
	defLogLevel      = "debug"
	defSecretKey     = "clientsecret"
	defAdminIdentity = "admin@example.com"
	defAdminSecret   = "12345678"
	defHTTPPort      = "9191"
	defGRPCPort      = "9192"
	defServerCert    = ""
	defServerKey     = ""
	defJaegerURL     = "http://localhost:6831"

	envLogLevel      = "MF_CLIENTS_LOG_LEVEL"
	envSecretKey     = "MF_CLIENTS_SECRET_KEY"
	envAdminIdentity = "MF_CLIENTS_ADMIN_EMAIL"
	envAdminSecret   = "MF_CLIENTS_ADMIN_PASSWORD"
	envHTTPPort      = "MF_CLIENTS_HTTP_PORT"
	envGRPCPort      = "MF_CLIENTS_GRPC_PORT"
	envServerCert    = "MF_CLIENTS_SERVER_CERT"
	envServerKey     = "MF_CLIENTS_SERVER_KEY"
	envJaegerURL     = "MF_CLIENTS_JAEGER_URL"
)

type config struct {
	logLevel      string
	secretKey     string
	adminIdentity string
	adminSecret   string
	dbConfig      postgres.Config
	httpPort      string
	grpcPort      string
	serverCert    string
	serverKey     string
	jaegerURL     string
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	tp, err := initJaeger(svcName, cfg.jaegerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := otel.Tracer(svcName)

	// Setup new redis cache client
	cacheClient, err := redisClient.Setup(envPrefixCache)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer cacheClient.Close()

	// Setup new auth grpc client
	auth, authHandler, err := authClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer authHandler.Close()
	logger.Info("Successfully connected to auth grpc server " + authHandler.Secure())

	csvc, gsvc, psvc := newService(db, auth, cacheClient, tracer, cfg, logger)

	g.Go(func() error {
		return startHTTPServer(ctx, csvc, gsvc, psvc, cfg.httpPort, cfg.serverCert, cfg.serverKey, logger)
	})
	g.Go(func() error {
		return startGRPCServer(ctx, psvc, cfg.grpcPort, cfg.serverCert, cfg.serverKey, logger)
	})

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("%s service shutdown by signal: %s", svcName, sig))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func loadConfig() config {
	return config{
		logLevel:      mainflux.Env(envLogLevel, defLogLevel),
		secretKey:     mainflux.Env(envSecretKey, defSecretKey),
		adminIdentity: mainflux.Env(envAdminIdentity, defAdminIdentity),
		adminSecret:   mainflux.Env(envAdminSecret, defAdminSecret),
		httpPort:      mainflux.Env(envHTTPPort, defHTTPPort),
		grpcPort:      mainflux.Env(envGRPCPort, defGRPCPort),
		serverCert:    mainflux.Env(envServerCert, defServerCert),
		serverKey:     mainflux.Env(envServerKey, defServerKey),
		jaegerURL:     mainflux.Env(envJaegerURL, defJaegerURL),
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

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, auth mainflux.AuthServiceClient, cacheClient *redis.Client, tracer trace.Tracer, c config, logger logger.Logger) (clients.Service, groups.GroupService, policies.PolicyService) {
	database := postgres.NewDatabase(db, tracer)
	cRepo := cpostgres.NewClientRepo(database)
	gRepo := gpostgres.NewGroupRepo(database)
	pRepo := ppostgres.NewPolicyRepo(database)

	idp := uuid.New()

	chanCache := rediscache.NewChannelCache(cacheClient)

	thingCache := rediscache.NewThingCache(cacheClient)

	csvc := clients.NewService(auth, cRepo, thingCache, pRepo, idp)
	gsvc := groups.NewService(auth, gRepo, pRepo, idp)
	psvc := policies.NewService(auth, pRepo, thingCache, chanCache, idp)

	csvc = ctracing.TracingMiddleware(csvc, tracer)
	csvc = capi.LoggingMiddleware(csvc, logger)
	csvc = capi.MetricsMiddleware(
		csvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: svcName,
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: svcName,
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	gsvc = gtracing.TracingMiddleware(gsvc, tracer)
	gsvc = gapi.LoggingMiddleware(gsvc, logger)
	gsvc = gapi.MetricsMiddleware(
		gsvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "clients_groups",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "clients_groups",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	psvc = ppracing.TracingMiddleware(psvc, tracer)
	psvc = papi.LoggingMiddleware(psvc, logger)
	psvc = papi.MetricsMiddleware(
		psvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "client_policies",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "client_policies",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	return csvc, gsvc, psvc
}

func startHTTPServer(ctx context.Context, csvc clients.Service, gsvc groups.Service, psvc policies.Service, port string, certFile string, keyFile string, logger logger.Logger) error {
	p := fmt.Sprintf(":%s", port)
	errCh := make(chan error)
	m := bone.New()
	capi.MakeClientsHandler(csvc, m, logger)
	gapi.MakeGroupsHandler(gsvc, m, logger)
	papi.MakePolicyHandler(psvc, m, logger)
	server := &http.Server{Addr: p, Handler: m}

	switch {
	case certFile != "" || keyFile != "":
		logger.Info(fmt.Sprintf("Clients service started using https, cert %s key %s, exposed port %s", certFile, keyFile, port))
		go func() {
			errCh <- server.ListenAndServeTLS(certFile, keyFile)
		}()
	default:
		logger.Info(fmt.Sprintf("Clients service started using http, exposed port %s", port))
		go func() {
			errCh <- server.ListenAndServe()
		}()
	}

	select {
	case <-ctx.Done():
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), stopWaitTime)
		defer cancelShutdown()
		if err := server.Shutdown(ctxShutdown); err != nil {
			logger.Error(fmt.Sprintf("Clients service error occurred during shutdown at %s: %s", p, err))
			return fmt.Errorf("clients service occurred during shutdown at %s: %w", p, err)
		}
		logger.Info(fmt.Sprintf("Clients service shutdown of http at %s", p))
		return nil
	case err := <-errCh:
		return err
	}

}

func startGRPCServer(ctx context.Context, svc policies.Service, port string, certFile string, keyFile string, logger logger.Logger) error {
	p := fmt.Sprintf(":%s", port)
	errCh := make(chan error)

	listener, err := net.Listen("tcp", p)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	var server *grpc.Server
	switch {
	case certFile != "" || keyFile != "":
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load auth certificates: %w", err)
		}
		logger.Info(fmt.Sprintf("Clients gRPC service started using https on port %s with cert %s key %s", port, certFile, keyFile))
		server = grpc.NewServer(grpc.Creds(creds))
	default:
		logger.Info(fmt.Sprintf("Clients gRPC service started using http on port %s", port))
		server = grpc.NewServer()
	}

	reflection.Register(server)
	policies.RegisterAuthServiceServer(server, grpcapi.NewServer(svc))
	logger.Info(fmt.Sprintf("Clients gRPC service started, exposed port %s", port))
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		c := make(chan bool)
		go func() {
			defer close(c)
			server.GracefulStop()
		}()
		select {
		case <-c:
		case <-time.After(stopWaitTime):
		}
		logger.Info(fmt.Sprintf("Authentication gRPC service shutdown at %s", p))
		return nil
	case err := <-errCh:
		return err
	}
}

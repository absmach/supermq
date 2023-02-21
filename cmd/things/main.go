package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-zoo/bone"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/internal"
	authClient "github.com/mainflux/mainflux/internal/clients/grpc/auth"
	pgClient "github.com/mainflux/mainflux/internal/clients/postgres"
	redisClient "github.com/mainflux/mainflux/internal/clients/redis"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	grpcserver "github.com/mainflux/mainflux/internal/server/grpc"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things/clients"
	capi "github.com/mainflux/mainflux/things/clients/api"
	cpostgres "github.com/mainflux/mainflux/things/clients/postgres"
	redisthcache "github.com/mainflux/mainflux/things/clients/redis"
	ctracing "github.com/mainflux/mainflux/things/clients/tracing"
	"github.com/mainflux/mainflux/things/groups"
	gapi "github.com/mainflux/mainflux/things/groups/api"
	gpostgres "github.com/mainflux/mainflux/things/groups/postgres"
	gtracing "github.com/mainflux/mainflux/things/groups/tracing"
	tpolicies "github.com/mainflux/mainflux/things/policies"
	grpcapi "github.com/mainflux/mainflux/things/policies/api/grpc"
	papi "github.com/mainflux/mainflux/things/policies/api/http"
	ppostgres "github.com/mainflux/mainflux/things/policies/postgres"
	redischcache "github.com/mainflux/mainflux/things/policies/redis"
	ppracing "github.com/mainflux/mainflux/things/policies/tracing"
	"github.com/mainflux/mainflux/things/postgres"
	thingsPg "github.com/mainflux/mainflux/things/postgres"
	upolicies "github.com/mainflux/mainflux/users/policies"
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
	defSvcAuthGrpcPort = "8181"
)

type config struct {
	LogLevel        string `env:"MF_THINGS_LOG_LEVEL"          envDefault:"info"`
	StandaloneEmail string `env:"MF_THINGS_STANDALONE_EMAIL"   envDefault:""`
	StandaloneToken string `env:"MF_THINGS_STANDALONE_TOKEN"   envDefault:""`
	JaegerURL       string `env:"MF_JAEGER_URL"                envDefault:"localhost:6831"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// Create new things configuration
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := mflog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}
	// Create new database for things
	dbConfig := pgClient.Config{Name: defDB}
	db, err := pgClient.SetupWithConfig(envPrefix, *thingsPg.Migration(), dbConfig)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer db.Close()

	tp, err := initJaeger(svcName, cfg.JaegerURL)
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
		logger.Fatal(err.Error())
	}
	defer cacheClient.Close()

	// Setup new auth grpc client
	auth, authHandler, err := authClient.Setup(envPrefix, cfg.JaegerURL)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer authHandler.Close()
	logger.Info("Successfully connected to auth grpc server " + authHandler.Secure())

	csvc, gsvc, psvc := newService(db, auth, cacheClient, tracer, logger)

	mux := bone.New()
	httpServerConfig := server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		logger.Fatal(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err))
	}
	hsc := httpserver.New(ctx, cancel, "things-clients", httpServerConfig, capi.MakeClientsHandler(csvc, mux, logger), logger)

	httpServerConfig = server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s gRPC server configuration : %s", svcName, err.Error())
	}
	hsg := httpserver.New(ctx, cancel, "things-groups", httpServerConfig, gapi.MakeGroupsHandler(gsvc, mux, logger), logger)

	httpServerConfig = server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s gRPC server configuration : %s", svcName, err.Error())
	}
	hsp := httpserver.New(ctx, cancel, "things-policies", httpServerConfig, papi.MakePolicyHandler(csvc, psvc, mux, logger), logger)

	registerThingsServiceServer := func(srv *grpc.Server) {
		reflection.Register(srv)
		tpolicies.RegisterThingsServiceServer(srv, grpcapi.NewServer(csvc, psvc))
	}
	grpcServerConfig := server.Config{Port: defSvcAuthGrpcPort}
	if err := env.Parse(&grpcServerConfig, env.Options{Prefix: envPrefixAuthGrpc, AltPrefix: envPrefix}); err != nil {
		logger.Fatal(fmt.Sprintf("failed to load %s gRPC server configuration : %s", svcName, err))
	}
	gs := grpcserver.New(ctx, cancel, svcName, grpcServerConfig, registerThingsServiceServer, logger)

	//Start all servers
	g.Go(func() error {
		return hsp.Start()
	})
	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hsc, hsg, hsp, gs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
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

func newService(db *sqlx.DB, auth upolicies.AuthServiceClient, cacheClient *redis.Client, tracer trace.Tracer, logger logger.Logger) (clients.Service, groups.GroupService, tpolicies.PolicyService) {
	database := postgres.NewDatabase(db, tracer)
	cRepo := cpostgres.NewClientRepo(database)
	gRepo := gpostgres.NewGroupRepo(database)
	pRepo := ppostgres.NewPolicyRepo(database)

	idp := uuid.New()

	policyCache := redischcache.NewPolicyCache(cacheClient)
	thingCache := redisthcache.NewThingCache(cacheClient)

	csvc := clients.NewService(auth, cRepo, thingCache, idp)
	gsvc := groups.NewService(auth, gRepo, pRepo, idp)
	psvc := tpolicies.NewService(auth, pRepo, thingCache, policyCache, idp)

	csvc = ctracing.TracingMiddleware(csvc, tracer)
	csvc = capi.LoggingMiddleware(csvc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	csvc = capi.MetricsMiddleware(csvc, counter, latency)

	gsvc = gtracing.TracingMiddleware(gsvc, tracer)
	gsvc = gapi.LoggingMiddleware(gsvc, logger)
	counter, latency = internal.MakeMetrics(fmt.Sprintf("%s_groups", svcName), "api")
	gsvc = gapi.MetricsMiddleware(gsvc, counter, latency)

	psvc = ppracing.TracingMiddleware(psvc, tracer)
	psvc = papi.LoggingMiddleware(psvc, logger)
	counter, latency = internal.MakeMetrics(fmt.Sprintf("%s_policies", svcName), "api")
	psvc = papi.MetricsMiddleware(psvc, counter, latency)

	return csvc, gsvc, psvc
}

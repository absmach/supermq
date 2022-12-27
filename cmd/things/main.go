// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	authClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
	pgClient "github.com/mainflux/mainflux/internal/client/postgres"
	redisClient "github.com/mainflux/mainflux/internal/client/redis"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	grpcserver "github.com/mainflux/mainflux/internal/server/grpc"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/things/api"
	authgrpcapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	authhttpapi "github.com/mainflux/mainflux/things/api/auth/http"
	thhttpapi "github.com/mainflux/mainflux/things/api/things/http"
	thingsPg "github.com/mainflux/mainflux/things/postgres"
	rediscache "github.com/mainflux/mainflux/things/redis"
	"github.com/mainflux/mainflux/things/tracing"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	svcName           = "things"
	envPrefix         = "MF_THINGS_"
	envPrefixCache    = "MF_THINGS_CACHE_"
	envPrefixES       = "MF_THINGS_ES_"
	envPrefixHttp     = "MF_THINGS_HTTP_"
	envPrefixGrpc     = "MF_THINGS_GRPC_"
	envPrefixAuth     = "MF_THINGS_AUTH_"
	envPrefixAuthHttp = "MF_THINGS_AUTH_HTTP_"
	envPrefixAuthGrpc = "MF_THINGS_AUTH_GRPC_"
)

type config struct {
	logLevel        string `env:"MF_THINGS_LOG_LEVEL"          envDefault:"debug"`
	standaloneEmail string `env:"MF_THINGS_STANDALONE_EMAIL"   envDefault:"debug"`
	standaloneToken string `env:"MF_THINGS_STANDALONE_TOKEN"   envDefault:"debug"`
	jaegerURL       string `env:"MF_JAEGER_URL"                envDefault:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// create new things configuration
	cfg := config{}
	// load things configuration from environment variables
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s service configuration : %s", svcName, err.Error())
	}

	// create new logger
	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// Setup new database for things
	db, err := pgClient.Setup(envPrefix, *thingsPg.Migration())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	// Setup new redis cache client
	cacheClient, err := redisClient.Setup(envPrefixCache)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer cacheClient.Close()

	// Setup new redis event store client
	esClient, err := redisClient.Setup(envPrefixES)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer esClient.Close()

	// Setup new auth grpc client
	auth, authGrpcClient, authGrpcTracerCloser, authGrpcSecure, err := authClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer authGrpcClient.Close()
	defer authGrpcTracerCloser.Close()
	logger.Info("Successfully connected to auth grpc server " + authGrpcSecure)

	// create tracer for things database
	dbTracer, dbCloser := internalauth.Jaeger("things_db", cfg.jaegerURL, logger)
	defer dbCloser.Close()

	// create tracer for things cache
	cacheTracer, cacheCloser := internalauth.Jaeger("things_cache", cfg.jaegerURL, logger)
	defer cacheCloser.Close()

	//create new things service
	svc := newService(auth, dbTracer, cacheTracer, db, cacheClient, esClient, logger)

	// create tracer for HTTP handler things
	thingsTracer, thingsCloser := internalauth.Jaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	/////////////////// THINGS HTTP SERVER /////////////////////
	// create new HTTP  server config
	httpServerConfig := server.Config{}
	// load grpc server config from environment variables
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s gRPC server configuration : %s", svcName, err.Error()))
	}

	hs1 := httpserver.New(ctx, cancel, "thing-http", httpServerConfig, thhttpapi.MakeHandler(thingsTracer, svc, logger), logger)

	/////////////////// THINGS AUTH HTTP SERVER /////////////////////
	// create new things auth http server config
	authHttpServerConfig := server.Config{}
	// load grpc server config from environment variables
	if err := env.Parse(&authHttpServerConfig, env.Options{Prefix: envPrefixAuthHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s gRPC server configuration : %s", svcName, err.Error()))
	}
	hs2 := httpserver.New(ctx, cancel, "auth-http", authHttpServerConfig, authhttpapi.MakeHandler(thingsTracer, svc, logger), logger)

	/////////////////// THINGS AUTH GRPC SERVER /////////////////////
	// register things grpc service server
	registerThingsServiceServer := func(srv *grpc.Server) {
		mainflux.RegisterThingsServiceServer(srv, authgrpcapi.NewServer(thingsTracer, svc))

	}
	// create new grpc server config
	grpcServerConfig := server.Config{}
	// load grpc server config from environment variables
	if err := env.Parse(&grpcServerConfig, env.Options{Prefix: envPrefixAuthGrpc, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s gRPC server configuration : %s", svcName, err.Error()))
	}
	//Create new things auth grpc server
	gs := grpcserver.New(ctx, cancel, svcName, grpcServerConfig, registerThingsServiceServer, logger)

	//Start all servers
	g.Go(func() error {
		return hs1.Start()
	})
	g.Go(func() error {
		return hs2.Start()
	})
	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs1, hs2, gs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Things service terminated: %s", err))
	}
}

func newService(auth mainflux.AuthServiceClient, dbTracer opentracing.Tracer, cacheTracer opentracing.Tracer, db *sqlx.DB, cacheClient *redis.Client, esClient *redis.Client, logger logger.Logger) things.Service {
	database := thingsPg.NewDatabase(db)

	thingsRepo := thingsPg.NewThingRepository(database)
	thingsRepo = tracing.ThingRepositoryMiddleware(dbTracer, thingsRepo)

	channelsRepo := thingsPg.NewChannelRepository(database)
	channelsRepo = tracing.ChannelRepositoryMiddleware(dbTracer, channelsRepo)

	chanCache := rediscache.NewChannelCache(cacheClient)
	chanCache = tracing.ChannelCacheMiddleware(cacheTracer, chanCache)

	thingCache := rediscache.NewThingCache(cacheClient)
	thingCache = tracing.ThingCacheMiddleware(cacheTracer, thingCache)
	idProvider := uuid.New()

	svc := things.New(auth, thingsRepo, channelsRepo, chanCache, thingCache, idProvider)
	svc = rediscache.NewEventStoreMiddleware(svc, esClient)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}

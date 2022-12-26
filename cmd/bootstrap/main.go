// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	r "github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/bootstrap"
	api "github.com/mainflux/mainflux/bootstrap/api"
	bootstrapRepo "github.com/mainflux/mainflux/bootstrap/postgres"
	rediscons "github.com/mainflux/mainflux/bootstrap/redis/consumer"
	redisprod "github.com/mainflux/mainflux/bootstrap/redis/producer"
	"github.com/mainflux/mainflux/internal"
	grpcClient "github.com/mainflux/mainflux/internal/client/grpc"
	jaegerClient "github.com/mainflux/mainflux/internal/client/jaeger"
	pgClient "github.com/mainflux/mainflux/internal/client/postgres"
	redisClient "github.com/mainflux/mainflux/internal/client/redis"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"golang.org/x/sync/errgroup"
)

const (
	svcName                 = "bootstrap"
	envPrefix               = "MF_BOOTSTRAP_"
	envPrefixHttp           = "MF_BOOTSTRAP_HTTP_"
	envThingsAuthGrpcPrefix = "MF_THINGS_AUTH_GRPC_"
	envAuthGrpcPrefix       = "MF_AUTH_GRPC_"
)

type config struct {
	logLevel       string `env:"MF_BOOTSTRAP_LOG_LEVEL"        envDefault:"debug"`
	encKey         []byte `env:"MF_BOOTSTRAP_ENCRYPT_KEY"      envDefault:"12345678910111213141516171819202"`
	thingsURL      string `env:"MF_THINGS_URL"                 envDefault:"http://localhost"`
	esConsumerName string `env:"MF_BOOTSTRAP_EVENT_CONSUMER"   envDefault:"bootstrap"`
	jaegerURL      string `env:"MF_JAEGER_URL"                 envDefault:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s configuration : %s", svcName, err.Error()))
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
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
	db, err := pgClient.SetupDB(dbConfig, *bootstrapRepo.Migration())
	if err != nil {
		log.Fatalf("Failed to setup %s database : %s", svcName, err.Error())
	}
	defer db.Close()

	///////////////// EVENT STORE REDIS CLIENT /////////////////////////
	// create new redis client config for bootstrap event store
	bootstrapESConfig := redisClient.Config{}
	if err := env.Parse(&bootstrapESConfig, env.Options{Prefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s bootstrap event store configuration : %s", svcName, err.Error()))
	}
	// create new redis client for bootstrap event store
	esClient, err := redisClient.Connect(bootstrapESConfig)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to setup %s bootstrap event store redis client : %s", svcName, err.Error()))
	}
	defer esClient.Close()

	///////////////// AUTH - GRPC CLIENT /////////////////////////
	// create new auth grpc client config
	authGrpcConfig := grpcClient.Config{}
	// load auth grpc client config from environment
	if err := env.Parse(&authGrpcConfig, env.Options{Prefix: envAuthGrpcPrefix, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s configuration : %s", svcName, err.Error()))
	}
	// create new auth grpc client
	authConn, secure, err := grpcClient.Connect(authGrpcConfig)
	if err != nil {
		log.Fatalf("Failed to connect with auth gRPC : %s ", err.Error())
	}
	defer authConn.Close()
	secureMsg := "without TLS"
	if secure {
		secureMsg = "with TLS"
	}
	logger.Info(fmt.Sprintf("Connected to auth gRPC %s", secureMsg))
	// create new auth grpc client tracer
	authTracer, authCloser, err := jaegerClient.NewTracer("auth", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer authCloser.Close()
	// create new auth grpc client api
	auth := authapi.NewClient(authTracer, authConn, authGrpcConfig.Timeout)

	///////////////// BOOTSTRAP SERVICE /////////////////////////
	svc := newService(auth, db, logger, esClient, cfg)

	///////////////// HTTP SERVER /////////////////////////
	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s HTTP server configuration : %s", svcName, err.Error()))
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, bootstrap.NewConfigReader(cfg.encKey), logger), logger)
	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	///////////////// SUBSCRIBE TO THINGS EVENT STORE/////////////////////////
	// create new redis client config for things event store
	thingESConfig := redisClient.Config{}
	if err := env.Parse(&thingESConfig, env.Options{Prefix: envPrefix, AltPrefix: envThingsAuthGrpcPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s things event store configuration : %s", svcName, err.Error()))
	}
	// create new redis client for things event store
	thingsESClient, err := redisClient.Connect(thingESConfig)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to setup %s things event store redis client : %s", svcName, err.Error()))
	}
	defer thingsESClient.Close()
	// subscribe to things event store
	go subscribeToThingsES(svc, thingsESClient, cfg.esConsumerName, logger)

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Bootstrap service terminated: %s", err))
	}
}

func newService(auth mainflux.AuthServiceClient, db *sqlx.DB, logger logger.Logger, esClient *r.Client, cfg config) bootstrap.Service {
	thingsRepo := bootstrapRepo.NewConfigRepository(db, logger)

	config := mfsdk.Config{
		ThingsURL: cfg.thingsURL,
	}

	sdk := mfsdk.NewSDK(config)

	svc := bootstrap.New(auth, thingsRepo, sdk, cfg.encKey)
	svc = redisprod.NewEventStoreMiddleware(svc, esClient)
	svc = api.NewLoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}

func subscribeToThingsES(svc bootstrap.Service, client *r.Client, consumer string, logger logger.Logger) {
	eventStore := rediscons.NewEventStore(svc, client, consumer, logger)
	logger.Info("Subscribed to Redis Event Store")
	if err := eventStore.Subscribe(context.Background(), "mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("Bootstrap service failed to subscribe to event sourcing: %s", err))
	}
}

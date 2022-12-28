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
	"github.com/mainflux/mainflux/bootstrap"
	api "github.com/mainflux/mainflux/bootstrap/api"
	bootstrapPg "github.com/mainflux/mainflux/bootstrap/postgres"
	rediscons "github.com/mainflux/mainflux/bootstrap/redis/consumer"
	redisprod "github.com/mainflux/mainflux/bootstrap/redis/producer"
	"github.com/mainflux/mainflux/internal"
	authClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
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
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatal(err.Error())
	}

	///////////////// POSTGRES CLIENT /////////////////////////
	// create new postgres client
	db, err := pgClient.Setup(envPrefix, *bootstrapPg.Migration())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	///////////////// EVENT STORE REDIS CLIENT /////////////////////////
	// create new redis client for bootstrap event store
	esClient, err := redisClient.Setup(envPrefix)
	if err != nil {
		log.Fatalf("failed to setup %s bootstrap event store redis client : %s", svcName, err.Error())
	}
	defer esClient.Close()

	///////////////// AUTH - GRPC CLIENT /////////////////////////
	// create new auth grpc client api
	auth, authGrpcClient, authGrpcTracerCloser, authGrpcSecure, err := authClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer authGrpcClient.Close()
	defer authGrpcTracerCloser.Close()
	logger.Info("Successfully connected to auth grpc server " + authGrpcSecure)

	///////////////// BOOTSTRAP SERVICE /////////////////////////
	svc := newService(auth, db, logger, esClient, cfg)

	///////////////// HTTP SERVER /////////////////////////
	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s HTTP server configuration : %s", svcName, err.Error())
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, bootstrap.NewConfigReader(cfg.encKey), logger), logger)
	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	///////////////// SUBSCRIBE TO THINGS EVENT STORE/////////////////////////
	thingsESClient, err := redisClient.Setup(envPrefix)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer thingsESClient.Close()
	// subscribe to things event store
	go subscribeToThingsES(svc, thingsESClient, cfg.esConsumerName, logger)

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Bootstrap service terminated: %s", err))
	}
}

func newService(auth mainflux.AuthServiceClient, db *sqlx.DB, logger logger.Logger, esClient *r.Client, cfg config) bootstrap.Service {
	repoConfig := bootstrapPg.NewConfigRepository(db, logger)

	config := mfsdk.Config{
		ThingsURL: cfg.thingsURL,
	}

	sdk := mfsdk.NewSDK(config)

	svc := bootstrap.New(auth, repoConfig, sdk, cfg.encKey)
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

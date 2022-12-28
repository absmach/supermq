// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal"
	authClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
	jaegerClient "github.com/mainflux/mainflux/internal/client/jaeger"
	mongoClient "github.com/mainflux/mainflux/internal/client/mongo"
	redisClient "github.com/mainflux/mainflux/internal/client/redis"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"github.com/mainflux/mainflux/pkg/uuid"
	localusers "github.com/mainflux/mainflux/things/standalone"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/api"
	twapi "github.com/mainflux/mainflux/twins/api/http"
	twmongodb "github.com/mainflux/mainflux/twins/mongodb"
	rediscache "github.com/mainflux/mainflux/twins/redis"
	"github.com/mainflux/mainflux/twins/tracing"
	opentracing "github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "twins"
	queue          = "twins"
	envPrefix      = "MF_TWINS_"
	envPrefixHttp  = "MF_TWINS_HTTP_"
	envPrefixCache = "MF_TWINS_CACHE_"
)

type config struct {
	logLevel        string `env:"MF_TWINS_LOG_LEVEL"          envDefault:"debug"`
	standaloneEmail string `env:"MF_TWINS_STANDALONE_EMAIL"   envDefault:""`
	standaloneToken string `env:"MF_TWINS_STANDALONE_TOKEN"   envDefault:""`
	channelID       string `env:"MF_TWINS_CHANNEL_ID"         envDefault:""`
	jaegerURL       string `env:"MF_JAEGER_URL"               envDefault:""`
	brokerURL       string `env:"MF_BROKER_URL"               envDefault:"nats://localhost:4222"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg, env.Options{Prefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}
	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	cacheClient, err := redisClient.Setup(envPrefixCache)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer cacheClient.Close()

	cacheTracer, cacheCloser, err := jaegerClient.NewTracer("twins_cache", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer cacheCloser.Close()

	db, err := mongoClient.Setup(envPrefix)
	if err != nil {
		log.Fatalf("Failed to setup postgres database : %s", err.Error())
	}

	dbTracer, dbCloser, err := jaegerClient.NewTracer("twins_db", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer dbCloser.Close()

	var auth mainflux.AuthServiceClient
	switch cfg.standaloneEmail != "" && cfg.standaloneToken != "" {
	case true:
		auth = localusers.NewAuthService(cfg.standaloneEmail, cfg.standaloneToken)
	default:
		authServiceClient, authGrpcClient, authTracerCloser, authGrpcSecure, err := authClient.Setup(envPrefix, cfg.jaegerURL)
		if err != nil {
			log.Fatal(err.Error())
		}
		defer authGrpcClient.Close()
		defer authTracerCloser.Close()
		auth = authServiceClient
		logger.Info("Successfully connected to auth grpc server " + authGrpcSecure)
	}

	pubSub, err := brokers.NewPubSub(cfg.brokerURL, queue, logger)
	if err != nil {
		log.Fatalf("Failed to connect to message broker: %s", err.Error())
	}
	defer pubSub.Close()

	svc := newService(svcName, pubSub, cfg.channelID, auth, dbTracer, db, cacheTracer, cacheClient, logger)

	tracer, closer, err := jaegerClient.NewTracer("twins", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer closer.Close()

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s HTTP server configuration : %s", svcName, err.Error()))
	}

	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, twapi.MakeHandler(tracer, svc, logger), logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Twins service terminated: %s", err))
	}
}

func newService(id string, ps messaging.PubSub, chanID string, users mainflux.AuthServiceClient, dbTracer opentracing.Tracer, db *mongo.Database, cacheTracer opentracing.Tracer, cacheClient *redis.Client, logger logger.Logger) twins.Service {
	twinRepo := twmongodb.NewTwinRepository(db)
	twinRepo = tracing.TwinRepositoryMiddleware(dbTracer, twinRepo)

	stateRepo := twmongodb.NewStateRepository(db)
	stateRepo = tracing.StateRepositoryMiddleware(dbTracer, stateRepo)

	idProvider := uuid.New()
	twinCache := rediscache.NewTwinCache(cacheClient)
	twinCache = tracing.TwinCacheMiddleware(cacheTracer, twinCache)

	svc := twins.New(ps, users, twinRepo, twinCache, stateRepo, idProvider, chanID, logger)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)
	err := ps.Subscribe(id, brokers.SubjectAllChannels, handle(logger, chanID, svc))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	return svc
}

func handle(logger logger.Logger, chanID string, svc twins.Service) handlerFunc {
	return func(msg messaging.Message) error {
		if msg.Channel == chanID {
			return nil
		}

		if err := svc.SaveStates(&msg); err != nil {
			logger.Error(fmt.Sprintf("State save failed: %s", err))
			return err
		}

		return nil
	}
}

type handlerFunc func(msg messaging.Message) error

func (h handlerFunc) Handle(msg messaging.Message) error {
	return h(msg)
}

func (h handlerFunc) Cancel() error {
	return nil
}

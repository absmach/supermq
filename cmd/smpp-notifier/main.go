// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/consumers/notifiers/api"
	notifierPg "github.com/mainflux/mainflux/consumers/notifiers/postgres"
	"github.com/mainflux/mainflux/consumers/notifiers/smpp"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"golang.org/x/sync/errgroup"

	mfsmpp "github.com/mainflux/mainflux/consumers/notifiers/smpp"
	"github.com/mainflux/mainflux/consumers/notifiers/tracing"
	authClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
	jaegerClient "github.com/mainflux/mainflux/internal/client/jaeger"
	pgClient "github.com/mainflux/mainflux/internal/client/postgres"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"github.com/mainflux/mainflux/pkg/ulid"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	svcName       = "smpp-notifier"
	envPrefix     = "MF_SMPP_NOTIFIER_"
	envPrefixHttp = "MF_SMPP_NOTIFIER_HTTP_"
)

type config struct {
	logLevel   string `env:"MF_SMPP_NOTIFIER_LOG_LEVEL"   envDefault:"debug"`
	from       string `env:"MF_SMPP_NOTIFIER_FROM_ADDR"   envDefault:""`
	configPath string `env:"MF_SMPP_NOTIFIER_CONFIG_PATH" envDefault:"/config.toml"`
	brokerURL  string `env:"MF_BROKER_URL"                envDefault:"nats://localhost:4222"`
	jaegerURL  string `env:"MF_JAEGER_URL"                envDefault:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db, err := pgClient.Setup(envPrefix, *notifierPg.Migration())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	smppConfig := smpp.Config{}
	if err := env.Parse(&smppConfig); err != nil {
		log.Fatalf("Failed to load SMPP configuration from environment : %s", err.Error())
	}

	pubSub, err := brokers.NewPubSub(cfg.brokerURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to message broker: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	auth, authGrpcClient, authGrpcTracerCloser, authGrpcSecure, err := authClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err)
	}
	defer authGrpcClient.Close()
	defer authGrpcTracerCloser.Close()
	logger.Info("Successfully connected to auth grpc server " + authGrpcSecure)

	tracer, closer, err := jaegerClient.NewTracer("smpp-notifier", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer closer.Close()

	dbTracer, dbCloser, err := jaegerClient.NewTracer("smpp-notifier_db", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to init Jaeger: %s", err.Error())
	}
	defer dbCloser.Close()

	svc := newService(db, dbTracer, auth, cfg, smppConfig, logger)

	if err = consumers.Start(svcName, pubSub, svc, cfg.configPath, logger); err != nil {
		log.Fatalf("Failed to create Postgres writer: %s", err.Error())
	}

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("Failed to load %s HTTP server configuration : %s", svcName, err.Error())
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, tracer, logger), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("SMPP notifier service terminated: %s", err))
	}

}

func newService(db *sqlx.DB, tracer opentracing.Tracer, auth mainflux.AuthServiceClient, c config, sc smpp.Config, logger logger.Logger) notifiers.Service {
	database := notifierPg.NewDatabase(db)
	repo := tracing.New(notifierPg.New(database), tracer)
	idp := ulid.New()
	notifier := mfsmpp.New(sc)
	svc := notifiers.New(auth, repo, idp, notifier, c.from)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("notifier", "smpp")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}

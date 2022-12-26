// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/consumers/writers/timescale"
	"github.com/mainflux/mainflux/internal"
	pgClient "github.com/mainflux/mainflux/internal/client/postgres"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"golang.org/x/sync/errgroup"
)

const (
	svcName       = "timescaledb-writer"
	envPrefix     = "MF_TIMESCALE_WRITER_"
	envPrefixHttp = "MF_TIMESCALE_WRITER_HTTP_"
)

type config struct {
	brokerURL  string `env:"MF_BROKER_URL"                   envDefault:"nats://localhost:4222" `
	logLevel   string `env:"MF_TIMESCALE_WRITER_LOG_LEVEL"   envDefault:"debug"`
	configPath string `env:"MF_TIMESCALE_WRITER_CONFIG_PATH" envDefault:"/config.toml"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load %s service configuration : %s", svcName, err.Error())
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatal(err.Error())
	}

	db, err := pgClient.Setup(envPrefix, *timescale.Migration())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	repo := newService(db, logger)

	pubSub, err := brokers.NewPubSub(cfg.brokerURL, "", logger)
	if err != nil {
		log.Fatalf("failed to connect to message broker: %s", err.Error())
	}
	defer pubSub.Close()

	if err = consumers.Start(svcName, pubSub, repo, cfg.configPath, logger); err != nil {
		log.Fatalf("failed to create Timescale writer: %s", err.Error())
	}

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("Failed to load %s HTTP server configuration : %s", svcName, err.Error())
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svcName), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Timescale writer service terminated: %s", err))
	}
}

func newService(db *sqlx.DB, logger logger.Logger) consumers.Consumer {
	svc := timescale.New(db)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("timescale", "message_writer")
	svc = api.MetricsMiddleware(svc, counter, latency)
	return svc
}

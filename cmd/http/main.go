// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	adapter "github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/http/api"
	"github.com/mainflux/mainflux/internal"
	thingsClient "github.com/mainflux/mainflux/internal/clients/grpc/things"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"github.com/mainflux/mainflux/things/policies"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "http_adapter"
	envPrefix      = "MF_HTTP_ADAPTER_"
	envPrefixHttp  = "MF_HTTP_ADAPTER_HTTP_"
	defSvcHttpPort = "8180"
)

type config struct {
	LogLevel  string `env:"MF_HTTP_ADAPTER_LOG_LEVEL"   envDefault:"info"`
	BrokerURL string `env:"MF_BROKER_URL"               envDefault:"nats://localhost:4222"`
	JaegerURL string `env:"MF_JAEGER_URL"               envDefault:"http://jaeger:14268/api/traces"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := mflog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	tc, tcHandler, err := thingsClient.Setup(envPrefix, cfg.JaegerURL)
	if err != nil {
		logger.Fatal(err.Error())
	}
	defer tcHandler.Close()
	logger.Info("Successfully connected to things grpc server " + tcHandler.Secure())

	pub, err := brokers.NewPublisher(cfg.BrokerURL)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to connect to message broker: %s", err))
	}
	defer pub.Close()

	svc := newService(pub, tc, logger)

	httpServerConfig := server.Config{Port: defSvcHttpPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		logger.Fatal(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, logger), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("HTTP adapter service terminated: %s", err))
	}
}

func newService(pub messaging.Publisher, tc policies.ThingsServiceClient, logger mflog.Logger) adapter.Service {
	svc := adapter.New(pub, tc)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)
	return svc
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mainflux/mainflux"
	"golang.org/x/sync/errgroup"

	"github.com/mainflux/mainflux/internal"
	thingsClient "github.com/mainflux/mainflux/internal/client/grpc/things"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	logger "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	adapter "github.com/mainflux/mainflux/ws"
	"github.com/mainflux/mainflux/ws/api"
)

const (
	svcName       = "ws-adapter"
	envPrefix     = "MF_WS_ADAPTER_"
	envPrefixHttp = "MF_WS_ADAPTER_HTTP_"

	defPort              = "8190"
	defBrokerURL         = "nats://localhost:4222"
	defLogLevel          = "error"
	defClientTLS         = "false"
	defCACerts           = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"

	envPort          = "MF_WS_ADAPTER_PORT"
	envBrokerURL     = "MF_BROKER_URL"
	envLogLevel      = "MF_WS_ADAPTER_LOG_LEVEL"
	envClientTLS     = "MF_WS_ADAPTER_CLIENT_TLS"
	envCACerts       = "MF_WS_ADAPTER_CA_CERTS"
	envJaegerURL     = "MF_JAEGER_URL"
	envThingsAuthURL = "MF_THINGS_AUTH_GRPC_URL"
	envThingsTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel  string `env:"MF_WS_ADAPTER_LOG_LEVEL"   envDefault:"debug" `
	brokerURL string `env:"MF_BROKER_URL"             envDefault:"nats://localhost:4222" `
	jaegerURL string `env:"MF_JAEGER_URL"             envDefault:"" `
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s service configuration : %s", svcName, err.Error())
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tc, thingsGrpcClient, thingsTracerCloser, thingsGrpcSecure, err := thingsClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer thingsGrpcClient.Close()
	defer thingsTracerCloser.Close()
	logger.Info("Successfully connected to things grpc server " + thingsGrpcSecure)

	nps, err := brokers.NewPubSub(cfg.brokerURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to message broker: %s", err))
		os.Exit(1)
	}
	defer nps.Close()

	svc := newService(tc, nps, logger)

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s HTTP server configuration : %s", svcName, err.Error())
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, logger), logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("WS adapter service terminated: %s", err))
	}
}

func newService(tc mainflux.ThingsServiceClient, nps messaging.PubSub, logger logger.Logger) adapter.Service {
	svc := adapter.New(tc, nps)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("ws_adapter", "api")
	svc = api.MetricsMiddleware(svc, counter, latency)
	return svc
}

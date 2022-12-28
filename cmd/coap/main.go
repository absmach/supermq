// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/mainflux/mainflux/coap"
	"github.com/mainflux/mainflux/coap/api"
	"github.com/mainflux/mainflux/internal"
	thingsClient "github.com/mainflux/mainflux/internal/client/grpc/things"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	coapserver "github.com/mainflux/mainflux/internal/server/coap"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	logger "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"golang.org/x/sync/errgroup"
)

const (
	svcName       = "coap_adapter"
	envPrefix     = "MF_COAP_ADAPTER_"
	envPrefixHttp = "MF_COAP_ADAPTER_HTTP_"
	envPrefixCoap = "MF_COAP_ADAPTER_COAP_"
)

type config struct {
	logLevel  string `env:"MF_INFLUX_READER_LOG_LEVEL"  envDefault:"debug"`
	brokerURL string `env:"MF_BROKER_URL"               envDefault:"nats://localhost:4222"`
	jaegerURL string `env:"MF_JAEGER_URL"               envDefault:""`
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

	tc, thingsGrpcClient, thingsTracerCloser, thingsGrpcSecure, err := thingsClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer thingsGrpcClient.Close()
	defer thingsTracerCloser.Close()
	logger.Info("Successfully connected to things grpc server " + thingsGrpcSecure)

	nps, err := brokers.NewPubSub(cfg.brokerURL, "", logger)
	if err != nil {
		log.Fatalf("Failed to connect to message broker: %s", err.Error())
	}
	defer nps.Close()

	svc := coap.New(tc, nps)

	svc = api.LoggingMiddleware(svc, logger)

	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s HTTP server configuration : %s", svcName, err.Error())
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHTTPHandler(), logger)

	coapServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixCoap, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s CoAP server configuration : %s", svcName, err.Error())
	}
	cs := coapserver.New(ctx, cancel, svcName, coapServerConfig, api.MakeCoAPHandler(svc, logger), logger)

	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return cs.Start()
	})
	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, cs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("CoAP adapter service terminated: %s", err))
	}
}

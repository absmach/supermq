// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package main contains mqtt-adapter main function to start the mqtt-adapter service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	jaegerClient "github.com/mainflux/mainflux/internal/clients/jaeger"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/modbus"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	brokerstracing "github.com/mainflux/mainflux/pkg/messaging/brokers/tracing"
	"github.com/mainflux/mainflux/pkg/uuid"
	"golang.org/x/sync/errgroup"
)

const (
	svcName     = "mqtt"
	envPrefixES = "MF_MQTT_ADAPTER_ES_"
)

type config struct {
	LogLevel      string `env:"MF_MQTT_ADAPTER_LOG_LEVEL"                    envDefault:"info"`
	JaegerURL     string `env:"MF_JAEGER_URL"                                envDefault:"http://localhost:14268/api/traces"`
	BrokerURL     string `env:"MF_BROKER_URL"                                envDefault:"nats://localhost:4222"`
	SendTelemetry bool   `env:"MF_SEND_TELEMETRY"                            envDefault:"true"`
	InstanceID    string `env:"MF_MQTT_ADAPTER_INSTANCE_ID"                  envDefault:""`
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

	var exitCode int
	defer mflog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	tp, err := jaegerClient.NewProvider(svcName, cfg.JaegerURL, cfg.InstanceID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	nps, err := brokers.NewPubSub(cfg.BrokerURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer nps.Close()
	nps = brokerstracing.NewPubSub(server.Config{}, tracer, nps)

	svc := modbus.New(logger)

	if err := svc.Read(ctx, "modbus", nps, nps); err != nil {
		logger.Error(fmt.Sprintf("failed to forward message broker messages: %s", err))
		exitCode = 1
		return
	}

	/*if cfg.SendTelemetry {
		chc := chclient.New(svcName, mainflux.Version, logger, cancel)
		go chc.CallHome(ctx)
	}*/

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("mProxy shutdown by signal: %s", sig))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("mProxy terminated: %s", err))
	}
}

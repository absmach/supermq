// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package main contains modbus-adapter main function to start the modbus-adapter service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	chclient "github.com/mainflux/callhome/pkg/client"
	"github.com/mainflux/mainflux"
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
	svcName     = "modbus"
	envPrefixES = "MF_MODBUS_ADAPTER_"
)

type config struct {
	LogLevel      string `env:"MF_MODBUS_ADAPTER_LOG_LEVEL"   envDefault:"info"`
	JaegerURL     string `env:"MF_JAEGER_URL"               envDefault:"http://localhost:14268/api/traces"`
	BrokerURL     string `env:"MF_BROKER_URL"               envDefault:"nats://localhost:4222"`
	SendTelemetry bool   `env:"MF_SEND_TELEMETRY"           envDefault:"true"`
	InstanceID    string `env:"MF_MODBUS_ADAPTER_INSTANCE_ID" envDefault:""`
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

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Fatal(fmt.Sprintf("failed to generate instanceID: %s", err))
		}
	}

	tp, err := jaegerClient.NewProvider(svcName, cfg.JaegerURL, cfg.InstanceID)
	if err != nil {
		logger.Fatal(fmt.Sprintf("Failed to init Jaeger: %s", err))
	}
	var exitCode int
	defer mflog.ExitWithError(&exitCode)

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

	if err := svc.Read(ctx, svcName, nps, nps); err != nil {
		logger.Error(fmt.Sprintf("failed to forward read messages: %v", err))
		exitCode = 1
		return
	}

	if err := svc.Write(ctx, svcName, nps, nps); err != nil {
		logger.Error(fmt.Sprintf("failed to forward write messages: %v", err))
		exitCode = 1
		return
	}

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, mainflux.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("modbus shutdown by signal: %s", sig))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("modbus service terminated: %s", err))
	}
}

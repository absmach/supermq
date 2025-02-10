// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package main contains rabbitmq-auth main function to start the rabbitmq-auth service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	chclient "github.com/absmach/callhome/pkg/client"
	"github.com/absmach/supermq"
	smqlog "github.com/absmach/supermq/logger"
	"github.com/absmach/supermq/pkg/grpcclient"
	jaegerclient "github.com/absmach/supermq/pkg/jaeger"
	"github.com/absmach/supermq/pkg/server"
	httpserver "github.com/absmach/supermq/pkg/server/http"
	"github.com/absmach/supermq/pkg/uuid"
	"github.com/absmach/supermq/rabbitmq-auth/api"
	"github.com/caarlos0/env/v11"
	"golang.org/x/sync/errgroup"
)

const (
	svcName          = "rabbitmq-auth"
	envPrefixClients = "SMQ_CLIENTS_AUTH_GRPC_"
	envPrefixHTTP    = "SMQ_RABBITMQ_AUTH_HTTP_"
	defSvcHTTPPort   = "9011"
)

type config struct {
	LogLevel      string  `env:"SMQ_RABBITMQ_AUTH_LOG_LEVEL"   envDefault:"info"`
	JaegerURL     url.URL `env:"SMQ_JAEGER_URL"                envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry bool    `env:"SMQ_SEND_TELEMETRY"            envDefault:"true"`
	InstanceID    string  `env:"SMQ_RABBITMQ_AUTH_INSTANCE_ID" envDefault:""`
	TraceRatio    float64 `env:"SMQ_JAEGER_TRACE_RATIO"        envDefault:"1.0"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := smqlog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
	}

	var exitCode int
	defer smqlog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	tp, err := jaegerclient.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
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

	clientsClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&clientsClientCfg, env.Options{Prefix: envPrefixClients}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s auth configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	clientsClient, clientsHandler, err := grpcclient.SetupClientsClient(ctx, clientsClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer clientsHandler.Close()
	logger.Info("Clients service gRPC client successfully connected to clients gRPC server " + clientsHandler.Secure())

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, supermq.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	httpServerConfig := server.Config{
		Port: defSvcHTTPPort,
	}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}
	httpServerConfig.Host = ""

	httpSrv := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(logger, cfg.InstanceID, clientsClient), logger)

	g.Go(func() error {
		return httpSrv.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, httpSrv)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("mProxy terminated: %s", err))
	}
}

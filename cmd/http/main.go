// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package main contains http-adapter main function to start the http-adapter service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	chclient "github.com/mainflux/callhome/pkg/client"
	"github.com/mainflux/mainflux"
	adapter "github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/http/api"
	"github.com/mainflux/mainflux/internal"
	authapi "github.com/mainflux/mainflux/internal/clients/grpc/auth"
	jaegerclient "github.com/mainflux/mainflux/internal/clients/jaeger"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	brokerstracing "github.com/mainflux/mainflux/pkg/messaging/brokers/tracing"
	"github.com/mainflux/mainflux/pkg/messaging/handler"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/ws/tracing"
	mproxy "github.com/mainflux/mproxy/pkg/http"
	"github.com/mainflux/mproxy/pkg/session"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "http_adapter"
	envPrefix      = "MF_HTTP_ADAPTER_"
	defSvcHTTPPort = "80"
	targetHTTPPort = "81"
	targetHTTPHost = ""
)

type config struct {
	LogLevel      string `env:"MF_HTTP_ADAPTER_LOG_LEVEL"   envDefault:"info"`
	BrokerURL     string `env:"MF_MESSAGE_BROKER_URL"       envDefault:"nats://localhost:4222"`
	JaegerURL     string `env:"MF_JAEGER_URL"               envDefault:"http://jaeger:14268/api/traces"`
	SendTelemetry bool   `env:"MF_SEND_TELEMETRY"           envDefault:"true"`
	InstanceID    string `env:"MF_HTTP_ADAPTER_INSTANCE_ID" envDefault:""`
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

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefix}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	auth, aHandler, err := authapi.SetupAuthz("authz")
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer aHandler.Close()

	logger.Info("Successfully connected to things grpc server " + aHandler.Secure())

	tp, err := jaegerclient.NewProvider(svcName, cfg.JaegerURL, cfg.InstanceID)
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

	pub, err := brokers.NewPublisher(ctx, cfg.BrokerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer pub.Close()
	pub = brokerstracing.NewPublisher(httpServerConfig, tracer, pub)

	svc := newService(pub, auth, logger, tracer)
	targetServerCfg := server.Config{Port: targetHTTPPort, Host: targetHTTPHost}

	hs := httpserver.New(ctx, cancel, svcName, targetServerCfg, api.MakeHandler(cfg.InstanceID), logger)

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, mainflux.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	g.Go(func() error {
		return hs.Start()
	})

	svc := newService(pub, tc, logger, tracer)
	switch {
	case httpServerConfig.CertFile != "" || httpServerConfig.KeyFile != "":
		logger.Info(fmt.Sprintf("%s service https server listening at %s:%s with TLS cert %s and key %s", svcName, httpServerConfig.Host, httpServerConfig.Port, httpServerConfig.CertFile, httpServerConfig.KeyFile))
		if err := proxyHTTPS(ctx, httpServerConfig, logger, svc); err != nil {
			logger.Error(fmt.Sprintf("failed to start proxy server with error: %v", err))
		}
	default:
		logger.Info(fmt.Sprintf("%s service http server listening at %s:%s without TLS", svcName, httpServerConfig.Host, httpServerConfig.Port))
		if err := proxyHTTP(ctx, httpServerConfig, logger, svc); err != nil {
			logger.Error(fmt.Sprintf("failed to start proxy server with error: %v", err))
		}
	}

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("HTTP adapter service terminated: %s", err))
	}
}

func newService(pub messaging.Publisher, tc mainflux.AuthzServiceClient, logger mflog.Logger, tracer trace.Tracer) adapter.Service {
	svc := adapter.New(pub, tc)
	svc = tracing.New(tracer, svc)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = handler.MetricsMiddleware(svc, counter, latency)
	return svc
}

func proxyHTTP(ctx context.Context, cfg server.Config, logger mflog.Logger, handler session.Handler) error {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	target := fmt.Sprintf("%s:%s", targetHTTPHost, targetHTTPPort)
	mp, err := mproxy.NewProxy(address, target, handler, logger)
	if err != nil {
		return err
	}

	errCh := make(chan error)
	go func() {
		errCh <- mp.Listen()
	}()

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy HTTP shutdown at %s", target))
		return nil
	case err := <-errCh:
		return err
	}
}

func proxyHTTPS(ctx context.Context, cfg server.Config, logger mflog.Logger, handler session.Handler) error {
	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	target := fmt.Sprintf("%s:%s", targetHTTPHost, targetHTTPPort)
	mp, err := mproxy.NewProxy(address, target, handler, logger)
	if err != nil {
		return err
	}

	errCh := make(chan error)

	go func() {
		errCh <- mp.ListenTLS(cfg.CertFile, cfg.KeyFile)
	}()

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy MQTT WS shutdown at %s", target))
		return nil
	case err := <-errCh:
		return err
	}
}

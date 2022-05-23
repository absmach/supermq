// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/coap"
	"github.com/mainflux/mainflux/coap/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/internal/apiutil/mfserver"
	"github.com/mainflux/mainflux/internal/apiutil/mfserver/httpserver"
	logger "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	gocoap "github.com/plgd-dev/go-coap/v2"
	"golang.org/x/sync/errgroup"
)

const (
	svcName      = "coap_adapter"
	stopWaitTime = 5 * time.Second

	defPort              = "5683"
	defNatsURL           = "nats://localhost:4222"
	defLogLevel          = "error"
	defClientTLS         = "false"
	defCACerts           = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"

	envPort              = "MF_COAP_ADAPTER_PORT"
	envNatsURL           = "MF_NATS_URL"
	envLogLevel          = "MF_COAP_ADAPTER_LOG_LEVEL"
	envClientTLS         = "MF_COAP_ADAPTER_CLIENT_TLS"
	envCACerts           = "MF_COAP_ADAPTER_CA_CERTS"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	port              string
	natsURL           string
	logLevel          string
	clientTLS         bool
	caCerts           string
	jaegerURL         string
	thingsAuthURL     string
	thingsAuthTimeout time.Duration
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	conn := apiutil.ConnectToThings(cfg.clientTLS, cfg.caCerts, cfg.thingsAuthURL, svcName, logger)
	defer conn.Close()

	thingsTracer, thingsCloser := apiutil.InitJaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)

	nps, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nps.Close()

	svc := coap.New(tc, nps)

	svc = api.LoggingMiddleware(svc, logger)

	counter, latency := apiutil.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	g.Go(func() error {
		return startCOAPServer(ctx, cfg, svc, nil, logger)
	})

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHTTPHandler(), "", "", logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return mfserver.ServerStopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("CoAP adapter service terminated: %s", err))
	}
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		natsURL:           mainflux.Env(envNatsURL, defNatsURL),
		port:              mainflux.Env(envPort, defPort),
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: authTimeout,
	}
}

func startCOAPServer(ctx context.Context, cfg config, svc coap.Service, auth mainflux.ThingsServiceClient, l logger.Logger) error {
	p := fmt.Sprintf(":%s", cfg.port)
	errCh := make(chan error)
	l.Info(fmt.Sprintf("CoAP adapter service started, exposed port %s", cfg.port))
	go func() {
		errCh <- gocoap.ListenAndServe("udp", p, api.MakeCoAPHandler(svc, l))
	}()
	select {
	case <-ctx.Done():
		l.Info(fmt.Sprintf("CoAP adapter service shutdown of http at %s", p))
		return nil
	case err := <-errCh:
		return err
	}
}

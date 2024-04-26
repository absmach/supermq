// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package main contains mqtt-adapter main function to start the mqtt-adapter service.
package main

import (
	"context"
	// "crypto/tls"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	chclient "github.com/absmach/callhome/pkg/client"
	"github.com/absmach/magistrala"
	jaegerclient "github.com/absmach/magistrala/internal/clients/jaeger"
	"github.com/absmach/magistrala/internal/server"
	mglog "github.com/absmach/magistrala/logger"
	"github.com/absmach/magistrala/mqtt"
	"github.com/absmach/magistrala/mqtt/events"
	mqtttracing "github.com/absmach/magistrala/mqtt/tracing"
	"github.com/absmach/magistrala/pkg/auth"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/absmach/magistrala/pkg/messaging/brokers"
	brokerstracing "github.com/absmach/magistrala/pkg/messaging/brokers/tracing"
	"github.com/absmach/magistrala/pkg/messaging/handler"
	mqttpub "github.com/absmach/magistrala/pkg/messaging/mqtt"
	"github.com/absmach/magistrala/pkg/uuid"
	"github.com/absmach/mproxy"
	mproxymqtt "github.com/absmach/mproxy/pkg/mqtt"
	"github.com/absmach/mproxy/pkg/mqtt/websocket"
	"github.com/absmach/mproxy/pkg/session"
	"github.com/caarlos0/env/v10"
	"github.com/cenkalti/backoff/v4"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "mqtt"
	envPrefixAuthz = "MG_THINGS_AUTH_GRPC_"
	envPrefixMQTTNoTLS = "MG_MQTT_ADAPTER_MQTT_WITHOUT_TLS_"
	envPrefixMQTTTLS = "MG_MQTT_ADAPTER_MQTT_WITH_TLS_"
	envPrefixWSNoTLS = "MG_MQTT_ADAPTER_MQTT_WS_WITHOUT_TLS_"
	envPrefixWSTLS = "MG_MQTT_ADAPTER_MQTT_WS_WITH_TLS_"
)

type config struct {
	LogLevel              string        `env:"MG_MQTT_ADAPTER_LOG_LEVEL"                    envDefault:"info"`
	MQTTPort              string        `env:"MG_MQTT_ADAPTER_MQTT_PORT"                    envDefault:"1883"`
	MQTTTargetHost        string        `env:"MG_MQTT_ADAPTER_MQTT_TARGET_HOST"             envDefault:"localhost"`
	MQTTTargetPort        string        `env:"MG_MQTT_ADAPTER_MQTT_TARGET_PORT"             envDefault:"1883"`
	MQTTForwarderTimeout  time.Duration `env:"MG_MQTT_ADAPTER_FORWARDER_TIMEOUT"            envDefault:"30s"`
	MQTTTargetHealthCheck string        `env:"MG_MQTT_ADAPTER_MQTT_TARGET_HEALTH_CHECK"     envDefault:""`
	MQTTQoS               uint8         `env:"MG_MQTT_ADAPTER_MQTT_QOS"                     envDefault:"1"`
	HTTPPort              string        `env:"MG_MQTT_ADAPTER_WS_PORT"                      envDefault:"8080"`
	HTTPTargetHost        string        `env:"MG_MQTT_ADAPTER_WS_TARGET_HOST"               envDefault:"localhost"`
	HTTPTargetPort        string        `env:"MG_MQTT_ADAPTER_WS_TARGET_PORT"               envDefault:"8080"`
	HTTPTargetPath        string        `env:"MG_MQTT_ADAPTER_WS_TARGET_PATH"               envDefault:"/mqtt"`
	Instance              string        `env:"MG_MQTT_ADAPTER_INSTANCE"                     envDefault:""`
	JaegerURL             url.URL       `env:"MG_JAEGER_URL"                                envDefault:"http://localhost:14268/api/traces"`
	BrokerURL             string        `env:"MG_MESSAGE_BROKER_URL"                        envDefault:"nats://localhost:4222"`
	SendTelemetry         bool          `env:"MG_SEND_TELEMETRY"                            envDefault:"true"`
	InstanceID            string        `env:"MG_MQTT_ADAPTER_INSTANCE_ID"                  envDefault:""`
	ESURL                 string        `env:"MG_ES_URL"                                    envDefault:"nats://localhost:4222"`
	TraceRatio            float64       `env:"MG_JAEGER_TRACE_RATIO"                        envDefault:"1.0"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}
	log.Println(cfg)
	logger, err := mglog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
	}

	var exitCode int
	defer mglog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	if cfg.MQTTTargetHealthCheck != "" {
		notify := func(e error, next time.Duration) {
			logger.Info(fmt.Sprintf("Broker not ready: %s, next try in %s", e.Error(), next))
		}

		err := backoff.RetryNotify(healthcheck(cfg), backoff.NewExponentialBackOff(), notify)
		if err != nil {
			logger.Error(fmt.Sprintf("MQTT healthcheck limit exceeded, exiting. %s ", err))
			exitCode = 1
			return
		}
	}

	serverConfig := server.Config{
		Host: cfg.HTTPTargetHost,
		Port: cfg.HTTPTargetPort,
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
	tracer := tp.Tracer(svcName)

	bsub, err := brokers.NewPubSub(ctx, cfg.BrokerURL, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer bsub.Close()
	bsub = brokerstracing.NewPubSub(serverConfig, tracer, bsub)

	mpub, err := mqttpub.NewPublisher(fmt.Sprintf("mqtt://%s:%s", cfg.MQTTTargetHost, cfg.MQTTTargetPort), cfg.MQTTQoS, cfg.MQTTForwarderTimeout)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create MQTT publisher: %s", err))
		exitCode = 1
		return
	}
	defer mpub.Close()

	fwd := mqtt.NewForwarder(brokers.SubjectAllChannels, logger)
	fwd = mqtttracing.New(serverConfig, tracer, fwd, brokers.SubjectAllChannels)
	if err := fwd.Forward(ctx, svcName, bsub, mpub); err != nil {
		logger.Error(fmt.Sprintf("failed to forward message broker messages: %s", err))
		exitCode = 1
		return
	}

	np, err := brokers.NewPublisher(ctx, cfg.BrokerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect to message broker: %s", err))
		exitCode = 1
		return
	}
	defer np.Close()
	np = brokerstracing.NewPublisher(serverConfig, tracer, np)

	es, err := events.NewEventStore(ctx, cfg.ESURL, cfg.Instance)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create %s event store : %s", svcName, err))
		exitCode = 1
		return
	}

	authConfig := auth.Config{}
	if err := env.ParseWithOptions(&authConfig, env.Options{Prefix: envPrefixAuthz}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s auth configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	authClient, authHandler, err := auth.SetupAuthz(ctx, authConfig)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authHandler.Close()

	logger.Info("Successfully connected to things grpc server " + authHandler.Secure())

	h := mqtt.NewHandler(np, es, logger, authClient)
	h = handler.NewTracing(tracer, h)

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, magistrala.Version, logger, cancel)
		go chc.CallHome(ctx)
	}
	var interceptor session.Interceptor

	// mProxy server Configuration for MQTT without TLS
	mqttConfig, err := mproxy.NewConfig(env.Options{Prefix: envPrefixMQTTNoTLS})
	if err != nil {
		panic(err)
	}
	// mProxy server for MQTT without TLS
	mqttProxy := mp.New(mqttConfig, h, interceptor, logger)
	logger.Info(fmt.Sprintf("Starting MQTT without TLS on port %s", mqttConfig.Address))
	g.Go(func() error {
		return mqttProxy.Listen(ctx)
	})


	// mProxy server Configuration for MQTT with TLS
	mqttTLSConfig, err := mproxy.NewConfig(env.Options{Prefix: envPrefixMQTTTLS})
	if err != nil {
		panic(err)
	}

	// mProxy server for MQTT with TLS
	mqttTLSProxy := mp.New(mqttTLSConfig, h, interceptor, logger)
	logger.Info(fmt.Sprintf("Starting MQTT with TLS on port %s", mqttTLSConfig.Address))
	g.Go(func() error {
		return mqttTLSProxy.Listen(ctx)
	})

	// mProxy server Configuration for MQTT over Websocket without TLS
	wsConfig, err := mproxy.NewConfig(env.Options{Prefix: envPrefixWSNoTLS})
	if err != nil {
		panic(err)
	}

	// mProxy server for MQTT over Websocket without TLS
	wsProxy := websocket.New(wsConfig, h, interceptor, logger)
	logger.Info(fmt.Sprintf("Starting MQTT over WS without TLS on port %s", wsConfig.Address))
	g.Go(func() error {
		return wsProxy.Listen(ctx)
	})

	// //  mProxy server Configuration for MQTT with mTLS
	// mqttMTLSConfig, err := mproxy.NewConfig(env.Options{Prefix: mqttWithmTLS})
	// if err != nil {
	// 	panic(err)
	// }

	// // mProxy server for MQTT with mTLS
	// mqttMTlsProxy := mp.New(mqttMTLSConfig, h, interceptor, logger)
	// g.Go(func() error {
	// 	return mqttMTlsProxy.Listen(ctx)
	// })

	// // mProxy server Configuration for MQTT over Websocket with TLS
	// wsTLSConfig, err := mproxy.NewConfig(env.Options{Prefix: envPrefixWSTLS})
	// if err != nil {
	// 	panic(err)
	// }

	// // mProxy server for MQTT over Websocket with TLS
	// wsTLSProxy := websocket.New(wsTLSConfig, h, interceptor, logger)
	// logger.Info(fmt.Sprintf("Starting MQTT over WS with TLS on port %s", wsTLSConfig.Address))
	// g.Go(func() error {
	// 	return wsTLSProxy.Listen(ctx)
	// })

	// // mProxy server Configuration for MQTT over Websocket with mTLS
	// wsMTLSConfig, err := mproxy.NewConfig(env.Options{Prefix: mqttWSWithmTLS})
	// if err != nil {
	// 	panic(err)
	// }

	// // mProxy server for MQTT over Websocket with mTLS
	// wsMTLSProxy := websocket.New(wsMTLSConfig, h, interceptor, logger)
	// g.Go(func() error {
	// 	return wsMTLSProxy.Listen(ctx)
	// })

	g.Go(func() error {
		return stopSignalHandler(ctx, cancel, logger)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("mProxy terminated: %s", err))
	}
}

func proxyMQTT(ctx context.Context, cfg config, logger *slog.Logger, sessionHandler session.Handler, interceptor session.Interceptor) error {
	config := mproxy.Config{
		Address: fmt.Sprintf(":%s", cfg.MQTTPort),
		Target:  fmt.Sprintf("%s:%s", cfg.MQTTTargetHost, cfg.MQTTTargetPort),
	}
	mproxy := mproxymqtt.New(config, sessionHandler, interceptor, logger)

	// errCh := make(chan error)
	// go func() {
	// 	errCh <- mproxy.Listen(ctx)
	// }()

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy MQTT shutdown at %s", config.Target))
		return nil
	case err := <-errCh:
		return err
	}
}

func proxyWS(ctx context.Context, cfg config, logger *slog.Logger, sessionHandler session.Handler, interceptor session.Interceptor) error {
	config := mproxy.Config{
		Address:    fmt.Sprintf("%s:%s", "", cfg.HTTPPort),
		Target:     fmt.Sprintf("%s:%s", cfg.HTTPTargetHost, cfg.HTTPTargetPort),
		PathPrefix: "/mqtt",
	}

	wp := websocket.New(config, sessionHandler, interceptor, logger)
	http.HandleFunc("/mqtt", wp.ServeHTTP)

// 	errCh := make(chan error)

	go func() {
		errCh <- wp.Listen(ctx)
	}()

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy MQTT WS shutdown at %s", config.Target))
		return nil
	case err := <-errCh:
		return err
	}
}

func healthcheck(cfg config) func() error {
	return func() error {
		res, err := http.Get(cfg.MQTTTargetHealthCheck)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return errors.New(string(body))
		}
		return nil
	}
}

func stopSignalHandler(ctx context.Context, cancel context.CancelFunc, logger *slog.Logger) error {
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	select {
	case sig := <-c:
		defer cancel()
		logger.Info(fmt.Sprintf("%s service shutdown by signal: %s", svcName, sig))
		return nil
	case <-ctx.Done():
		return nil
	}
}

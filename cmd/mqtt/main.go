package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cenkalti/backoff/v4"
	thingsClient "github.com/mainflux/mainflux/internal/client/grpc/things"
	redisClient "github.com/mainflux/mainflux/internal/client/redis"
	"github.com/mainflux/mainflux/internal/env"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt"
	mqttredis "github.com/mainflux/mainflux/mqtt/redis"
	"github.com/mainflux/mainflux/pkg/auth"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	mqttpub "github.com/mainflux/mainflux/pkg/messaging/mqtt"
	mp "github.com/mainflux/mproxy/pkg/mqtt"
	"github.com/mainflux/mproxy/pkg/session"
	ws "github.com/mainflux/mproxy/pkg/websocket"
	"golang.org/x/sync/errgroup"
)

const (
	svcName            = "mqtt"
	envPrefix          = "MF_MQTT_ADAPTER_"
	envPrefixHttp      = "MF_MQTT_ADAPTER_HTTP_"
	envPrefixES        = "MF_MQTT_ADAPTER_ES_"
	envPrefixAuthCache = "MF_AUTH_CACHE_"
)

type config struct {
	logLevel              string        `env:"MF_MQTT_ADAPTER_LOG_LEVEL"                    envDefault:"debug"`
	mqttPort              string        `env:"MF_MQTT_ADAPTER_MQTT_PORT"                    envDefault:"8080"`
	mqttTargetHost        string        `env:"MF_MQTT_ADAPTER_MQTT_TARGET_HOST"             envDefault:"localhost"`
	mqttTargetPort        string        `env:"MF_MQTT_ADAPTER_MQTT_TARGET_PORT"             envDefault:"8080"`
	mqttForwarderTimeout  time.Duration `env:"MF_MQTT_ADAPTER_FORWARDER_TIMEOUT"            envDefault:"30s"`
	mqttTargetHealthCheck string        `env:"MF_MQTT_ADAPTER_MQTT_TARGET_HEALTH_CHECK"     envDefault:""`
	httpPort              string        `env:"MF_MQTT_ADAPTER_WS_PORT"                      envDefault:"8080"`
	httpTargetHost        string        `env:"MF_MQTT_ADAPTER_WS_TARGET_HOST"               envDefault:"localhost"`
	httpTargetPort        string        `env:"MF_MQTT_ADAPTER_WS_TARGET_PORT"               envDefault:"8080"`
	httpTargetPath        string        `env:"MF_MQTT_ADAPTER_WS_TARGET_PATH"               envDefault:"/mqtt"`
	instance              string        `env:"MF_MQTT_ADAPTER_INSTANCE"                     envDefault:""`
	jaegerURL             string        `env:"MF_JAEGER_URL"                                envDefault:""`
	brokerURL             string        `env:"MF_BROKER_URL"                                envDefault:"nats://localhost:4222"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if cfg.mqttTargetHealthCheck != "" {
		notify := func(e error, next time.Duration) {
			logger.Info(fmt.Sprintf("Broker not ready: %s, next try in %s", e.Error(), next))
		}

		err := backoff.RetryNotify(healthcheck(cfg), backoff.NewExponentialBackOff(), notify)
		if err != nil {
			log.Fatalf("MQTT healthcheck limit exceeded, exiting. %s ", err.Error())
		}
	}

	nps, err := brokers.NewPubSub(cfg.brokerURL, "mqtt", logger)
	if err != nil {
		log.Fatalf("Failed to connect to message broker: %s", err.Error())
	}
	defer nps.Close()

	mpub, err := mqttpub.NewPublisher(fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort), cfg.mqttForwarderTimeout)
	if err != nil {
		log.Fatalf("Failed to create MQTT publisher: %s", err.Error())
	}

	fwd := mqtt.NewForwarder(brokers.SubjectAllChannels, logger)
	if err := fwd.Forward(svcName, nps, mpub); err != nil {
		log.Fatalf("Failed to forward message broker messages: %s", err)
	}

	np, err := brokers.NewPublisher(cfg.brokerURL)
	if err != nil {
		log.Fatalf("Failed to connect to message broker: %s", err.Error())
	}
	defer np.Close()

	ec, err := redisClient.Setup(envPrefixES)
	if err != nil {
		log.Fatalf("Failed to setup %s event store redis client : %s", svcName, err.Error())
	}
	defer ec.Close()

	es := mqttredis.NewEventStore(ec, cfg.instance)

	ac, err := redisClient.Setup(envPrefixAuthCache)
	if err != nil {
		log.Fatalf("Failed to setup %s event store redis client : %s", svcName, err.Error())
	}
	defer ac.Close()

	tc, thingsGrpcClient, thingsTracerCloser, thingsGrpcSecure, err := thingsClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer thingsGrpcClient.Close()
	defer thingsTracerCloser.Close()
	logger.Info("Successfully connected to things grpc server " + thingsGrpcSecure)

	authClient := auth.New(ac, tc)

	h := mqtt.NewHandler([]messaging.Publisher{np}, es, logger, authClient)

	logger.Info(fmt.Sprintf("Starting MQTT proxy on port %s", cfg.mqttPort))
	g.Go(func() error {
		return proxyMQTT(ctx, cfg, logger, h)
	})

	logger.Info(fmt.Sprintf("Starting MQTT over WS  proxy on port %s", cfg.httpPort))
	g.Go(func() error {
		return proxyWS(ctx, cfg, logger, h)
	})

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

func proxyMQTT(ctx context.Context, cfg config, logger mflog.Logger, handler session.Handler) error {
	address := fmt.Sprintf(":%s", cfg.mqttPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mp.New(address, target, handler, logger)

	errCh := make(chan error)
	go func() {
		errCh <- mp.Listen()
	}()

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy MQTT shutdown at %s", target))
		return nil
	case err := <-errCh:
		return err
	}
}

func proxyWS(ctx context.Context, cfg config, logger mflog.Logger, handler session.Handler) error {
	target := fmt.Sprintf("%s:%s", cfg.httpTargetHost, cfg.httpTargetPort)
	wp := ws.New(target, cfg.httpTargetPath, "ws", handler, logger)
	http.Handle("/mqtt", wp.Handler())

	errCh := make(chan error)

	go func() {
		errCh <- wp.Listen(cfg.httpPort)
	}()

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("proxy MQTT WS shutdown at %s", target))
		return nil
	case err := <-errCh:
		return err
	}
}

func healthcheck(cfg config) func() error {
	return func() error {
		res, err := http.Get(cfg.mqttTargetHealthCheck)
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

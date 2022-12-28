// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	mqttPaho "github.com/eclipse/paho.mqtt.golang"
	r "github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/lora"
	"github.com/mainflux/mainflux/lora/api"
	"github.com/mainflux/mainflux/lora/mqtt"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"golang.org/x/sync/errgroup"

	redisClient "github.com/mainflux/mainflux/internal/client/redis"
	"github.com/mainflux/mainflux/lora/redis"
)

const (
	svcName           = "lora-adapter"
	envPrefix         = "MF_LORA_ADAPTER_"
	envPrefixHttp     = "MF_LORA_ADAPTER_HTTP_"
	envPrefixRouteMap = "MF_LORA_ADAPTER_ROUTE_MAP_"
	envPrefixThingsES = "MF_THINGS_ES_"

	thingsRMPrefix   = "thing"
	channelsRMPrefix = "channel"
	connsRMPrefix    = "connection"
)

type config struct {
	logLevel       string        `env:"MF_BOOTSTRAP_LOG_LEVEL"              envDefault:"debug"`
	loraMsgURL     string        `env:"MF_LORA_ADAPTER_MESSAGES_URL"        envDefault:"debug"`
	loraMsgUser    string        `env:"MF_LORA_ADAPTER_MESSAGES_USER"       envDefault:"debug"`
	loraMsgPass    string        `env:"MF_LORA_ADAPTER_MESSAGES_PASS"       envDefault:"debug"`
	loraMsgTopic   string        `env:"MF_LORA_ADAPTER_MESSAGES_TOPIC"      envDefault:"debug"`
	loraMsgTimeout time.Duration `env:"MF_LORA_ADAPTER_MESSAGES_TIMEOUT"    envDefault:"debug"`
	esConsumerName string        `env:"MF_LORA_ADAPTER_EVENT_CONSUMER"      envDefault:"debug"`
	brokerURL      string        `env:"MF_BROKER_URL"                       envDefault:"debug"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	rmConn, err := redisClient.Setup(envPrefixRouteMap)
	if err != nil {
		log.Fatalf("Failed to setup route map redis client : %s", err.Error())
	}
	defer rmConn.Close()

	pub, err := brokers.NewPublisher(cfg.brokerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to message broker: %s", err))
		os.Exit(1)
	}
	defer pub.Close()

	svc := newService(pub, rmConn, thingsRMPrefix, channelsRMPrefix, connsRMPrefix, logger)

	esConn, err := redisClient.Setup(envPrefixThingsES)
	if err != nil {
		log.Fatalf("Failed to setup things event store redis client : %s", err.Error())
	}
	defer esConn.Close()

	mqttConn := connectToMQTTBroker(cfg.loraMsgURL, cfg.loraMsgUser, cfg.loraMsgPass, cfg.loraMsgTimeout, logger)

	go subscribeToLoRaBroker(svc, mqttConn, cfg.loraMsgTimeout, cfg.loraMsgTopic, logger)
	go subscribeToThingsES(svc, esConn, cfg.esConsumerName, logger)

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("Failed to load %s HTTP server configuration : %s", svcName, err.Error())
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("LoRa adapter terminated: %s", err))
	}
}

func connectToMQTTBroker(url, user, password string, timeout time.Duration, logger logger.Logger) mqttPaho.Client {
	opts := mqttPaho.NewClientOptions()
	opts.AddBroker(url)
	opts.SetUsername(user)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(func(c mqttPaho.Client) {
		logger.Info("Connected to Lora MQTT broker")
	})
	opts.SetConnectionLostHandler(func(c mqttPaho.Client, err error) {
		logger.Error(fmt.Sprintf("MQTT connection lost: %s", err.Error()))
		os.Exit(1)
	})

	client := mqttPaho.NewClient(opts)

	if token := client.Connect(); token.WaitTimeout(timeout) && token.Error() != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Lora MQTT broker: %s", token.Error()))
		os.Exit(1)
	}

	return client
}

func subscribeToLoRaBroker(svc lora.Service, mc mqttPaho.Client, timeout time.Duration, topic string, logger logger.Logger) {
	mqtt := mqtt.NewBroker(svc, mc, timeout, logger)
	logger.Info("Subscribed to Lora MQTT broker")
	if err := mqtt.Subscribe(topic); err != nil {
		logger.Error(fmt.Sprintf("Failed to subscribe to Lora MQTT broker: %s", err))
		os.Exit(1)
	}
}

func subscribeToThingsES(svc lora.Service, client *r.Client, consumer string, logger logger.Logger) {
	eventStore := redis.NewEventStore(svc, client, consumer, logger)
	logger.Info("Subscribed to Redis Event Store")
	if err := eventStore.Subscribe(context.Background(), "mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("Lora-adapter service failed to subscribe to Redis event source: %s", err))
	}
}

func newRouteMapRepository(client *r.Client, prefix string, logger logger.Logger) lora.RouteMapRepository {
	logger.Info(fmt.Sprintf("Connected to %s Redis Route-map", prefix))
	return redis.NewRouteMapRepository(client, prefix)
}

func newService(pub messaging.Publisher, rmConn *r.Client, thingsRMPrefix, channelsRMPrefix, connsRMPrefix string, logger logger.Logger) lora.Service {
	thingsRM := newRouteMapRepository(rmConn, thingsRMPrefix, logger)
	chansRM := newRouteMapRepository(rmConn, channelsRMPrefix, logger)
	connsRM := newRouteMapRepository(rmConn, connsRMPrefix, logger)

	svc := lora.New(pub, thingsRM, chansRM, connsRM)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}

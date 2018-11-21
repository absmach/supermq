package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/lora"
	"github.com/mainflux/mainflux/lora/api"
	natsBroker "github.com/mainflux/mainflux/lora/nats"
	pahoBroker "github.com/mainflux/mainflux/lora/paho"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	apilora "github.com/brocaar/lora-app-server/api"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	router "github.com/mainflux/mainflux/lora/redis"
	"github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	defLoraMsgURL   = "tcp://localhost:1883"
	defLoraServURL  = "localhost:8080"
	defNatsURL      = nats.DefaultURL
	defLogLevel     = "error"
	defRouteMapURL  = "localhost:6379"
	defRouteMapPass = ""
	defRouteMapDB   = "0"
	envLoraMsgURL   = "MF_LORA_ADAPTER_LORA_MESSAGE_URL"
	envLoraServURL  = "MF_LORA_ADAPTER_LORA_SERVER_URL"
	envNatsURL      = "MF_NATS_URL"
	envLogLevel     = "MF_LORA_ADAPTER_LOG_LEVEL"
	envRouteMapURL  = "MF_LORA_ADAPTER_CACHE_URL"
	envRouteMapPass = "MF_LORA_ADAPTER_CACHE_PASS"
	envRouteMapDB   = "MF_LORA_ADAPTER_CACHE_DB"
)

type config struct {
	LoraMsgURL   string
	LoraServURL  string
	NatsURL      string
	LogLevel     string
	RouteMapURL  string
	RouteMapPass string
	RouteMapDB   string
}

type jwt struct {
	token string
}

func (t jwt) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (jwt) RequireTransportSecurity() bool {
	return true
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	natsConn := connectToNATS(cfg.NatsURL, logger)
	mqttConn := connectToMQTTBroker(cfg.LoraMsgURL, logger)
	grpcConn := connectToGRPC(cfg.LoraServURL, logger)
	defer grpcConn.Close()

	redisConn := connectToRouteMap(cfg.RouteMapURL, cfg.RouteMapPass, cfg.RouteMapDB, logger)

	asConn := apilora.NewApplicationServiceClient(grpcConn)
	routeMap := router.NewRouteMapRepository(redisConn)

	svc := lora.New(natsConn, asConn, routeMap, logger)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "lora_adapter",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "lora_adapter",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	go connectToMfxBroker(svc, natsConn, logger)
	go connectToLoraBroker(svc, mqttConn, natsConn, logger)

	errs := make(chan error, 1)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("LoRa adapter terminated: %s", err))
}

func loadConfig() config {
	return config{
		LoraMsgURL:   mainflux.Env(envLoraMsgURL, defLoraMsgURL),
		LoraServURL:  mainflux.Env(envLoraServURL, defLoraServURL),
		NatsURL:      mainflux.Env(envNatsURL, defNatsURL),
		LogLevel:     mainflux.Env(envLogLevel, defLogLevel),
		RouteMapURL:  mainflux.Env(envRouteMapURL, defRouteMapURL),
		RouteMapPass: mainflux.Env(envRouteMapPass, defRouteMapPass),
		RouteMapDB:   mainflux.Env(envRouteMapDB, defRouteMapDB),
	}
}

func connectToNATS(url string, logger logger.Logger) *nats.Conn {
	conn, err := nats.Connect(url)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer conn.Close()

	return conn
}

func connectToMQTTBroker(loraURL string, logger logger.Logger) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(loraURL)
	opts.SetUsername("")
	opts.SetPassword("")
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		logger.Info("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		logger.Error(fmt.Sprintf("MQTT connection lost: %s", err.Error()))
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Lora MQTT broker: %s", token.Error()))
		os.Exit(1)
	}

	return client
}

func connectToRouteMap(mappingURL, mappingPass string, mappingDB string, logger logger.Logger) *redis.Client {
	db, err := strconv.Atoi(mappingDB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to cache: %s", err))
		os.Exit(1)
	}

	return redis.NewClient(&redis.Options{
		Addr:     mappingURL,
		Password: mappingPass,
		DB:       db,
	})
}

func connectToGRPC(url string, logger logger.Logger) *grpc.ClientConn {
	conn, err := grpc.Dial(url,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")),
		grpc.WithPerRPCCredentials(jwt{
			token: "verysecret",
		}),
	)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to GRPC: %s", err))
		os.Exit(1)
	}

	return conn
}

func connectToMfxBroker(svc lora.Service, nc *nats.Conn, logger logger.Logger) {
	logger.Info(fmt.Sprintf("Lora-adapter service subscribed to MFX NATS broker."))
	if err := natsBroker.Subscribe(svc, nc, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to MFX NATS broker: %s", err))
		os.Exit(1)
	}
}

func connectToLoraBroker(svc lora.Service, mc mqtt.Client, nc *nats.Conn, logger logger.Logger) {
	logger.Info(fmt.Sprintf("Lora-adapter service subscribed to Lora MQTT broker."))
	if err := pahoBroker.Subscribe(svc, mc, nc, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Lora MQTT broker: %s", err))
		os.Exit(1)
	}
}

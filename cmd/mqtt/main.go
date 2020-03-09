package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	mqtt "github.com/mainflux/mainflux/mqtt"
	"github.com/mainflux/mainflux/mqtt/nats"
	mr "github.com/mainflux/mainflux/mqtt/redis"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	mp "github.com/mainflux/mproxy/pkg/mqtt"
	broker "github.com/nats-io/nats.go"
	opentracing "github.com/opentracing/opentracing-go"
	jconfig "github.com/uber/jaeger-client-go/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defLogLevel          = "error"
	defNatsURL           = broker.DefaultURL
	defMQTTHost          = "0.0.0.0"
	defMQTTPort          = "1883"
	defMQTTTargetHost    = "0.0.0.0"
	defMQTTTargetPort    = "1884"
	defClientTLS         = "false"
	defCACerts           = ""
	defInstance          = ""
	defESURL             = mainflux.DefRedisURL
	defESPass            = ""
	defESDB              = "0"
	defJaegerURL         = ""
	defThingsAuthURL     = mainflux.DefThingsAuthURL
	defThingsAuthTimeout = "1" // in seconds

	envLogLevel          = "MF_MQTT_ADAPTER_LOG_LEVEL"
	envNatsURL           = "MF_NATS_URL"
	envMQTTHost          = "MF_MQTT_ADAPTER_MQTT_HOST"
	envMQTTPort          = "MF_MQTT_ADAPTER_MQTT_PORT"
	envMQTTTargetHost    = "MF_MQTT_ADAPTER_MQTT_TARGET_HOST"
	envMQTTTargetPort    = "MF_MQTT_ADAPTER_MQTT_TARGET_PORT"
	envClientTLS         = "MF_MQTT_ADAPTER_CLIENT_TLS"
	envCACerts           = "MF_MQTT_ADAPTER_CA_CERTS"
	envInstance          = "MF_MQTT_ADAPTER_INSTANCE"
	envESURL             = "MF_MQTT_ADAPTER_ES_URL"
	envESPass            = "MF_MQTT_ADAPTER_ES_PASS"
	envESDB              = "MF_MQTT_ADAPTER_ES_DB"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	mqttHost          string
	mqttPort          string
	mqttTargetHost    string
	mqttTargetPort    string
	jaegerURL         string
	logLevel          string
	natsURL           string
	clientTLS         bool
	caCerts           string
	instance          string
	esURL             string
	esPass            string
	esDB              string
	thingsAuthURL     string
	thingsAuthTimeout time.Duration
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	nc, err := broker.Connect(cfg.natsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nc.Close()

	conn := connectToThings(cfg, logger)
	defer conn.Close()

	tracer, closer := initJaeger("mproxy", cfg.jaegerURL, logger)
	defer closer.Close()

	thingsTracer, thingsCloser := initJaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	cc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)
	pub := nats.NewMessagePublisher(nc)

	rc := connectToRedis(cfg.esURL, cfg.esPass, cfg.esDB, logger)
	defer rc.Close()

	es := mr.NewEventStore(rc, cfg.instance)
	pubs := []mainflux.MessagePublisher{pub}

	// Event handler for MQTT hooks
	evt := mqtt.New(cc, pubs, es, logger, tracer)

	errs := make(chan error, 2)

	logger.Info(fmt.Sprintf("Starting MQTT proxy on port %s ", cfg.mqttPort))
	go proxyMQTT(cfg, logger, evt, errs)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("mProxy terminated: %s", err))
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	timeout, err := strconv.ParseInt(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		mqttHost:       mainflux.Env(envMQTTHost, defMQTTHost),
		mqttPort:       mainflux.Env(envMQTTPort, defMQTTPort),
		mqttTargetHost: mainflux.Env(envMQTTTargetHost, defMQTTTargetHost),
		mqttTargetPort: mainflux.Env(envMQTTTargetPort, defMQTTTargetPort),

		natsURL:   mainflux.Env(envNatsURL, defNatsURL),
		logLevel:  mainflux.Env(envLogLevel, defLogLevel),
		clientTLS: tls,
		caCerts:   mainflux.Env(envCACerts, defCACerts),
		instance:  mainflux.Env(envInstance, defInstance),
		esURL:     mainflux.Env(envESURL, defESURL),
		esPass:    mainflux.Env(envESPass, defESPass),
		esDB:      mainflux.Env(envESDB, defESDB),
		jaegerURL: mainflux.Env(envJaegerURL, defJaegerURL),

		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: time.Duration(timeout) * time.Second,
	}
}

func initJaeger(svcName, url string, logger logger.Logger) (opentracing.Tracer, io.Closer) {
	if url == "" {
		return opentracing.NoopTracer{}, ioutil.NopCloser(nil)
	}

	tracer, closer, err := jconfig.Configuration{
		ServiceName: svcName,
		Sampler: &jconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jconfig.ReporterConfig{
			LocalAgentHostPort: url,
			LogSpans:           true,
		},
	}.NewTracer()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger client: %s", err))
		os.Exit(1)
	}

	return tracer, closer
}

func connectToThings(cfg config, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption
	if cfg.clientTLS {
		if cfg.caCerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(cfg.caCerts, "")
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to load certs: %s", err))
				os.Exit(1)
			}
			opts = append(opts, grpc.WithTransportCredentials(tpc))
		}
	} else {
		logger.Info("gRPC communication is not encrypted")
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(cfg.thingsAuthURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to things service: %s", err))
		os.Exit(1)
	}
	return conn
}

func connectToRedis(redisURL, redisPass, redisDB string, logger logger.Logger) *redis.Client {
	db, err := strconv.Atoi(redisDB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to redis: %s", err))
		os.Exit(1)
	}

	return redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPass,
		DB:       db,
	})
}

func proxyMQTT(cfg config, logger logger.Logger, evt *mqtt.Event, errs chan error) {
	mp := mp.New(cfg.mqttHost, cfg.mqttPort, cfg.mqttTargetHost, cfg.mqttTargetPort, evt, logger)

	errs <- mp.Proxy()
}

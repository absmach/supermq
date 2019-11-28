//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mainflux/mainflux/twins/paho"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/api"
	"github.com/mainflux/mainflux/twins/nats"
	localusers "github.com/mainflux/mainflux/twins/users"
	"github.com/mainflux/mainflux/twins/uuid"

	twinshttpapi "github.com/mainflux/mainflux/twins/api/twins/http"
	twinsmongodb "github.com/mainflux/mainflux/twins/mongodb"
	usersapi "github.com/mainflux/mainflux/users/api/grpc"
	"go.mongodb.org/mongo-driver/mongo"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	broker "github.com/nats-io/go-nats"
	opentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	jconfig "github.com/uber/jaeger-client-go/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defLogLevel        = "info"
	defHTTPPort        = "9021"
	defJaegerURL       = ""
	defServerCert      = ""
	defServerKey       = ""
	defDBName          = "mainflux"
	defDBHost          = "localhost"
	defDBPort          = "27017"
	defSingleUserEmail = ""
	defSingleUserToken = ""
	defClientTLS       = "false"
	defCACerts         = ""
	defUsersURL        = "localhost:8181"
	defUsersTimeout    = "1" // in seconds
	defMqttURL         = "tcp://localhost:1883"
	defThingID         = ""
	defThingKey        = ""
	defChannelID       = ""
	defNatsURL         = broker.DefaultURL

	envLogLevel        = "MF_TWINS_LOG_LEVEL"
	envHTTPPort        = "MF_TWINS_HTTP_PORT"
	envJaegerURL       = "MF_JAEGER_URL"
	envServerCert      = "MF_TWINS_SERVER_CERT"
	envServerKey       = "MF_TWINS_SERVER_KEY"
	envDBName          = "MF_MONGODB_NAME"
	envDBHost          = "MF_MONGODB_HOST"
	envDBPort          = "MF_MONGODB_PORT"
	envSingleUserEmail = "MF_TWINS_SINGLE_USER_EMAIL"
	envSingleUserToken = "MF_TWINS_SINGLE_USER_TOKEN"
	envClientTLS       = "MF_TWINS_CLIENT_TLS"
	envCACerts         = "MF_TWINS_CA_CERTS"
	envUsersURL        = "MF_USERS_URL"
	envUsersTimeout    = "MF_TWINS_USERS_TIMEOUT"
	envMqttURL         = "MF_TWINS_MQTT_URL"
	envThingID         = "MF_TWINS_THING_ID"
	envThingKey        = "MF_TWINS_THING_KEY"
	envChannelID       = "MF_TWINS_CHANNEL_ID"
	envNatsURL         = "MF_NATS_URL"
)

type config struct {
	logLevel        string
	httpPort        string
	jaegerURL       string
	serverCert      string
	serverKey       string
	dbCfg           twinsmongodb.Config
	singleUserEmail string
	singleUserToken string
	clientTLS       bool
	caCerts         string
	usersURL        string
	usersTimeout    time.Duration
	mqttURL         string
	thingID         string
	thingKey        string
	channelID       string
	NatsURL         string
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db, err := twinsmongodb.Connect(cfg.dbCfg, logger)
	if err != nil {
		log.Fatalf(err.Error())
	}

	usersTracer, usersCloser := initJaeger("users", cfg.jaegerURL, logger)
	defer usersCloser.Close()

	users, close := createUsersClient(cfg, usersTracer, logger)
	if close != nil {
		defer close()
	}

	dbTracer, dbCloser := initJaeger("twins_db", cfg.jaegerURL, logger)
	defer dbCloser.Close()

	pc := paho.Connect(cfg.mqttURL, cfg.thingID, cfg.thingKey, logger)
	mc := paho.New(pc, cfg.channelID)

	mcTracer, mcCloser := initJaeger("twins_mqtt", cfg.jaegerURL, logger)
	defer mcCloser.Close()

	nc, err := broker.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nc.Close()

	ncTracer, ncCloser := initJaeger("twins_nats", cfg.jaegerURL, logger)
	defer ncCloser.Close()

	tracer, closer := initJaeger("twins", cfg.jaegerURL, logger)
	defer closer.Close()

	svc := newService(nc, ncTracer, mc, mcTracer, users, dbTracer, db, logger)
	errs := make(chan error, 2)

	go startHTTPServer(twinshttpapi.MakeHandler(tracer, svc), cfg.httpPort, cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Twins service terminated: %s", err))
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	timeout, err := strconv.ParseInt(mainflux.Env(envUsersTimeout, defUsersTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envUsersTimeout, err.Error())
	}

	dbCfg := twinsmongodb.Config{
		Name: mainflux.Env(envDBName, defDBName),
		Host: mainflux.Env(envDBHost, defDBHost),
		Port: mainflux.Env(envDBPort, defDBPort),
	}

	return config{
		logLevel:        mainflux.Env(envLogLevel, defLogLevel),
		httpPort:        mainflux.Env(envHTTPPort, defHTTPPort),
		serverCert:      mainflux.Env(envServerCert, defServerCert),
		serverKey:       mainflux.Env(envServerKey, defServerKey),
		jaegerURL:       mainflux.Env(envJaegerURL, defJaegerURL),
		dbCfg:           dbCfg,
		singleUserEmail: mainflux.Env(envSingleUserEmail, defSingleUserEmail),
		singleUserToken: mainflux.Env(envSingleUserToken, defSingleUserToken),
		clientTLS:       tls,
		caCerts:         mainflux.Env(envCACerts, defCACerts),
		usersURL:        mainflux.Env(envUsersURL, defUsersURL),
		usersTimeout:    time.Duration(timeout) * time.Second,
		mqttURL:         mainflux.Env(envMqttURL, defMqttURL),
		thingID:         mainflux.Env(envThingID, defThingID),
		channelID:       mainflux.Env(envChannelID, defChannelID),
		thingKey:        mainflux.Env(envThingKey, defThingKey),
		NatsURL:         mainflux.Env(envNatsURL, defNatsURL),
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

func createUsersClient(cfg config, tracer opentracing.Tracer, logger logger.Logger) (mainflux.UsersServiceClient, func() error) {
	if cfg.singleUserEmail != "" && cfg.singleUserToken != "" {
		return localusers.NewSingleUserService(cfg.singleUserEmail, cfg.singleUserToken), nil
	}

	conn := connectToUsers(cfg, logger)
	return usersapi.NewClient(tracer, conn, cfg.usersTimeout), conn.Close
}

func connectToUsers(cfg config, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption
	if cfg.clientTLS {
		if cfg.caCerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(cfg.caCerts, "")
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create tls credentials: %s", err))
				os.Exit(1)
			}
			opts = append(opts, grpc.WithTransportCredentials(tpc))
		}
	} else {
		opts = append(opts, grpc.WithInsecure())
		logger.Info("gRPC communication is not encrypted")
	}

	conn, err := grpc.Dial(cfg.usersURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to users service: %s", err))
		os.Exit(1)
	}

	logger.Info("Connected to users")

	return conn
}

func newService(nc *broker.Conn, ncTracer opentracing.Tracer, mc paho.Mqtt, mcTracer opentracing.Tracer, users mainflux.UsersServiceClient, dbTracer opentracing.Tracer, db *mongo.Database, logger logger.Logger) twins.Service {
	twinRepo := twinsmongodb.NewTwinRepository(db)
	idp := uuid.New()

	// TODO twinRepo = tracing.TwinRepositoryMiddleware(dbTracer, thingsRepo)
	nats.Subscribe(nc, mc, twinRepo, logger)

	svc := twins.New(nc, mc, users, twinRepo, idp)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "twins",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "twins",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	return svc
}

func startHTTPServer(handler http.Handler, port string, cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	if cfg.serverCert != "" || cfg.serverKey != "" {
		logger.Info(fmt.Sprintf("Twins service started using https on port %s with cert %s key %s",
			port, cfg.serverCert, cfg.serverKey))
		errs <- http.ListenAndServeTLS(p, cfg.serverCert, cfg.serverKey, handler)
		return
	}
	logger.Info(fmt.Sprintf("Twins service started using http on port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, handler)
}

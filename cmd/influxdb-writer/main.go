// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	influxdata "github.com/influxdata/influxdb-client-go/v2"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/consumers/writers/influxdb"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/transformers"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	svcName = "influxdb-writer"

	defNatsURL          = "nats://190.190.190.81:4222"
	defLogLevel         = "error"
	defPort             = "8180"
	defDB               = "mainflux"
	defDBUrl            = "http://localhost:8086"
	defConfigPath       = "/config.toml"
	defContentType      = "application/senml+json"
	defTransformer      = "senml"
	defDBToken          = ""
	defOrg              = "mainflux"
	defBucket           = "mainflux"
	defMainfluxApiToken = ""
	defMainfluxUrl      = "http://190.190.190.81"

	envNatsURL          = "MF_NATS_URL"
	envLogLevel         = "MF_INFLUX_WRITER_LOG_LEVEL"
	envPort             = "MF_INFLUX_WRITER_PORT"
	envDB               = "MF_INFLUXDB_DB"
	envDBUrl            = "MF_INFLUX_WRITER_DB_URL"
	envConfigPath       = "MF_INFLUX_WRITER_CONFIG_PATH"
	envContentType      = "MF_INFLUX_WRITER_CONTENT_TYPE"
	envTransformer      = "MF_INFLUX_WRITER_TRANSFORMER"
	envDBToken          = "MF_INFLUXDB_ADMIN_TOKEN"
	envOrg              = "MF_INFLUXDB_ORG"
	envBucket           = "MF_INFLUXDB_BUCKET"
	envMainfluxApiToken = "MF_API_KEY"
	envMainfluxUrl      = "MF_URL"
)

type config struct {
	natsURL          string
	logLevel         string
	port             string
	dbName           string
	dbUrl            string
	dbToken          string
	configPath       string
	contentType      string
	transformer      string
	org              string
	bucket           string
	mainfluxApiToken string
	mainfluxUrl      string
}

func main() {
	cfg, dbAddr, dbToken := loadConfigs()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	client := influxdata.NewClientWithOptions(dbAddr, dbToken, influxdata.DefaultOptions().SetBatchSize(5000))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		os.Exit(1)
	}
	defer client.Close()

	repo := influxdb.New(client, cfg.org, cfg.bucket, cfg.mainfluxApiToken, cfg.mainfluxUrl)

	counter, latency := makeMetrics()
	repo = api.LoggingMiddleware(repo, logger)
	repo = api.MetricsMiddleware(repo, counter, latency)
	t := makeTransformer(cfg, logger)

	if err := consumers.Start(pubSub, repo, t, cfg.configPath, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to start InfluxDB writer: %s", err))
		os.Exit(1)
	}

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPService(cfg.port, logger, errs)

	err = <-errs
	logger.Error(fmt.Sprintf("InfluxDB writer service terminated: %s", err))
}

func loadConfigs() (config, string, string) {
	cfg := config{
		natsURL:          mainflux.Env(envNatsURL, defNatsURL),
		logLevel:         mainflux.Env(envLogLevel, defLogLevel),
		port:             mainflux.Env(envPort, defPort),
		dbName:           mainflux.Env(envDB, defDB),
		dbUrl:            mainflux.Env(envDBUrl, defDBUrl),
		dbToken:          mainflux.Env(envDBToken, defDBToken),
		configPath:       mainflux.Env(envConfigPath, defConfigPath),
		contentType:      mainflux.Env(envContentType, defContentType),
		transformer:      mainflux.Env(envTransformer, defTransformer),
		org:              mainflux.Env(envOrg, defOrg),
		bucket:           mainflux.Env(envBucket, defBucket),
		mainfluxApiToken: mainflux.Env(envMainfluxApiToken, defMainfluxApiToken),
		mainfluxUrl:      mainflux.Env(envMainfluxUrl, defMainfluxUrl),
	}
	return cfg, cfg.dbUrl, cfg.dbToken
}

func makeMetrics() (*kitprometheus.Counter, *kitprometheus.Summary) {
	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_count",
		Help:      "Number of database inserts.",
	}, []string{"method"})

	latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of inserts in microseconds.",
	}, []string{"method"})

	return counter, latency
}

func makeTransformer(cfg config, logger logger.Logger) transformers.Transformer {
	switch strings.ToUpper(cfg.transformer) {
	case "SENML":
		logger.Info("Using SenML transformer")
		return senml.New(cfg.contentType)
	case "JSON":
		logger.Info("Using JSON transformer")
		return json.New()
	default:
		logger.Error(fmt.Sprintf("Can't create transformer: unknown transformer type %s", cfg.transformer))
		os.Exit(1)
		return nil
	}
}

func startHTTPService(port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("InfluxDB writer service started, exposed port %s", p))
	errs <- http.ListenAndServe(p, api.MakeHandler(svcName))
}

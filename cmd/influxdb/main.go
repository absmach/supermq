package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/writers"
	influxdb "github.com/mainflux/mainflux/writers/influxdb"
	nats "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	senML        = "out.senml"
	prefix       = "http://"
	defNatsURL   = nats.DefaultURL
	defPort      = "8180"
	defPointName = "messages"
	defDBName    = "mainflux"
	defDBHost    = "localhost"
	defDBPort    = "8086"
	defDBUser    = "mainflux"
	defDBPass    = "mainflux"

	envNatsURL   = "MF_NATS_URL"
	envPort      = "MF_INFLUXDB_WRITER_PORT"
	envPointName = "MF_INFLUXDB_POINT_NAME"
	envDBName    = "MF_INFLUXDB_DB_NAME"
	envDBHost    = "MF_INFLUXDB_DB_HOST"
	envDBPort    = "MF_INFLUXDB_DB_PORT"
	envDBUser    = "MF_INFLUXDB_DB_USER"
	envDBPass    = "MF_INFLUXDB_DB_PASS"
)

type config struct {
	NatsURL   string
	Port      string
	PointName string
	DBName    string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
}

func makeMetrices() (*kitprometheus.Counter, *kitprometheus.Summary) {
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

func makeConfig() config {
	cfg := config{
		NatsURL:   mainflux.Env(envNatsURL, defNatsURL),
		PointName: mainflux.Env(envPointName, defPointName),
		Port:      mainflux.Env(envPort, defPort),
		DBName:    mainflux.Env(envDBName, defDBName),
		DBHost:    mainflux.Env(envDBHost, defDBHost),
		DBPort:    mainflux.Env(envDBPort, defDBPort),
		DBUser:    mainflux.Env(envDBUser, defDBUser),
		DBPass:    mainflux.Env(envDBPass, defDBPass),
	}
	return cfg
}

func main() {
	cfg := makeConfig()
	logger := log.New(os.Stdout)

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nc.Close()

	errs := make(chan error, 2)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	clientCfg := influxdata.HTTPConfig{
		Addr:     fmt.Sprintf("%s%s:%s", prefix, cfg.DBHost, cfg.DBPort),
		Username: cfg.DBUser,
		Password: cfg.DBPass,
	}

	client, err := influxdata.NewHTTPClient(clientCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		return
	}
	defer client.Close()

	repo, err := influxdb.New(client, cfg.DBName, cfg.PointName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB writer: %s", err.Error()))
		return
	}

	counter, latency := makeMetrices()
	if err := writers.Start("influxdb-writer", nc, logger, repo, counter, latency); err != nil {
		logger.Error(fmt.Sprintf("Failed to start message writer: %s", err))
		return
	}

	go func() {
		p := fmt.Sprintf(":%s", cfg.Port)
		logger.Info(fmt.Sprintf("Influxdb writer service started, exposed port %s", cfg.Port))
		errs <- http.ListenAndServe(p, influxdb.MakeHandler())
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Influxdb writer service terminated: %s", err))
}

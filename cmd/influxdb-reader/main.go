package main

import (
	"context"
	"fmt"
	"log"
	"os"

	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux/internal"
	authClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
	thingsClient "github.com/mainflux/mainflux/internal/client/grpc/things"
	influxClient "github.com/mainflux/mainflux/internal/client/influxdb"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/influxdb"
	"golang.org/x/sync/errgroup"
)

const (
	svcName           = "influxdb-reader"
	envPrefix         = "MF_INFLUX_READER_"
	envPrefixHttp     = "MF_INFLUX_READER_HTTP_"
	envPrefixInfluxdb = "MF_INFLUXDB_"
)

type config struct {
	LogLevel  string `env:"MF_INFLUX_READER_LOG_LEVEL"  envDefault:"debug"`
	JaegerURL string `env:"MF_JAEGER_URL"               envDefault:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := logger.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tc, tcHandler, err := thingsClient.Setup(envPrefix, cfg.JaegerURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer tcHandler.Close()
	logger.Info("Successfully connected to things grpc server " + tcHandler.Secure())

	auth, authHandler, err := authClient.Setup(envPrefix, cfg.JaegerURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer authHandler.Close()
	logger.Info("Successfully connected to auth grpc server " + authHandler.Secure())

	influxdbConfig := influxClient.Config{}
	if err := env.Parse(&influxdbConfig, env.Options{Prefix: envPrefixInfluxdb}); err != nil {
		log.Fatalf("failed to load InfluxDB client configuration from environment variable : %s", err.Error())
	}
	client, err := influxClient.Connect(influxdbConfig)
	if err != nil {
		log.Fatalf("failed to connect to InfluxDB : %s", err.Error())
	}
	defer client.Close()

	repo := newService(client, influxdbConfig.DbName, logger)

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf("failed to load %s HTTP server configuration : %s", svcName, err.Error())
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(repo, tc, auth, svcName, logger), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("InfluxDB reader service terminated: %s", err))
	}
}

func newService(client influxdata.Client, dbName string, logger logger.Logger) readers.MessageRepository {
	repo := influxdb.New(client, dbName)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("influxdb", "message_reader")
	repo = api.MetricsMiddleware(repo, counter, latency)

	return repo
}

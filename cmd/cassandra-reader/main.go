// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
	"github.com/mainflux/mainflux/internal"
	cassandraClient "github.com/mainflux/mainflux/internal/client/cassandra"
	authClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
	thingsClient "github.com/mainflux/mainflux/internal/client/grpc/things"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/cassandra"
	"golang.org/x/sync/errgroup"
)

const (
	svcName                 = "cassandra-reader"
	envPrefix               = "MF_CASSANDRA_READER_"
	envPrefixHttp           = "MF_CASSANDRA_READER_HTTP_"
	envThingsAuthGrpcPrefix = "MF_THINGS_AUTH_GRPC_"
	envAuthGrpcPrefix       = "MF_AUTH_GRPC_"
	sep                     = ","
	defLogLevel             = "error"
	defPort                 = "8180"
)

type config struct {
	logLevel  string `env:"MF_CASSANDRA_READER_LOG_LEVEL"     envDefault:"debug" `
	jaegerURL string `env:"MF_JAEGER_URL"                     envDefault:"" `
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// create cassandra reader service configurations
	cfg := config{}
	// load cassandra reader service configurations from environment
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load %s service configuration : %s", svcName, err.Error())
	}

	// create new logger
	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	// create new thing grpc client
	tc, thingsGrpcClient, thingsTracerCloser, thingsGrpcSecure, err := thingsClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer thingsGrpcClient.Close()
	defer thingsTracerCloser.Close()
	logger.Info("Successfully connected to things grpc server " + thingsGrpcSecure)

	// create new auth grpc client
	auth, authGrpcClient, authTracerCloser, authGrpcSecure, err := authClient.Setup(envPrefix, cfg.jaegerURL)

	if err != nil {
		log.Fatal(err.Error())
	}
	defer authGrpcClient.Close()
	defer authTracerCloser.Close()
	logger.Info("Successfully connected to auth grpc server " + authGrpcSecure)

	////////// CASSANDRA READER REPO /////////////
	// create new cassandra client
	cassaSession, err := cassandraClient.Setup(envPrefix)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer cassaSession.Close()

	repo := newService(cassaSession, logger)

	///////////////// HTTP SERVER //////////////////////////
	// create new http server config
	httpServerConfig := server.Config{}
	// load http server config from environment variables
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s HTTP server configuration : %s", svcName, err.Error()))
	}
	// create new http server
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(repo, tc, auth, svcName, logger), logger)

	//Start servers
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Cassandra reader service terminated: %s", err))
	}
}

func newService(cassaSession *gocql.Session, logger logger.Logger) readers.MessageRepository {
	repo := cassandra.New(cassaSession)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("cassandra", "message_reader")
	repo = api.MetricsMiddleware(repo, counter, latency)
	return repo
}

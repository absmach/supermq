// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gocql/gocql"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/internal"
	cassandraClient "github.com/mainflux/mainflux/internal/client/cassandra"
	grpcClient "github.com/mainflux/mainflux/internal/client/grpc"
	jaegerClient "github.com/mainflux/mainflux/internal/client/jaeger"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/cassandra"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"golang.org/x/sync/errgroup"
)

const (
	svcName        = "cassandra-reader"
	envPrefix      = "MF_CASSANDRA_READER_"
	envPrefixHttp  = "MF_CASSANDRA_READER_HTTP_"
	envThingPrefix = "MF_THINGS_"
	envAuthPrefix  = "MF_AUTH_"
	sep            = ","
	defLogLevel    = "error"
	defPort        = "8180"
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

	///////////////// CASSANDRA CLIENT /////////////////////////
	// create new cassandra config
	cassandraConfig := cassandraClient.Config{}
	// load cassandra config from environment
	if err := env.Parse(&cassandraConfig, env.Options{Prefix: envPrefix}); err != nil {
		log.Fatalf("Failed to load Cassandra database configuration : %s", err.Error())
	}
	// create new to cassandra client
	cassaSession, err := cassandraClient.Connect(cassandraConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Cassandra database : %s", err.Error())
	}
	defer cassaSession.Close()

	///////////////// THING GRPC CLIENT /////////////////////////
	// create thing grpc config
	thingGrpcConfig := grpcClient.Config{}
	// load things grpc client config from environment
	if err := env.Parse(&thingGrpcConfig, env.Options{Prefix: envPrefix, AltPrefix: envThingPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to thing grpc client configuration : %s", err.Error()))
	}
	// connect to thing grpc server
	thingGrpcClient, secure, err := grpcClient.Connect(thingGrpcConfig)
	if err != nil {
		log.Fatalf("Failed to connect to things gRPC server : %s ", err.Error())
	}
	defer thingGrpcClient.Close()
	secureMsg := "without TLS"
	if secure {
		secureMsg = "with TLS"
	}
	logger.Info(fmt.Sprintf("Connected to things gRPC server %s", secureMsg))
	// initialize things tracer for things grpc client
	thingsTracer, thingsCloser, err := jaegerClient.NewTracer("things", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to initialize to jaeger : %s ", err.Error())
	}
	defer thingsCloser.Close()
	// create new thing grpc client
	tc := thingsapi.NewClient(thingGrpcClient, thingsTracer, thingGrpcConfig.Timeout)

	///////////////// AUTH GRPC CLIENT //////////////////////////
	// loading auth grpc config
	authGrpcConfig := grpcClient.Config{}
	if err := env.Parse(&authGrpcConfig, env.Options{Prefix: envPrefix, AltPrefix: envThingPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("Failed to auth grpc client configuration : %s", err.Error()))
	}

	// connect to auth grpc server
	authGrpcClient, secure, err := grpcClient.Connect(thingGrpcConfig)
	if err != nil {
		log.Fatalf("Failed to connect to things gRPC : %s ", err.Error())
	}
	defer authGrpcClient.Close()
	secureMsg = "without TLS"
	if secure {
		secureMsg = "with TLS"
	}
	logger.Info(fmt.Sprintf("Connected to auth gRPC %s", secureMsg))

	// initialize auth tracer for auth grpc client
	authTracer, authCloser, err := jaegerClient.NewTracer("auth", cfg.jaegerURL)
	if err != nil {
		log.Fatalf("Failed to initialize to jaeger : %s ", err.Error())
	}
	defer authCloser.Close()

	// create new auth grpc client
	auth := authapi.NewClient(authTracer, authGrpcClient, authGrpcConfig.Timeout)

	////////// CASSANDRA READER REPO /////////////
	repo := newService(cassaSession, logger)

	///////////////// HTTP SERVER //////////////////////////
	// create new http server config
	httpServerConfig := server.Config{}
	// load http server config from environment variables
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefix, AltPrefix: envPrefixHttp}); err != nil {
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

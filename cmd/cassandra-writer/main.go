// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocql/gocql"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/consumers/writers/cassandra"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"golang.org/x/sync/errgroup"
)

const (
	svcName = "cassandra-writer"
	sep     = ","

	defBrokerURL  = "nats://localhost:4222"
	defLogLevel   = "error"
	defPort       = "8180"
	defCluster    = "127.0.0.1"
	defKeyspace   = "mainflux"
	defDBUser     = "mainflux"
	defDBPass     = "mainflux"
	defDBPort     = "9042"
	defConfigPath = "/config.toml"

	envBrokerURL  = "MF_BROKER_URL"
	envLogLevel   = "MF_CASSANDRA_WRITER_LOG_LEVEL"
	envPort       = "MF_CASSANDRA_WRITER_PORT"
	envCluster    = "MF_CASSANDRA_WRITER_DB_CLUSTER"
	envKeyspace   = "MF_CASSANDRA_WRITER_DB_KEYSPACE"
	envDBUser     = "MF_CASSANDRA_WRITER_DB_USER"
	envDBPass     = "MF_CASSANDRA_WRITER_DB_PASS"
	envDBPort     = "MF_CASSANDRA_WRITER_DB_PORT"
	envConfigPath = "MF_CASSANDRA_WRITER_CONFIG_PATH"
)

type config struct {
	brokerURL  string
	logLevel   string
	port       string
	configPath string
	dbCfg      cassandra.DBConfig
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// create cassandra writer service configurations
	cfg := loadConfig()

	// create new logger
	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubSub, err := brokers.NewPubSub(cfg.brokerURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to message broker: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	session := connectToCassandra(cfg.dbCfg, logger)
	defer session.Close()

	repo := newService(session, logger)

	if err := consumers.Start(svcName, pubSub, repo, cfg.configPath, logger); err != nil {
		if err := consumers.Start(svcName, pubSub, repo, cfg.configPath, logger); err != nil {
			logger.Error(fmt.Sprintf("Failed to create Cassandra writer: %s", err))
		}

		hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(svcName), "", "", logger)
		g.Go(func() error {
			return hs.Start()
		})

		g.Go(func() error {
			return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
		})

		if err := g.Wait(); err != nil {
			logger.Error(fmt.Sprintf("Cassandra writer service terminated: %s", err))
		}
	}
	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Cassandra writer service terminated: %s", err))
	}
}

func loadConfig() config {
	dbPort, err := strconv.Atoi(mainflux.Env(envDBPort, defDBPort))
	if err != nil {
		log.Fatal(err)
	}

	dbCfg := cassandra.DBConfig{
		Hosts:    strings.Split(mainflux.Env(envCluster, defCluster), sep),
		Keyspace: mainflux.Env(envKeyspace, defKeyspace),
		User:     mainflux.Env(envDBUser, defDBUser),
		Pass:     mainflux.Env(envDBPass, defDBPass),
		Port:     dbPort,
	}

	return config{
		brokerURL:  mainflux.Env(envBrokerURL, defBrokerURL),
		logLevel:   mainflux.Env(envLogLevel, defLogLevel),
		port:       mainflux.Env(envPort, defPort),
		configPath: mainflux.Env(envConfigPath, defConfigPath),
		dbCfg:      dbCfg,
	}
}

func connectToCassandra(dbCfg cassandra.DBConfig, logger logger.Logger) *gocql.Session {
	session, err := cassandra.Connect(dbCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Cassandra cluster: %s", err))
		os.Exit(1)
	}

	return session
}

func newService(session *gocql.Session, logger logger.Logger) consumers.Consumer {
	repo := cassandra.New(session)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("cassandra", "message_writer")
	repo = api.MetricsMiddleware(repo, counter, latency)
	return repo
}

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
	"time"

	"github.com/gocql/gocql"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
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
	svcName = "cassandra-reader"

	sep         = ","
	defLogLevel = "error"
	defPort     = "8180"

	defClientTLS         = "false"
	defCACerts           = ""
	defServerCert        = ""
	defServerKey         = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"
	defUsersAuthURL      = "localhost:8181"
	defUsersAuthTimeout  = "1s"

	envLogLevel = "MF_CASSANDRA_READER_LOG_LEVEL"
	envPort     = "MF_CASSANDRA_READER_PORT"

	envClientTLS         = "MF_CASSANDRA_READER_CLIENT_TLS"
	envCACerts           = "MF_CASSANDRA_READER_CA_CERTS"
	envServerCert        = "MF_CASSANDRA_READER_SERVER_CERT"
	envServerKey         = "MF_CASSANDRA_READER_SERVER_KEY"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
	envUsersAuthURL      = "MF_AUTH_GRPC_URL"
	envUsersAuthTimeout  = "MF_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel          string        `env:"MF_CASSANDRA_READER_LOG_LEVEL"     default:"debug" `
	port              string        `env:"MF_CASSANDRA_READER_PORT"     default:"8180" `
	clientTLS         bool          `env:"DB_CLUSTER"     default:"false" `
	caCerts           string        `env:"DB_CLUSTER"     default:"" `
	serverCert        string        `env:"DB_CLUSTER"     default:"" `
	serverKey         string        `env:"DB_CLUSTER"     default:"" `
	jaegerURL         string        `env:"DB_CLUSTER"     default:"" `
	thingsAuthURL     string        `env:"DB_CLUSTER"     default:"" `
	usersAuthURL      string        `env:"DB_CLUSTER"     default:"" `
	thingsAuthTimeout time.Duration `env:"DB_CLUSTER"     default:"" `
	usersAuthTimeout  time.Duration `env:"DB_CLUSTER"     default:"" `
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	session := connectToCassandra(cfg.dbCfg, logger)
	defer session.Close()

	conn := internalauth.ConnectToThings(cfg.clientTLS, cfg.caCerts, cfg.thingsAuthURL, svcName, logger)
	defer conn.Close()

	thingsTracer, thingsCloser := internalauth.Jaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)
	authTracer, authCloser := internalauth.Jaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()

	authConn := internalauth.ConnectToAuth(cfg.clientTLS, cfg.caCerts, cfg.usersAuthURL, svcName, logger)
	defer authConn.Close()

	auth := authapi.NewClient(authTracer, authConn, cfg.usersAuthTimeout)

	repo := newService(session, logger)

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(repo, tc, auth, svcName, logger), cfg.serverCert, cfg.serverKey, logger)
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

	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	usersAuthTimeout, err := time.ParseDuration(mainflux.Env(envUsersAuthTimeout, defUsersAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		port:              mainflux.Env(envPort, defPort),
		dbCfg:             dbCfg,
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		serverCert:        mainflux.Env(envServerCert, defServerCert),
		serverKey:         mainflux.Env(envServerKey, defServerKey),
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		usersAuthURL:      mainflux.Env(envUsersAuthURL, defUsersAuthURL),
		usersAuthTimeout:  usersAuthTimeout,
		thingsAuthTimeout: authTimeout,
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

func newService(session *gocql.Session, logger logger.Logger) readers.MessageRepository {
	repo := cassandra.New(session)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("cassandra", "message_reader")
	repo = api.MetricsMiddleware(repo, counter, latency)

	return repo
}

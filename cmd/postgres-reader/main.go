// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	"github.com/mainflux/mainflux/internal/server"
	mfhttpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/postgres"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"golang.org/x/sync/errgroup"
)

const (
	svcName      = "postgres-reader"
	sep          = ","
	stopWaitTime = 5 * time.Second

	defLogLevel          = "error"
	defPort              = "8180"
	defClientTLS         = "false"
	defCACerts           = ""
	defDBHost            = "localhost"
	defDBPort            = "5432"
	defDBUser            = "mainflux"
	defDBPass            = "mainflux"
	defDB                = "mainflux"
	defDBSSLMode         = "disable"
	defDBSSLCert         = ""
	defDBSSLKey          = ""
	defDBSSLRootCert     = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"
	defUsersAuthURL      = "localhost:8181"
	defUsersAuthTimeout  = "1s"

	envLogLevel          = "MF_POSTGRES_READER_LOG_LEVEL"
	envPort              = "MF_POSTGRES_READER_PORT"
	envClientTLS         = "MF_POSTGRES_READER_CLIENT_TLS"
	envCACerts           = "MF_POSTGRES_READER_CA_CERTS"
	envDBHost            = "MF_POSTGRES_READER_DB_HOST"
	envDBPort            = "MF_POSTGRES_READER_DB_PORT"
	envDBUser            = "MF_POSTGRES_READER_DB_USER"
	envDBPass            = "MF_POSTGRES_READER_DB_PASS"
	envDB                = "MF_POSTGRES_READER_DB"
	envDBSSLMode         = "MF_POSTGRES_READER_DB_SSL_MODE"
	envDBSSLCert         = "MF_POSTGRES_READER_DB_SSL_CERT"
	envDBSSLKey          = "MF_POSTGRES_READER_DB_SSL_KEY"
	envDBSSLRootCert     = "MF_POSTGRES_READER_DB_SSL_ROOT_CERT"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
	envUsersAuthURL      = "MF_AUTH_GRPC_URL"
	envUsersAuthTimeout  = "MF_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel          string
	port              string
	clientTLS         bool
	caCerts           string
	dbConfig          postgres.Config
	jaegerURL         string
	thingsAuthURL     string
	usersAuthURL      string
	thingsAuthTimeout time.Duration
	usersAuthTimeout  time.Duration
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

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

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	repo := newService(db, logger)

	hs := mfhttpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(repo, tc, auth, svcName, logger), cfg.dbConfig.SSLCert, cfg.dbConfig.SSLKey, logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Postgres reader service terminated: %s", err))
	}
}

func loadConfig() config {
	dbConfig := postgres.Config{
		Host:        mainflux.Env(envDBHost, defDBHost),
		Port:        mainflux.Env(envDBPort, defDBPort),
		User:        mainflux.Env(envDBUser, defDBUser),
		Pass:        mainflux.Env(envDBPass, defDBPass),
		Name:        mainflux.Env(envDB, defDB),
		SSLMode:     mainflux.Env(envDBSSLMode, defDBSSLMode),
		SSLCert:     mainflux.Env(envDBSSLCert, defDBSSLCert),
		SSLKey:      mainflux.Env(envDBSSLKey, defDBSSLKey),
		SSLRootCert: mainflux.Env(envDBSSLRootCert, defDBSSLRootCert),
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
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		dbConfig:          dbConfig,
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		usersAuthURL:      mainflux.Env(envUsersAuthURL, defUsersAuthURL),
		thingsAuthTimeout: authTimeout,
		usersAuthTimeout:  usersAuthTimeout,
	}
}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, logger logger.Logger) readers.MessageRepository {
	svc := postgres.New(db)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("postgres", "message_reader")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}

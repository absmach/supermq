// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

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

	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/emailer"
	"github.com/mainflux/mainflux/users/token"
	"github.com/mainflux/mainflux/users/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/users/api"
	httpapi "github.com/mainflux/mainflux/users/api/http"
	"github.com/mainflux/mainflux/users/bcrypt"
	"github.com/mainflux/mainflux/users/postgres"
	opentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	jconfig "github.com/uber/jaeger-client-go/config"
)

const (
	defLogLevel      = "error"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "mainflux"
	defDBPass        = "mainflux"
	defDBName        = "users"
	defDBSSLMode     = "disable"
	defDBSSLCert     = ""
	defDBSSLKey      = ""
	defDBSSLRootCert = ""
	defHTTPPort      = "8180"
	defGRPCPort      = "8181"
	defSecret        = "users"
	defServerCert    = ""
	defServerKey     = ""
	defJaegerURL     = ""

	defAuthHTTPPort = "8989"
	defAuthGRPCPort = "8181"
	defAuthTimeout  = "1" // in seconds
	defAuthTLS      = "false"
	defAuthCACerts  = ""
	defAuthURL      = "localhost:8181"

	defEmailLogLevel    = "debug"
	defEmailDriver      = "smtp"
	defEmailHost        = "localhost"
	defEmailPort        = "25"
	defEmailUsername    = "root"
	defEmailPassword    = ""
	defEmailFromAddress = ""
	defEmailFromName    = ""
	defEmailTemplate    = "email.tmpl"

	defTokenSecret        = "mainflux-secret"
	defTokenDuration      = "5"
	defTokenResetEndpoint = "/reset-request" // URL where user lands after click on the reset link from email

	envLogLevel      = "MF_USERS_LOG_LEVEL"
	envDBHost        = "MF_USERS_DB_HOST"
	envDBPort        = "MF_USERS_DB_PORT"
	envDBUser        = "MF_USERS_DB_USER"
	envDBPass        = "MF_USERS_DB_PASS"
	envDBName        = "MF_USERS_DB"
	envDBSSLMode     = "MF_USERS_DB_SSL_MODE"
	envDBSSLCert     = "MF_USERS_DB_SSL_CERT"
	envDBSSLKey      = "MF_USERS_DB_SSL_KEY"
	envDBSSLRootCert = "MF_USERS_DB_SSL_ROOT_CERT"
	envHTTPPort      = "MF_USERS_HTTP_PORT"
	envGRPCPort      = "MF_USERS_GRPC_PORT"
	envSecret        = "MF_USERS_SECRET"
	envServerCert    = "MF_USERS_SERVER_CERT"
	envServerKey     = "MF_USERS_SERVER_KEY"
	envJaegerURL     = "MF_JAEGER_URL"

	envAuthHTTPPort = "MF_AUTH_HTTP_PORT"
	envAuthGRPCPort = "MF_AUTH_GRPC_PORT"
	envAuthTimeout  = "MF_AUTH_TIMEOUT"
	envAuthTLS      = "MF_AUTH_CLIENT_TLS"
	envAuthCACerts  = "MF_AUTH_CA_CERTS"
	envAuthURL      = "MF_AUTH_URL"

	envEmailDriver      = "MF_EMAIL_DRIVER"
	envEmailHost        = "MF_EMAIL_HOST"
	envEmailPort        = "MF_EMAIL_PORT"
	envEmailUsername    = "MF_EMAIL_USERNAME"
	envEmailPassword    = "MF_EMAIL_PASSWORD"
	envEmailFromAddress = "MF_EMAIL_FROM_ADDRESS"
	envEmailFromName    = "MF_EMAIL_FROM_NAME"
	envEmailLogLevel    = "MF_EMAIL_LOG_LEVEL"
	envEmailTemplate    = "MF_EMAIL_TEMPLATE"

	envTokenSecret        = "MF_TOKEN_SECRET"
	envTokenDuration      = "MF_TOKEN_DURATION"
	envTokenResetEndpoint = "MF_TOKEN_RESET_ENDPOINT"
)

type config struct {
	logLevel     string
	dbConfig     postgres.Config
	authHTTPPort string
	authGRPCPort string
	authTimeout  time.Duration
	authTLS      bool
	authCACerts  string
	authURL      string
	emailConf    email.Config
	tokenConf    tokenConfig
	httpPort     string
	grpcPort     string
	secret       string
	serverCert   string
	serverKey    string
	jaegerURL    string
	resetURL     string
}

type tokenConfig struct {
	hmacSampleSecret []byte // secret for signing token
	tokenDuration    string // token in duration in min
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	authTracer, closer := initJaeger("auth", cfg.jaegerURL, logger)
	defer closer.Close()

	auth, close := connectToAuth(cfg, authTracer, logger)
	if close != nil {
		defer close()
	}

	tracer, closer := initJaeger("users", cfg.jaegerURL, logger)
	defer closer.Close()

	dbTracer, dbCloser := initJaeger("users_db", cfg.jaegerURL, logger)
	defer dbCloser.Close()

	svc := newService(db, dbTracer, auth, cfg, logger)
	errs := make(chan error, 2)

	go startHTTPServer(tracer, svc, cfg.httpPort, cfg.serverCert, cfg.serverKey, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Users service terminated: %s", err))
}

func loadConfig() config {
	timeout, err := strconv.ParseInt(mainflux.Env(envAuthTimeout, defAuthTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthTimeout, err.Error())
	}

	tls, err := strconv.ParseBool(mainflux.Env(envAuthTLS, defAuthTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envAuthTLS)
	}

	dbConfig := postgres.Config{
		Host:        mainflux.Env(envDBHost, defDBHost),
		Port:        mainflux.Env(envDBPort, defDBPort),
		User:        mainflux.Env(envDBUser, defDBUser),
		Pass:        mainflux.Env(envDBPass, defDBPass),
		Name:        mainflux.Env(envDBName, defDBName),
		SSLMode:     mainflux.Env(envDBSSLMode, defDBSSLMode),
		SSLCert:     mainflux.Env(envDBSSLCert, defDBSSLCert),
		SSLKey:      mainflux.Env(envDBSSLKey, defDBSSLKey),
		SSLRootCert: mainflux.Env(envDBSSLRootCert, defDBSSLRootCert),
	}

	emailConf := email.Config{
		Driver:      mainflux.Env(envEmailDriver, defEmailDriver),
		FromAddress: mainflux.Env(envEmailFromAddress, defEmailFromAddress),
		FromName:    mainflux.Env(envEmailFromName, defEmailFromName),
		Host:        mainflux.Env(envEmailHost, defEmailHost),
		Port:        mainflux.Env(envEmailPort, defEmailPort),
		Username:    mainflux.Env(envEmailUsername, defEmailUsername),
		Password:    mainflux.Env(envEmailPassword, defEmailPassword),
		Template:    mainflux.Env(envEmailTemplate, defEmailTemplate),
	}

	tokenConf := tokenConfig{
		hmacSampleSecret: []byte(mainflux.Env(envTokenSecret, defTokenSecret)),
		tokenDuration:    mainflux.Env(envTokenDuration, defTokenDuration),
	}

	return config{
		logLevel:     mainflux.Env(envLogLevel, defLogLevel),
		dbConfig:     dbConfig,
		authHTTPPort: mainflux.Env(envAuthHTTPPort, defAuthHTTPPort),
		authGRPCPort: mainflux.Env(envAuthGRPCPort, defAuthGRPCPort),
		authTimeout:  time.Duration(timeout) * time.Second,
		authTLS:      tls,
		authURL:      mainflux.Env(envAuthURL, defAuthURL),
		emailConf:    emailConf,
		tokenConf:    tokenConf,
		httpPort:     mainflux.Env(envHTTPPort, defHTTPPort),
		grpcPort:     mainflux.Env(envGRPCPort, defGRPCPort),
		secret:       mainflux.Env(envSecret, defSecret),
		serverCert:   mainflux.Env(envServerCert, defServerCert),
		serverKey:    mainflux.Env(envServerKey, defServerKey),
		jaegerURL:    mainflux.Env(envJaegerURL, defJaegerURL),
		resetURL:     mainflux.Env(envTokenResetEndpoint, defTokenResetEndpoint),
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
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
		os.Exit(1)
	}

	return tracer, closer
}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func connectToAuth(cfg config, tracer opentracing.Tracer, logger logger.Logger) (mainflux.AuthServiceClient, func() error) {
	var opts []grpc.DialOption
	if cfg.authTLS {
		if cfg.authCACerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(cfg.authCACerts, "")
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

	conn, err := grpc.Dial(cfg.authURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to users service: %s", err))
		os.Exit(1)
	}

	return authapi.NewClient(tracer, conn, cfg.authTimeout), conn.Close
}

func newService(db *sqlx.DB, tracer opentracing.Tracer, auth mainflux.AuthServiceClient, c config, logger logger.Logger) users.Service {
	database := postgres.NewDatabase(db)
	repo := tracing.UserRepositoryMiddleware(postgres.New(database), tracer)
	hasher := bcrypt.New()
	emailer, err := emailer.New(c.resetURL, &c.emailConf)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to configure e-mailing util: %s", err.Error()))
	}
	tDur, err := strconv.Atoi(mainflux.Env(envTokenDuration, defTokenDuration))
	if err != nil {
		logger.Error(err.Error())
	}
	tokenizer := token.New(c.tokenConf.hmacSampleSecret, tDur)

	svc := users.New(repo, hasher, auth, emailer, tokenizer)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "users",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "users",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)
	return svc
}

func startHTTPServer(tracer opentracing.Tracer, svc users.Service, port string, certFile string, keyFile string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	if certFile != "" || keyFile != "" {
		logger.Info(fmt.Sprintf("Users service started using https, cert %s key %s, exposed port %s", certFile, keyFile, port))
		errs <- http.ListenAndServeTLS(p, certFile, keyFile, httpapi.MakeHandler(svc, tracer, logger))
	} else {
		logger.Info(fmt.Sprintf("Users service started using http, exposed port %s", port))
		errs <- http.ListenAndServe(p, httpapi.MakeHandler(svc, tracer, logger))
	}
}

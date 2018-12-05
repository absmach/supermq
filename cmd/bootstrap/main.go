package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"nov/bootstrap"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	api "nov/bootstrap/api"
	"nov/bootstrap/jwt"
	"nov/bootstrap/postgres"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	mfsdk "github.com/mainflux/mainflux/sdk/go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	defLogLevel   = "error"
	defDBHost     = "localhost"
	defDBPort     = "5432"
	defDBUser     = "mainflux"
	defDBPass     = "mainflux"
	defDBName     = "bootstrap"
	defDBSSLMode  = "disable"
	defClientTLS  = "false"
	defCACerts    = ""
	defPort       = "8180"
	defThingsURL  = "localhost:8181"
	defServerCert = ""
	defServerKey  = ""
	defAPIKey     = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1NDQwMzU1ODUsImlhdCI6MTU0Mzk5OTU4NSwiaXNzIjoibWFpbmZsdXgiLCJzdWIiOiJqb2huLmRvZUBlbWFpbC5jb20ifQ.ERo2p9CBeXgU-lFY8EK5sN10xV2ulVVUJ3On1S19QlY"
	defBaseURL    = "http://localhost:8182"
	defConfigFile = "examples/edged/config.yml"

	envLogLevel   = "MF_BOOTSTRAP_LOG_LEVEL"
	envDBHost     = "MF_BOOTSTRAP_DB_HOST"
	envDBPort     = "MF_BOOTSTRAP_DB_PORT"
	envDBUser     = "MF_BOOTSTRAP_DB_USER"
	envDBPass     = "MF_BOOTSTRAP_DB_PASS"
	envDBName     = "MF_BOOTSTRAP_DB"
	envDBSSLMode  = "MF_BOOTSTRAP_DB_SSL_MODE"
	envClientTLS  = "MF_BOOTSTRAP_CLIENT_TLS"
	envCACerts    = "MF_BOOTSTRAP_CA_CERTS"
	envPort       = "MF_BOOTSTRAP_PORT"
	envThingsURL  = "MF_THINGS_URL"
	envServerCert = "MF_BOOTSTRAP_SERVER_CERT"
	envServerKey  = "MF_BOOTSTRAP_SERVER_KEY"
	envAPIKey     = "MF_BOOTSTRAP_API_KEY"
	envBaseURL    = "MF_SDK_BASE_URL"
	envConfigFile = "MF_BOOTSTRAP_CONFIG_FILE"
)

type config struct {
	logLevel   string
	dbHost     string
	dbPort     string
	dbUser     string
	dbPass     string
	dbName     string
	dbSSLMode  string
	clientTLS  bool
	caCerts    string
	httpPort   string
	thingsURL  string
	serverCert string
	serverKey  string
	apiKey     string
	baseURL    string
	configFile string
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(cfg, logger)
	defer db.Close()

	svc := newService(db, logger, cfg)
	errs := make(chan error, 2)

	go startHTTPServer(svc, cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Bootstrap service terminated: %s", err))
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		tls = false
	}
	cfgFile := mainflux.Env(envConfigFile, defConfigFile)

	f, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load configuration file: %s", err.Error()))
	}

	return config{
		logLevel:   mainflux.Env(envLogLevel, defLogLevel),
		dbHost:     mainflux.Env(envDBHost, defDBHost),
		dbPort:     mainflux.Env(envDBPort, defDBPort),
		dbUser:     mainflux.Env(envDBUser, defDBUser),
		dbPass:     mainflux.Env(envDBPass, defDBPass),
		dbName:     mainflux.Env(envDBName, defDBName),
		dbSSLMode:  mainflux.Env(envDBSSLMode, defDBSSLMode),
		clientTLS:  tls,
		caCerts:    mainflux.Env(envCACerts, defCACerts),
		httpPort:   mainflux.Env(envPort, defPort),
		thingsURL:  mainflux.Env(envThingsURL, defThingsURL),
		serverCert: mainflux.Env(envServerCert, defServerCert),
		serverKey:  mainflux.Env(envServerKey, defServerKey),
		apiKey:     mainflux.Env(envAPIKey, defAPIKey),
		baseURL:    mainflux.Env(envBaseURL, defBaseURL),
		configFile: string(f),
	}
}

func connectToDB(cfg config, logger logger.Logger) *sql.DB {
	db, err := postgres.Connect(cfg.dbHost, cfg.dbPort, cfg.dbName, cfg.dbUser, cfg.dbPass, cfg.dbSSLMode)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sql.DB, logger logger.Logger, cfg config) bootstrap.Service {
	thingsRepo := postgres.NewThingRepository(db, logger)

	config := mfsdk.Config{
		BaseURL: cfg.baseURL,
	}
	sdk := mfsdk.NewSDK(config)
	svc := bootstrap.New(thingsRepo, cfg.apiKey, jwt.New(), sdk, cfg.configFile)
	svc = api.NewLoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "bootstrap",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "bootstrap",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)
	return svc
}

func startHTTPServer(svc bootstrap.Service, cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	if cfg.serverCert != "" || cfg.serverKey != "" {
		logger.Info(fmt.Sprintf("Bootstrap service started using https on port %s with cert %s key %s",
			cfg.httpPort, cfg.serverCert, cfg.serverKey))
		errs <- http.ListenAndServeTLS(p, cfg.serverCert, cfg.serverKey, api.MakeHandler(svc, bootstrap.NewConfigReader()))
	} else {
		logger.Info(fmt.Sprintf("Bootstrap service started using http on port %s", cfg.httpPort))
		errs <- http.ListenAndServe(p, api.MakeHandler(svc, bootstrap.NewConfigReader()))
	}
}

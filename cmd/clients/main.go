package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-zoo/bone"
	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/clients/clients"
	capi "github.com/mainflux/mainflux/clients/clients/api"
	"github.com/mainflux/mainflux/clients/clients/emailer"
	cpostgres "github.com/mainflux/mainflux/clients/clients/postgres"
	ctracing "github.com/mainflux/mainflux/clients/clients/tracing"
	"github.com/mainflux/mainflux/clients/groups"
	gapi "github.com/mainflux/mainflux/clients/groups/api"
	gpostgres "github.com/mainflux/mainflux/clients/groups/postgres"
	gtracing "github.com/mainflux/mainflux/clients/groups/tracing"
	"github.com/mainflux/mainflux/clients/hasher"
	"github.com/mainflux/mainflux/clients/jwt"
	"github.com/mainflux/mainflux/clients/policies"
	grpcapi "github.com/mainflux/mainflux/clients/policies/api/grpc"
	papi "github.com/mainflux/mainflux/clients/policies/api/http"
	ppostgres "github.com/mainflux/mainflux/clients/policies/postgres"
	ppracing "github.com/mainflux/mainflux/clients/policies/tracing"
	"github.com/mainflux/mainflux/clients/postgres"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	svcName      = "clients"
	stopWaitTime = 5 * time.Second

	defLogLevel           = "debug"
	defSecretKey          = "clientsecret"
	defAdminIdentity      = "admin@example.com"
	defAdminSecret        = "12345678"
	defDBHost             = "localhost"
	defDBPort             = "5432"
	defDBUser             = "mainflux"
	defDBPass             = "mainflux"
	defDB                 = "clients"
	defDBSSLMode          = "disable"
	defDBSSLCert          = ""
	defDBSSLKey           = ""
	defDBSSLRootCert      = ""
	defHTTPPort           = "9191"
	defGRPCPort           = "9192"
	defServerCert         = ""
	defServerKey          = ""
	defJaegerURL          = "http://localhost:6831"
	defPassRegex          = "^.{8,}$"
	defEmailHost          = "localhost"
	defEmailPort          = "25"
	defEmailUsername      = "root"
	defEmailPassword      = ""
	defEmailFromAddress   = ""
	defEmailFromName      = ""
	defEmailTemplate      = "email.tmpl"
	defTokenResetEndpoint = "/reset-request" // URL where user lands after click on the reset link from email

	envLogLevel           = "MF_CLIENTS_LOG_LEVEL"
	envSecretKey          = "MF_CLIENTS_SECRET_KEY"
	envAdminIdentity      = "MF_CLIENTS_ADMIN_EMAIL"
	envAdminSecret        = "MF_CLIENTS_ADMIN_PASSWORD"
	envDBHost             = "MF_CLIENTS_DB_HOST"
	envDBPort             = "MF_CLIENTS_DB_PORT"
	envDBUser             = "MF_CLIENTS_DB_USER"
	envDBPass             = "MF_CLIENTS_DB_PASS"
	envDB                 = "MF_CLIENTS_DB"
	envDBSSLMode          = "MF_CLIENTS_DB_SSL_MODE"
	envDBSSLCert          = "MF_CLIENTS_DB_SSL_CERT"
	envDBSSLKey           = "MF_CLIENTS_DB_SSL_KEY"
	envDBSSLRootCert      = "MF_CLIENTS_DB_SSL_ROOT_CERT"
	envHTTPPort           = "MF_CLIENTS_HTTP_PORT"
	envGRPCPort           = "MF_CLIENTS_GRPC_PORT"
	envServerCert         = "MF_CLIENTS_SERVER_CERT"
	envServerKey          = "MF_CLIENTS_SERVER_KEY"
	envJaegerURL          = "MF_CLIENTS_JAEGER_URL"
	envPassRegex          = "MF_USERS_PASS_REGEX"
	envEmailHost          = "MF_EMAIL_HOST"
	envEmailPort          = "MF_EMAIL_PORT"
	envEmailUsername      = "MF_EMAIL_USERNAME"
	envEmailPassword      = "MF_EMAIL_PASSWORD"
	envEmailFromAddress   = "MF_EMAIL_FROM_ADDRESS"
	envEmailFromName      = "MF_EMAIL_FROM_NAME"
	envEmailTemplate      = "MF_EMAIL_TEMPLATE"
	envTokenResetEndpoint = "MF_TOKEN_RESET_ENDPOINT"
)

type config struct {
	logLevel      string
	secretKey     string
	adminIdentity string
	adminSecret   string
	dbConfig      postgres.Config
	emailConf     email.Config
	resetURL      string
	httpPort      string
	grpcPort      string
	serverCert    string
	serverKey     string
	jaegerURL     string
	passRegex     *regexp.Regexp
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	tp, err := initJaeger(svcName, cfg.jaegerURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger: %s", err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Error(fmt.Sprintf("Error shutting down tracer provider: %v", err))
		}
	}()
	tracer := otel.Tracer(svcName)

	csvc, gsvc, psvc := newService(db, tracer, cfg, logger)

	g.Go(func() error {
		return startHTTPServer(ctx, csvc, gsvc, psvc, cfg.httpPort, cfg.serverCert, cfg.serverKey, logger)
	})
	g.Go(func() error {
		return startGRPCServer(ctx, csvc, psvc, cfg.grpcPort, cfg.serverCert, cfg.serverKey, logger)
	})

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("%s service shutdown by signal: %s", svcName, sig))
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func loadConfig() config {
	passRegex, err := regexp.Compile(mainflux.Env(envPassRegex, defPassRegex))
	if err != nil {
		log.Fatalf("Invalid password validation rules %s\n", envPassRegex)
	}

	emailConf := email.Config{
		FromAddress: mainflux.Env(envEmailFromAddress, defEmailFromAddress),
		FromName:    mainflux.Env(envEmailFromName, defEmailFromName),
		Host:        mainflux.Env(envEmailHost, defEmailHost),
		Port:        mainflux.Env(envEmailPort, defEmailPort),
		Username:    mainflux.Env(envEmailUsername, defEmailUsername),
		Password:    mainflux.Env(envEmailPassword, defEmailPassword),
		Template:    mainflux.Env(envEmailTemplate, defEmailTemplate),
	}

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

	return config{
		logLevel:      mainflux.Env(envLogLevel, defLogLevel),
		secretKey:     mainflux.Env(envSecretKey, defSecretKey),
		adminIdentity: mainflux.Env(envAdminIdentity, defAdminIdentity),
		adminSecret:   mainflux.Env(envAdminSecret, defAdminSecret),
		dbConfig:      dbConfig,
		emailConf:     emailConf,
		resetURL:      mainflux.Env(envTokenResetEndpoint, defTokenResetEndpoint),
		httpPort:      mainflux.Env(envHTTPPort, defHTTPPort),
		grpcPort:      mainflux.Env(envGRPCPort, defGRPCPort),
		serverCert:    mainflux.Env(envServerCert, defServerCert),
		serverKey:     mainflux.Env(envServerKey, defServerKey),
		jaegerURL:     mainflux.Env(envJaegerURL, defJaegerURL),
		passRegex:     passRegex,
	}

}

func initJaeger(svcName, url string) (*tracesdk.TracerProvider, error) {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
		tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(exporter)),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(svcName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, tracer trace.Tracer, c config, logger logger.Logger) (clients.Service, groups.GroupService, policies.PolicyService) {
	database := postgres.NewDatabase(db, tracer)
	cRepo := cpostgres.NewClientRepo(database)
	gRepo := gpostgres.NewGroupRepo(database)
	pRepo := ppostgres.NewPolicyRepo(database)

	idp := uuid.New()
	hsr := hasher.New()

	tokenizer := jwt.NewTokenRepo([]byte(c.secretKey))
	tokenizer = jwt.NewTokenRepoMiddleware(tokenizer, tracer)

	emailer, err := emailer.New(c.resetURL, &c.emailConf)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to configure e-mailing util: %s", err.Error()))
	}

	csvc := clients.NewService(cRepo, pRepo, tokenizer, emailer, hsr, idp, c.passRegex)
	gsvc := groups.NewService(gRepo, pRepo, tokenizer, idp)
	psvc := policies.NewService(pRepo, tokenizer, idp)

	csvc = ctracing.TracingMiddleware(csvc, tracer)
	csvc = capi.LoggingMiddleware(csvc, logger)
	csvc = capi.MetricsMiddleware(
		csvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: svcName,
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: svcName,
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	gsvc = gtracing.TracingMiddleware(gsvc, tracer)
	gsvc = gapi.LoggingMiddleware(gsvc, logger)
	gsvc = gapi.MetricsMiddleware(
		gsvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "clients_groups",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "clients_groups",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	psvc = ppracing.TracingMiddleware(psvc, tracer)
	psvc = papi.LoggingMiddleware(psvc, logger)
	psvc = papi.MetricsMiddleware(
		psvc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "client_policies",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "client_policies",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	if err := createAdmin(c, cRepo, hsr, csvc); err != nil {
		logger.Error(fmt.Sprintf("Failed to create admin client: %s", err))
		os.Exit(1)
	}
	return csvc, gsvc, psvc
}

func startHTTPServer(ctx context.Context, csvc clients.Service, gsvc groups.Service, psvc policies.Service, port string, certFile string, keyFile string, logger logger.Logger) error {
	p := fmt.Sprintf(":%s", port)
	errCh := make(chan error)
	m := bone.New()
	capi.MakeClientsHandler(csvc, m, logger)
	gapi.MakeGroupsHandler(gsvc, m, logger)
	papi.MakePolicyHandler(psvc, m, logger)
	server := &http.Server{Addr: p, Handler: m}

	switch {
	case certFile != "" || keyFile != "":
		logger.Info(fmt.Sprintf("Clients service started using https, cert %s key %s, exposed port %s", certFile, keyFile, port))
		go func() {
			errCh <- server.ListenAndServeTLS(certFile, keyFile)
		}()
	default:
		logger.Info(fmt.Sprintf("Clients service started using http, exposed port %s", port))
		go func() {
			errCh <- server.ListenAndServe()
		}()
	}

	select {
	case <-ctx.Done():
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), stopWaitTime)
		defer cancelShutdown()
		if err := server.Shutdown(ctxShutdown); err != nil {
			logger.Error(fmt.Sprintf("Clients service error occurred during shutdown at %s: %s", p, err))
			return fmt.Errorf("clients service occurred during shutdown at %s: %w", p, err)
		}
		logger.Info(fmt.Sprintf("Clients service shutdown of http at %s", p))
		return nil
	case err := <-errCh:
		return err
	}

}

func startGRPCServer(ctx context.Context, csvc clients.Service, psvc policies.Service, port string, certFile string, keyFile string, logger logger.Logger) error {
	p := fmt.Sprintf(":%s", port)
	errCh := make(chan error)

	listener, err := net.Listen("tcp", p)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	var server *grpc.Server
	switch {
	case certFile != "" || keyFile != "":
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("failed to load auth certificates: %w", err)
		}
		logger.Info(fmt.Sprintf("Clients gRPC service started using https on port %s with cert %s key %s", port, certFile, keyFile))
		server = grpc.NewServer(grpc.Creds(creds))
	default:
		logger.Info(fmt.Sprintf("Clients gRPC service started using http on port %s", port))
		server = grpc.NewServer()
	}

	reflection.Register(server)
	policies.RegisterAuthServiceServer(server, grpcapi.NewServer(csvc, psvc))
	logger.Info(fmt.Sprintf("Clients gRPC service started, exposed port %s", port))
	go func() {
		errCh <- server.Serve(listener)
	}()

	select {
	case <-ctx.Done():
		c := make(chan bool)
		go func() {
			defer close(c)
			server.GracefulStop()
		}()
		select {
		case <-c:
		case <-time.After(stopWaitTime):
		}
		logger.Info(fmt.Sprintf("Authentication gRPC service shutdown at %s", p))
		return nil
	case err := <-errCh:
		return err
	}
}

func createAdmin(c config, crepo clients.ClientRepository, hsr clients.Hasher, svc clients.Service) error {
	id, err := uuid.New().ID()
	if err != nil {
		return err
	}
	hash, err := hsr.Hash(c.adminSecret)
	if err != nil {
		return err
	}

	client := clients.Client{
		ID:   id,
		Name: "admin",
		Credentials: clients.Credentials{
			Identity: c.adminIdentity,
			Secret:   hash,
		},
		Metadata: clients.Metadata{
			"role": "admin",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role:      clients.AdminRole,
		Status:    clients.EnabledStatus,
	}

	if _, err := crepo.RetrieveByIdentity(context.Background(), client.Credentials.Identity); err == nil {
		return nil
	}

	// Create an admin
	if _, err = crepo.Save(context.Background(), client); err != nil {
		return err
	}
	_, err = svc.IssueToken(context.Background(), c.adminIdentity, c.adminSecret)
	if err != nil {
		return err
	}

	return nil
}

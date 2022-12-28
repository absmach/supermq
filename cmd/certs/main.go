// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/certs/api"
	vault "github.com/mainflux/mainflux/certs/pki"
	"github.com/mainflux/mainflux/certs/postgres"
	certsPg "github.com/mainflux/mainflux/certs/postgres"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/env"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"golang.org/x/sync/errgroup"

	"github.com/jmoiron/sqlx"
	authClient "github.com/mainflux/mainflux/internal/client/grpc/auth"
	pgClient "github.com/mainflux/mainflux/internal/client/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

const (
	svcName       = "certs"
	envPrefix     = "MF_CERTS_"
	envPrefixHttp = "MF_CERTS_HTTP_"
)

var (
	errFailedCertLoading     = errors.New("failed to load certificate")
	errFailedCertDecode      = errors.New("failed to decode certificate")
	errCACertificateNotExist = errors.New("CA certificate does not exist")
	errCAKeyNotExist         = errors.New("CA certificate key does not exist")
)

type config struct {
	logLevel  string `env:"MF_CERTS_LOG_LEVEL"        envDefault:"debug"`
	certsURL  string `env:"MF_SDK_CERTS_URL"          envDefault:"http://localhost"`
	thingsURL string `env:"MF_THINGS_URL"             envDefault:"http://things:8182"`
	jaegerURL string `env:"MF_JAEGER_URL"             envDefault:""`

	// Sign and issue certificates without 3rd party PKI
	signCAPath    string `env:"MF_CERTS_SIGN_CA_PATH"        envDefault:"ca.crt"`
	signCAKeyPath string `env:"MF_CERTS_SIGN_CA_KEY_PATH"    envDefault:"ca.key"`
	// used in pki mock
	signRSABits    int    `env:"MF_CERTS_SIGN_RSA_BITS"       envDefault:""`
	signHoursValid string `env:"MF_CERTS_SIGN_HOURS_VALID"    envDefault:"2048h"`

	// 3rd party PKI API access settings
	pkiPath  string `env:"MF_CERTS_VAULT_HOST"         envDefault:"pki_int"`
	pkiToken string `env:"MF_VAULT_PKI_INT_PATH"       envDefault:""`
	pkiHost  string `env:"MF_VAULT_CA_ROLE_NAME"       envDefault:""`
	pkiRole  string `env:"MF_VAULT_TOKEN"              envDefault:"mainflux"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Failed to load %s configuration : %s", svcName, err.Error())
	}

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tlsCert, caCert, err := loadCertificates(cfg)
	if err != nil {
		logger.Error("Failed to load CA certificates for issuing client certs")
	}

	if cfg.pkiHost == "" {
		log.Fatalf("No host specified for PKI engine")
	}

	pkiClient, err := vault.NewVaultClient(cfg.pkiToken, cfg.pkiHost, cfg.pkiPath, cfg.pkiRole)
	if err != nil {
		log.Fatalf("Failed to configure client for PKI engine")
	}

	db, err := pgClient.Setup(envPrefix, *certsPg.Migration())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	auth, authGrpcClient, authTracerCloser, authGrpcSecure, err := authClient.Setup(envPrefix, cfg.jaegerURL)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer authGrpcClient.Close()
	defer authTracerCloser.Close()
	logger.Info("Successfully connected to auth grpc server " + authGrpcSecure)

	svc := newService(auth, db, logger, nil, tlsCert, caCert, cfg, pkiClient)

	httpServerConfig := server.Config{}
	if err := env.Parse(&httpServerConfig, env.Options{Prefix: envPrefixHttp, AltPrefix: envPrefix}); err != nil {
		log.Fatalf(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
	}
	hs := httpserver.New(ctx, cancel, svcName, httpServerConfig, api.MakeHandler(svc, logger), logger)

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Certs service terminated: %s", err))
	}
}

func newService(auth mainflux.AuthServiceClient, db *sqlx.DB, logger logger.Logger, esClient *redis.Client, tlsCert tls.Certificate, x509Cert *x509.Certificate, cfg config, pkiAgent vault.Agent) certs.Service {
	certsRepo := postgres.NewRepository(db, logger)
	config := mfsdk.Config{
		CertsURL:  cfg.certsURL,
		ThingsURL: cfg.thingsURL,
	}
	sdk := mfsdk.NewSDK(config)
	svc := certs.New(auth, certsRepo, sdk, pkiAgent)
	svc = api.NewLoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)
	return svc
}

func loadCertificates(conf config) (tls.Certificate, *x509.Certificate, error) {
	var tlsCert tls.Certificate
	var caCert *x509.Certificate

	if conf.signCAPath == "" || conf.signCAKeyPath == "" {
		return tlsCert, caCert, nil
	}

	if _, err := os.Stat(conf.signCAPath); os.IsNotExist(err) {
		return tlsCert, caCert, errCACertificateNotExist
	}

	if _, err := os.Stat(conf.signCAKeyPath); os.IsNotExist(err) {
		return tlsCert, caCert, errCAKeyNotExist
	}

	tlsCert, err := tls.LoadX509KeyPair(conf.signCAPath, conf.signCAKeyPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertLoading, err)
	}

	b, err := ioutil.ReadFile(conf.signCAPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertLoading, err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		log.Fatalf("No PEM data found, failed to decode CA")
	}

	caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertDecode, err)
	}

	return tlsCert, caCert, nil
}

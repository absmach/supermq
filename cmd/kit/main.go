//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/kit"
	"github.com/mainflux/mainflux/kit/api"
	"github.com/mainflux/mainflux/logger"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	kithttpapi "github.com/mainflux/mainflux/kit/api/kit/http"
	opentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	jconfig "github.com/uber/jaeger-client-go/config"
)

const (
	defLogLevel   = "error"
	defHTTPPort   = "9021"
	defJaegerURL  = ""
	defServerCert = ""
	defServerKey  = ""
	defSecret     = "secret"

	envLogLevel   = "MF_KIT_LOG_LEVEL"
	envHTTPPort   = "MF_KIT_HTTP_PORT"
	envJaegerURL  = "MF_JAEGER_URL"
	envServerCert = "MF_KIT_SERVER_CERT"
	envServerKey  = "MF_KIT_SERVER_KEY"
	envSecret     = "MF_KIT_SECRET"
)

type config struct {
	logLevel     string
	httpPort     string
	authHTTPPort string
	authGRPCPort string
	jaegerURL    string
	serverCert   string
	serverKey    string
	secret       string
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	kitTracer, kitCloser := initJaeger("kit", cfg.jaegerURL, logger)
	defer kitCloser.Close()

	svc := newService(cfg.secret, logger)
	errs := make(chan error, 2)

	go startHTTPServer(kithttpapi.MakeHandler(kitTracer, svc), cfg.httpPort, cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Kit service terminated: %s", err))
}

func loadConfig() config {
	return config{
		logLevel:     mainflux.Env(envLogLevel, defLogLevel),
		httpPort:     mainflux.Env(envHTTPPort, defHTTPPort),
		authHTTPPort: mainflux.Env(envAuthHTTPPort, defAuthHTTPPort),
		authGRPCPort: mainflux.Env(envAuthGRPCPort, defAuthGRPCPort),
		serverCert:   mainflux.Env(envServerCert, defServerCert),
		serverKey:    mainflux.Env(envServerKey, defServerKey),
		secret:       mainflux.Env(envSecret, defSecret),
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
		logger.Error(fmt.Sprintf("Failed to init Jaeger client: %s", err))
		os.Exit(1)
	}

	return tracer, closer
}

func newService(secret string, logger logger.Logger) kit.Service {
	svc := kit.New(secret)

	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "kit",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "kit",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	return svc
}

func startHTTPServer(handler http.Handler, port string, cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	if cfg.serverCert != "" || cfg.serverKey != "" {
		logger.Info(fmt.Sprintf("Kit service started using https on port %s with cert %s key %s",
			port, cfg.serverCert, cfg.serverKey))
		errs <- http.ListenAndServeTLS(p, cfg.serverCert, cfg.serverKey, handler)
		return
	}
	logger.Info(fmt.Sprintf("Kit service started using http on port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, handler)
}

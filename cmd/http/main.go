package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux"
	adapter "github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/http/api"
	"github.com/mainflux/mainflux/http/nats"
	manager "github.com/mainflux/mainflux/manager/client"
	broker "github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	defPort       string = "8180"
	defNatsURL    string = broker.DefaultURL
	defManagerURL string = "http://localhost:8180"
	envPort       string = "MF_HTTP_ADAPTER_PORT"
	envNatsURL    string = "MF_NATS_URL"
	envManagerURL string = "MF_MANAGER_URL"
)

type config struct {
	ManagerURL string
	NatsURL    string
	Port       string
}

func main() {
	cfg := config{
		ManagerURL: mainflux.Env(envManagerURL, defManagerURL),
		NatsURL:    mainflux.Env(envNatsURL, defNatsURL),
		Port:       mainflux.Env(envPort, defPort),
	}

	logger := log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	nc, err := broker.Connect(cfg.NatsURL)
	if err != nil {
		logger.Log("aborted", err)
		os.Exit(1)
	}
	defer nc.Close()

	repo := nats.NewMessagePublisher(nc)
	svc := adapter.NewService(repo)

	svc = api.NewLoggingService(logger, svc)

	fields := []string{"method"}
	svc = api.NewMetricService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "http_adapter",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fields),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "http_adapter",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fields),
		svc,
	)

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.Port)
		mc := manager.NewClient(cfg.ManagerURL)
		errs <- http.ListenAndServe(p, api.MakeHandler(svc, mc))
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	logger.Log("terminated", <-errs)
}

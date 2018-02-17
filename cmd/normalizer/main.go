package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/normalizer"
	nats "github.com/nats-io/go-nats"
	"go.uber.org/zap"
)

const (
	defNatsURL string = nats.DefaultURL
	defPort    string = "8180"
	envNatsURL string = "MF_NATS_URL"
	envPort    string = "MF_NORMALIZER_PORT"
)

type config struct {
	NatsURL string
	Port    string
}

func main() {
	cfg := config{
		NatsURL: mainflux.Env(envNatsURL, defNatsURL),
		Port:    mainflux.Env(envPort, defPort),
	}

	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error("NATS connection failure.", zap.Error(err))
		os.Exit(1)
	}
	defer nc.Close()

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.Port)
		errs <- http.ListenAndServe(p, normalizer.MakeHandler())
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	normalizer.Subscribe(nc)
	logger.Info("Service terminated.", zap.Error(<-errs))
}

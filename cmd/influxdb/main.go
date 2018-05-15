package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gogo/protobuf/proto"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	influxdb "github.com/mainflux/mainflux/influxdb"
	log "github.com/mainflux/mainflux/logger"
	nats "github.com/nats-io/go-nats"
)

const (
	defNatsURL   = nats.DefaultURL
	defPort      = "8180"
	defPointName = "messages"
	defDBName    = "mainflux"
	defDBHost    = "localhost"
	defDBPort    = "8086"
	defDBUser    = "mainflux"
	defDBPass    = "mainflux"

	envNatsURL   = "MF_NATS_URL"
	envPort      = "MF_INFLUXDB_WRITER_PORT"
	envPointName = "MF_INFLUXDB_POINT_NAME"
	envDBName    = "MF_INFLUXDB_DB_NAME"
	envDBHost    = "MF_INFLUXDB_DB_HOST"
	envDBPort    = "MF_INFLUXDB_DB_PORT"
	envDBUser    = "MF_INFLUXDB_DB_USER"
	envDBPass    = "MF_INFLUXDB_DB_PASS"
)

type config struct {
	NatsURL   string
	Port      string
	PointName string
	DBName    string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
}

func main() {
	cfg := config{
		NatsURL:   mainflux.Env(envNatsURL, defNatsURL),
		PointName: mainflux.Env(envPointName, defPointName),
		Port:      mainflux.Env(envPort, defPort),
		DBName:    mainflux.Env(envDBName, defDBName),
		DBHost:    mainflux.Env(envDBHost, defDBHost),
		DBPort:    mainflux.Env(envDBPort, defDBPort),
		DBUser:    mainflux.Env(envDBUser, defDBUser),
		DBPass:    mainflux.Env(envDBPass, defDBPass),
	}

	logger := log.New(os.Stdout)

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nc.Close()

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.Port)
		logger.Info(fmt.Sprintf("Influxdb writer service started, exposed port %s", cfg.Port))
		errs <- http.ListenAndServe(p, influxdb.MakeHandler())
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	clientCfg := client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", cfg.DBHost, cfg.DBPort),
		Username: cfg.DBUser,
		Password: cfg.DBPass,
	}

	writer, err := influxdb.New(clientCfg, cfg.DBName, cfg.PointName)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB writer: %s", err.Error()))
		return
	}

	nc.Subscribe("normalized", handleMsg(logger, writer))
	// normalizer.Subscribe(nc, logger, counter, latency)

	err = <-errs
	logger.Error(fmt.Sprintf("Influxdb writer service terminated: %s", err))
}

func handleMsg(logger log.Logger, w influxdb.Writer) nats.MsgHandler {

	return func(m *nats.Msg) {
		msg := &mainflux.Message{}

		if err := proto.Unmarshal(m.Data, msg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
		}

		if err := w.Save(*msg); err != nil {
			logger.Warn(fmt.Sprintf("InfluxDB Writer failed to save message: %s", err))
			return
		}
	}
}

package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/clients"
	"github.com/mainflux/mainflux/clients/api"
	httpapi "github.com/mainflux/mainflux/clients/api/http"
	"github.com/mainflux/mainflux/clients/bcrypt"
	"github.com/mainflux/mainflux/clients/jwt"
	"github.com/mainflux/mainflux/clients/postgres"
	log "github.com/mainflux/mainflux/logger"
	pb "github.com/mainflux/mainflux/users/api/grpc"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

const (
	defDBHost    = "localhost"
	defDBPort    = "5432"
	defDBUser    = "mainflux"
	defDBPass    = "mainflux"
	defDBName    = "clients"
	defPort      = "8180"
	defUsersAddr = "localhost:8181"
	defSecret    = "clients"
	envDBHost    = "MF_DB_HOST"
	envDBPort    = "MF_DB_PORT"
	envDBUser    = "MF_DB_USER"
	envDBPass    = "MF_DB_PASS"
	envDBName    = "MF_CLIENTS_DB"
	envPort      = "MF_CLIENTS_PORT"
	envUsersAddr = "MF_USERS_ADDR"
	envSecret    = "MF_CLIENTS_SECRET"
)

type config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	Port      string
	UsersAddr string
	Secret    string
}

func main() {
	cfg := config{
		DBHost:    mainflux.Env(envDBHost, defDBHost),
		DBPort:    mainflux.Env(envDBPort, defDBPort),
		DBUser:    mainflux.Env(envDBUser, defDBUser),
		DBPass:    mainflux.Env(envDBPass, defDBPass),
		DBName:    mainflux.Env(envDBName, defDBName),
		Port:      mainflux.Env(envPort, defPort),
		UsersAddr: mainflux.Env(envUsersAddr, defUsersAddr),
		Secret:    mainflux.Env(envSecret, defSecret),
	}

	logger := log.New(os.Stdout)

	db, err := postgres.Connect(cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPass)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	defer db.Close()

	conn, err := grpc.Dial(cfg.UsersAddr, grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to users service: %s", err))
		os.Exit(1)
	}
	defer conn.Close()

	users := pb.NewClient(conn)
	clientsRepo := postgres.NewClientRepository(db)
	channelsRepo := postgres.NewChannelRepository(db)
	hasher := bcrypt.New()
	idp := jwt.New(cfg.Secret)

	svc := clients.New(users, clientsRepo, channelsRepo, hasher, idp)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "clients",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "clients",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.Port)
		logger.Info(fmt.Sprintf("Clients service started, exposed port %s", cfg.Port))
		errs <- http.ListenAndServe(p, httpapi.MakeHandler(svc))
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Clients service terminated: %s", err))
}

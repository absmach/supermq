package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/manager"
	"github.com/mainflux/mainflux/manager/api"
	"github.com/mainflux/mainflux/manager/bcrypt"
	"github.com/mainflux/mainflux/manager/jwt"
	"github.com/mainflux/mainflux/manager/postgres"
	pb "github.com/mainflux/mainflux/users/api/grpc"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

const (
	defDBHost    = "localhost"
	defDBPort    = "5432"
	defDBUser    = "mainflux"
	defDBPass    = "mainflux"
	defDBName    = "manager"
	defPort      = "8180"
	defUsersAddr = "localhost:8181"
	defSecret    = "manager"
	envDBHost    = "MF_DB_HOST"
	envDBPort    = "MF_DB_PORT"
	envDBUser    = "MF_DB_USER"
	envDBPass    = "MF_DB_PASS"
	envDBName    = "MF_MANAGER_DB"
	envPort      = "MF_MANAGER_PORT"
	envUsersAddr = "MF_USERS_ADDR"
	envSecret    = "MF_MANAGER_SECRET"
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

	users := pb.NewUsersServiceClient(conn)
	clients := postgres.NewClientRepository(db)
	channels := postgres.NewChannelRepository(db)
	hasher := bcrypt.New()
	idp := jwt.New(cfg.Secret)

	svc := manager.New(users, clients, channels, hasher, idp)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "manager",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "manager",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.Port)
		logger.Info(fmt.Sprintf("Manager service started, exposed port %s", cfg.Port))
		errs <- http.ListenAndServe(p, api.MakeHandler(svc))
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Manager service terminated: %s", err))
}

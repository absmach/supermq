package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/api"
	"github.com/mainflux/mainflux/users/bcrypt"
	"github.com/mainflux/mainflux/users/jwt"
	"github.com/mainflux/mainflux/users/postgres"
)

const (
	defDBHost = "localhost"
	defDBPort = "5432"
	defDBUser = "mainflux"
	defDBPass = "mainflux"
	defDBName = "users"
	defPort   = "8180"
	defSecret = "users"
	envDBHost = "MF_USERS_DB_HOST"
	envDBPort = "MF_USERS_DB_PORT"
	envDBUser = "MF_USERS_DB_USER"
	envDBPass = "MF_USERS_DB_PASS"
	envDBName = "MF_USERS_DB"
	envPort   = "MF_USERS_PORT"
	envSecret = "MF_USERS_SECRET"
)

type config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
	Port   string
	Secret string
}

func main() {
	cfg := config{
		DBHost: mainflux.Env(envDBHost, defDBHost),
		DBPort: mainflux.Env(envDBPort, defDBPort),
		DBUser: mainflux.Env(envDBUser, defDBUser),
		DBPass: mainflux.Env(envDBPass, defDBPass),
		DBName: mainflux.Env(envDBName, defDBName),
		Port:   mainflux.Env(envPort, defPort),
		Secret: mainflux.Env(envSecret, defSecret),
	}

	logger := log.New(os.Stdout)

	db, err := postgres.Connect(cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBUser, cfg.DBPass)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	defer db.Close()

	repo := postgres.New(db)
	hasher := bcrypt.New()
	idp := jwt.New(cfg.Secret)

	svc := users.New(repo, hasher, idp)

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.Port)
		logger.Info(fmt.Sprintf("Users service started, exposed port %s", cfg.Port))
		errs <- http.ListenAndServe(p, api.MakeHandler(svc))
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Users service terminated: %s", err))
}

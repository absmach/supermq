// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"

	chclient "github.com/absmach/callhome/pkg/client"
	"github.com/absmach/supermq"
	"github.com/absmach/supermq/internal/email"
	smqlog "github.com/absmach/supermq/logger"
	"github.com/absmach/supermq/notifications"
	httpapi "github.com/absmach/supermq/notifications/api"
	"github.com/absmach/supermq/notifications/emailer"
	"github.com/absmach/supermq/notifications/events"
	notificationspg "github.com/absmach/supermq/notifications/postgres"
	"github.com/absmach/supermq/pkg/events/store"
	jaegerclient "github.com/absmach/supermq/pkg/jaeger"
	pgclient "github.com/absmach/supermq/pkg/postgres"
	"github.com/absmach/supermq/pkg/server"
	httpserver "github.com/absmach/supermq/pkg/server/http"
	"github.com/absmach/supermq/pkg/uuid"
	"github.com/caarlos0/env/v11"
	"golang.org/x/sync/errgroup"
)

const (
	svcName            = "notifications"
	envPrefixUsersDB   = "SMQ_NOTIFICATIONS_USERS_DB_"
	envPrefixDomainsDB = "SMQ_NOTIFICATIONS_DOMAINS_DB_"
	envPrefixHTTP      = "SMQ_NOTIFICATIONS_HTTP_"
	defSvcHTTPPort     = "9031"
)

type config struct {
	LogLevel                        string  `env:"SMQ_NOTIFICATIONS_LOG_LEVEL"                    envDefault:"info"`
	ESURL                           string  `env:"SMQ_ES_URL"                                     envDefault:"nats://localhost:4222"`
	JaegerURL                       url.URL `env:"SMQ_JAEGER_URL"                                 envDefault:"http://localhost:4318/v1/traces"`
	SendTelemetry                   bool    `env:"SMQ_SEND_TELEMETRY"                             envDefault:"true"`
	InstanceID                      string  `env:"SMQ_NOTIFICATIONS_INSTANCE_ID"                  envDefault:""`
	TraceRatio                      float64 `env:"SMQ_JAEGER_TRACE_RATIO"                         envDefault:"1.0"`
	InvitationSentEmailTemplate     string  `env:"SMQ_NOTIFICATIONS_INVITATION_SENT_TEMPLATE"     envDefault:"email.tmpl"`
	InvitationAcceptedEmailTemplate string  `env:"SMQ_NOTIFICATIONS_INVITATION_ACCEPTED_TEMPLATE" envDefault:"email.tmpl"`
	EmailFooter                     string  `env:"SMQ_NOTIFICATIONS_EMAIL_FOOTER"                 envDefault:"Notification Service"`
	EmailSubjectPrefix              string  `env:"SMQ_NOTIFICATIONS_EMAIL_SUBJECT_PREFIX"         envDefault:""`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := smqlog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}

	var exitCode int
	defer smqlog.ExitWithError(&exitCode)

	if cfg.InstanceID == "" {
		if cfg.InstanceID, err = uuid.New().ID(); err != nil {
			logger.Error(fmt.Sprintf("failed to generate instanceID: %s", err))
			exitCode = 1
			return
		}
	}

	usersDBConfig := pgclient.Config{Name: "users"}
	if err := env.ParseWithOptions(&usersDBConfig, env.Options{Prefix: envPrefixUsersDB}); err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	usersDB, err := pgclient.Connect(usersDBConfig)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer usersDB.Close()

	domainsDBConfig := pgclient.Config{Name: "domains"}
	if err := env.ParseWithOptions(&domainsDBConfig, env.Options{Prefix: envPrefixDomainsDB}); err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	domainsDB, err := pgclient.Connect(domainsDBConfig)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer domainsDB.Close()

	tp, err := jaegerclient.NewProvider(ctx, svcName, cfg.JaegerURL, cfg.InstanceID, cfg.TraceRatio)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to init Jaeger: %s", err))
		exitCode = 1
		return
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Error(fmt.Sprintf("error shutting down tracer provider: %s", err))
		}
	}()
	tracer := tp.Tracer(svcName)

	emailConfig := email.Config{}
	if err := env.Parse(&emailConfig); err != nil {
		logger.Error(fmt.Sprintf("failed to load email configuration: %s", err))
		exitCode = 1
		return
	}

	emailConfig.Template = cfg.InvitationSentEmailTemplate

	emailerConfig := emailer.Config{
		Footer:        cfg.EmailFooter,
		SubjectPrefix: cfg.EmailSubjectPrefix,
	}
	emailNotifier, err := emailer.New(&emailConfig, emailerConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create email notifier: %s", err))
		exitCode = 1
		return
	}

	notifiers := []notifications.Notifier{emailNotifier}

	svcConfig := notifications.Config{
		InvitationSentTemplate:     cfg.InvitationSentEmailTemplate,
		InvitationAcceptedTemplate: cfg.InvitationAcceptedEmailTemplate,
		DefaultSubjectPrefix:       cfg.EmailSubjectPrefix,
		DefaultFooter:              cfg.EmailFooter,
	}
	svc := newService(notifiers, svcConfig, logger)

	subscriber, err := store.NewSubscriber(ctx, cfg.ESURL, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to create subscriber: %s", err))
		exitCode = 1
		return
	}

	logger.Info("Subscribed to Event Store")

	usersDatabase := pgclient.NewDatabase(usersDB, usersDBConfig, tracer)
	domainsDatabase := pgclient.NewDatabase(domainsDB, domainsDBConfig, tracer)

	userRepo := notificationspg.NewUserRepository(usersDatabase)
	domainRepo := notificationspg.NewDomainRepository(domainsDatabase)
	invitationRepo := notificationspg.NewInvitationRepository(domainsDatabase)

	if err := events.Start(ctx, svcName, subscriber, svc, userRepo, domainRepo, invitationRepo, logger); err != nil {
		logger.Error(fmt.Sprintf("failed to start %s event consumer: %s", svcName, err))
		exitCode = 1
		return
	}

	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err.Error()))
		exitCode = 1
		return
	}

	hs := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, httpapi.MakeHandler(svcName, cfg.InstanceID), logger)

	if cfg.SendTelemetry {
		chc := chclient.New(svcName, supermq.Version, logger, cancel)
		go chc.CallHome(ctx)
	}

	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}
}

func newService(notifiers []notifications.Notifier, config notifications.Config, logger *slog.Logger) notifications.Service {
	return notifications.NewService(notifiers, config, logger)
}

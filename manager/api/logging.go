package api

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/mainflux/mainflux/manager"
)

var _ manager.Service = (*loggingService)(nil)

type loggingService struct {
	logger log.Logger
	svc    manager.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc manager.Service, logger log.Logger) manager.Service {
	return &loggingService{logger, svc}
}

func (ls *loggingService) Register(user manager.User) (err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "register",
			"email", user.Email,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.Register(user)
}

func (ls *loggingService) Login(user manager.User) (token string, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "login",
			"email", user.Email,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.Login(user)
}

func (ls *loggingService) AddClient(key string, client manager.Client) (id string, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "add_client",
			"key", key,
			"id", id,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.AddClient(key, client)
}

func (ls *loggingService) UpdateClient(key string, client manager.Client) (err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "update_client",
			"key", key,
			"id", client.ID,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.UpdateClient(key, client)
}

func (ls *loggingService) ViewClient(key string, id string) (client manager.Client, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "view_client",
			"key", key,
			"id", id,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.ViewClient(key, id)
}

func (ls *loggingService) ListClients(key string) (clients []manager.Client, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "list_clients",
			"key", key,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.ListClients(key)
}

func (ls *loggingService) RemoveClient(key string, id string) (err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "remove_client",
			"key", key,
			"id", id,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.RemoveClient(key, id)
}

func (ls *loggingService) CreateChannel(key string, channel manager.Channel) (id string, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "create_channel",
			"key", key,
			"id", id,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.CreateChannel(key, channel)
}

func (ls *loggingService) UpdateChannel(key string, channel manager.Channel) (err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "update_channel",
			"key", key,
			"id", channel.ID,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.UpdateChannel(key, channel)
}

func (ls *loggingService) ViewChannel(key string, id string) (channel manager.Channel, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "view_channel",
			"key", key,
			"id", id,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.ViewChannel(key, id)
}

func (ls *loggingService) ListChannels(key string) (channels []manager.Channel, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "list_channels",
			"key", key,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.ListChannels(key)
}

func (ls *loggingService) RemoveChannel(key string, id string) (err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "remove_channel",
			"key", key,
			"id", id,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.RemoveChannel(key, id)
}

func (ls *loggingService) Identity(key string) (id string, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "identity",
			"id", id,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.Identity(key)
}

func (ls *loggingService) CanAccess(key string, id string) (pub string, err error) {
	defer func(begin time.Time) {
		ls.logger.Log(
			"method", "can_access",
			"key", key,
			"id", id,
			"publisher", pub,
			"error", err,
			"took", time.Since(begin),
		)
	}(time.Now())

	return ls.svc.CanAccess(key, id)
}

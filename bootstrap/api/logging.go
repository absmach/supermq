package http

import (
	"fmt"
	"nov/bootstrap"
	"time"

	log "github.com/mainflux/mainflux/logger"
)

var _ bootstrap.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    bootstrap.Service
}

// NewLoggingMiddleware adds logging facilities to the core service.
func NewLoggingMiddleware(svc bootstrap.Service, logger log.Logger) bootstrap.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Add(key string, thing bootstrap.Thing) (saved bootstrap.Thing, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method add for key %s and thing %s took %s to complete", key, saved.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Add(key, thing)
}

func (lm *loggingMiddleware) View(id, key string) (saved bootstrap.Thing, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view for key %s and thing %s took %s to complete", key, saved.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.View(id, key)
}

func (lm *loggingMiddleware) Remove(id, key string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method remove for key %s and thing %s took %s to complete", key, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Remove(id, key)
}

func (lm *loggingMiddleware) Bootstrap(externalID string) (cfg bootstrap.Config, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method bootstrap for thing with external id %s took %s to complete", externalID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Bootstrap(externalID)
}

func (lm *loggingMiddleware) ChangeStatus(id, key string, status bootstrap.Status) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method changeStatus for key %s and thing %s took %s to complete", key, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ChangeStatus(id, key, status)
}

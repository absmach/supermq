package api

import (
	"context"
	"fmt"
	"time"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/things/policies"
)

var _ policies.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    policies.Service
}

func LoggingMiddleware(svc policies.Service, logger log.Logger) policies.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) AddPolicy(ctx context.Context, token string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method add_policy for client %s and token %s took %s to complete", p.Subject, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.AddPolicy(ctx, token, p)
}

func (lm *loggingMiddleware) DeletePolicy(ctx context.Context, token string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method delete_policy for client %s in object %s and token %s took %s to complete", p.Subject, p.Object, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.DeletePolicy(ctx, token, p)
}

func (lm *loggingMiddleware) CanAccessByKey(ctx context.Context, chanID, key string) (id string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method access_by_id for channel %s in key %s took %s to complete", chanID, key, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.CanAccessByKey(ctx, chanID, key)
}

func (lm *loggingMiddleware) CanAccessByID(ctx context.Context, chanID, thingID string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method access_by_id for channel %s in thing %s took %s to complete", chanID, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.CanAccessByID(ctx, chanID, thingID)
}

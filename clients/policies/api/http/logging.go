package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/clients/policies"
	log "github.com/mainflux/mainflux/logger"
)

var _ policies.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    policies.Service
}

func LoggingMiddleware(svc policies.Service, logger log.Logger) policies.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Authorize(ctx context.Context, domain string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method authorize for client %s took %s to complete", p.Subject, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.Authorize(ctx, domain, p)
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

func (lm *loggingMiddleware) UpdatePolicy(ctx context.Context, token string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_policy for client %s and token %s took %s to complete", p.Subject, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdatePolicy(ctx, token, p)
}

func (lm *loggingMiddleware) ListPolicy(ctx context.Context, token string, cp policies.Page) (cg policies.PolicyPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_policy for client %s and token %s took %s to complete", cp.Subject, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListPolicy(ctx, token, cp)
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

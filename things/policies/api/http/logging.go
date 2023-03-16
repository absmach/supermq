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

func (lm *loggingMiddleware) Authorize(ctx context.Context, entityType string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method authorize for channel with id %s by client with id %s took %s to complete", p.Object, p.Subject, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.Authorize(ctx, entityType, p)
}

func (lm *loggingMiddleware) AuthorizeByKey(ctx context.Context, entityType string, p policies.Policy) (id string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method authorize_by_key for channel with id %s by client with id %s took %s to complete", p.Object, p.Subject, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.AuthorizeByKey(ctx, entityType, p)
}

func (lm *loggingMiddleware) AddPolicy(ctx context.Context, token string, p policies.Policy) (policy policies.Policy, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method add_policy for client with id %s using token %s took %s to complete", p.Subject, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.AddPolicy(ctx, token, p)
}

func (lm *loggingMiddleware) UpdatePolicy(ctx context.Context, token string, p policies.Policy) (policy policies.Policy, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_policy for client with id %s using token %s took %s to complete", p.Subject, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.UpdatePolicy(ctx, token, p)
}

func (lm *loggingMiddleware) ListPolicies(ctx context.Context, token string, p policies.Page) (policypage policies.PolicyPage, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method add_policy for client with id %s using token %s took %s to complete", p.Subject, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.ListPolicies(ctx, token, p)
}

func (lm *loggingMiddleware) DeletePolicy(ctx context.Context, token string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method delete_policy for client with id %s using token %s took %s to complete", p.Subject, token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())
	return lm.svc.DeletePolicy(ctx, token, p)
}

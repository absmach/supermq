// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// +build !test

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/auth"
	log "github.com/mainflux/mainflux/logger"
)

var _ auth.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    auth.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc auth.Service, logger log.Logger) auth.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Issue(ctx context.Context, issuer string, newKey auth.Key) (key auth.Key, err error) {
	defer func(begin time.Time) {
		d := "infinite duration"
		if key.ExpiresAt != nil {
			d = fmt.Sprintf("the key with expiration date %v", key.ExpiresAt)
		}
		message := fmt.Sprintf("Method issue for %s took %s to complete", d, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Issue(ctx, issuer, newKey)
}

func (lm *loggingMiddleware) Revoke(ctx context.Context, owner, id string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method revoke for key %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Revoke(ctx, owner, id)
}

func (lm *loggingMiddleware) Retrieve(ctx context.Context, owner, id string) (key auth.Key, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method retrieve for key %s took %s to complete", id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Retrieve(ctx, owner, id)
}

func (lm *loggingMiddleware) Identify(ctx context.Context, key string) (id string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method identify took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Identify(ctx, key)
}

//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !test

package api

import (
	"context"
	"fmt"
	"time"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/twins"
)

var _ twins.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    twins.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc twins.Service, logger log.Logger) twins.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) AddTwin(ctx context.Context, token string, twin twins.Twin, def twins.Definition) (saved twins.Twin, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method add_twin for for token %s and twin %s took %s to complete", token, twin.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.AddTwin(ctx, token, twin, def)
}

func (lm *loggingMiddleware) UpdateTwin(ctx context.Context, token string, twin twins.Twin, def twins.Definition) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_twin for for token %s and twin %s took %s to complete", token, twin.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateTwin(ctx, token, twin, def)
}

func (lm *loggingMiddleware) ViewTwin(ctx context.Context, token, id string) (viewed twins.Twin, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_twin for for token %s and twin %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewTwin(ctx, token, id)
}

func (lm *loggingMiddleware) ListTwins(ctx context.Context, token string, offset uint64, limit uint64, name string, metadata twins.Metadata) (tw twins.Page, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_twins for for token %s took %s to complete", token, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListTwins(ctx, token, offset, limit, name, metadata)
}

func (lm *loggingMiddleware) ListTwinsByThing(ctx context.Context, token, thing string, offset uint64, limit uint64) (tw twins.Page, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_twins_by_thing for for token %s and thing %s took %s to complete", token, thing, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListTwinsByThing(ctx, token, thing, offset, limit)
}

func (lm *loggingMiddleware) RemoveTwin(ctx context.Context, token, id string) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method remove_twin for for token %s and twin %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveTwin(ctx, token, id)
}

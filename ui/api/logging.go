// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// +build !test

package api

import (
	"context"
	"fmt"
	"time"

	log "github.com/mainflux/mainflux/logger"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/ui"
)

var _ ui.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    ui.Service
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(svc ui.Service, logger log.Logger) ui.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Index(ctx context.Context, token string) (b []byte, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method index took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Index(ctx, token)
}

func (lm *loggingMiddleware) CreateThings(ctx context.Context, token string, things ...sdk.Thing) (b []byte, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_things took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateThings(ctx, token, things...)
}

func (lm *loggingMiddleware) ViewThing(ctx context.Context, token, id string) (b []byte, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_thing for token %s and thing %s took %s to complete", token, id, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewThing(ctx, token, id)
}

func (lm *loggingMiddleware) UpdateThing(ctx context.Context, token string, thing sdk.Thing) (b []byte, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method update_thing for token %s and thing %s took %s to complete", token, thing.ID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateThing(ctx, token, thing)
}

func (lm *loggingMiddleware) ListThings(ctx context.Context, token string) (b []byte, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_things took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListThings(ctx, token)
}

func (lm *loggingMiddleware) CreateChannels(ctx context.Context, token string, channels ...sdk.Channel) (b []byte, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_channels took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateChannels(ctx, token, channels...)
}

func (lm *loggingMiddleware) ListChannels(ctx context.Context, token string) (b []byte, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_channels took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListChannels(ctx, token)
}

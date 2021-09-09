// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// +build !test

package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux/commands"
	log "github.com/mainflux/mainflux/logger"
)

var _ commands.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    commands.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc commands.Service, logger log.Logger) commands.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) CreateCommand(secret string) (response string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method create_command for secret %s took %s to complete", secret, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.CreateCommand(secret)
}

func (lm *loggingMiddleware) ViewCommand(secret string) (response string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method ViewCommand for secret %s took %s to complete", secret, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewCommand(secret)
}

func (lm *loggingMiddleware) ListCommand(secret string) (response string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method ListCommand for secret %s took %s to complete", secret, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListCommand(secret)
}

func (lm *loggingMiddleware) UpdateCommand(secret string) (response string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method UpdateCommand for secret %s took %s to complete", secret, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.UpdateCommand(secret)
}

func (lm *loggingMiddleware) RemoveCommand(secret string) (response string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method RemoveCommand for secret %s took %s to complete", secret, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RemoveCommand(secret)
}

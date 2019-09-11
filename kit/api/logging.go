//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !test

package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux/kit"
	log "github.com/mainflux/mainflux/logger"
)

var _ kit.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    kit.Service
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc kit.Service, logger log.Logger) kit.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Ping(secret string) (response string, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method ping for secret %s took %s to complete", secret, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Ping(secret)
}

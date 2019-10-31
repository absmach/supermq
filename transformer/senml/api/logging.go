// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/transformer"
)

var _ transformer.Transformer = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger logger.Logger
	svc    transformer.Transformer
}

// LoggingMiddleware adds logging facilities to the core service.
func LoggingMiddleware(svc transformer.Transformer, logger logger.Logger) transformer.Transformer {
	return &loggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}

func (lm loggingMiddleware) Transform(msg mainflux.Message) (msgs interface{}, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method transform took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Transform(msg)
}

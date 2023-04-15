// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux/consumers"
	log "github.com/mainflux/mainflux/logger"
)

var _ consumers.SyncConsumer = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger       log.Logger
	syncConsumer consumers.SyncConsumer
}

// SyncLoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(syncConsumer consumers.SyncConsumer, logger log.Logger) consumers.SyncConsumer {
	return &loggingMiddleware{
		logger:       logger,
		syncConsumer: syncConsumer,
	}
}

func (slm *loggingMiddleware) ConsumeBlocking(msgs interface{}) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method consume took %s to complete", time.Since(begin))
		if err != nil {
			slm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		slm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return slm.syncConsumer.ConsumeBlocking(msgs)
}

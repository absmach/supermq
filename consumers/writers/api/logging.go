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

// var _ consumers.AsyncConsumer = (*asyncLoggingMiddleware)(nil)
var _ consumers.SyncConsumer = (*syncLoggingMiddleware)(nil)

// type asyncLoggingMiddleware struct {
// 	logger        log.Logger
// 	asyncConsumer consumers.AsyncConsumer
// }

// // AsyncLoggingMiddleware adds logging facilities to the adapter.
// func AsyncLoggingMiddleware(asyncConsumer consumers.AsyncConsumer, logger log.Logger) consumers.AsyncConsumer {
// 	return &asyncLoggingMiddleware{
// 		logger:        logger,
// 		asyncConsumer: asyncConsumer,
// 	}
// }

// func (alm *asyncLoggingMiddleware) ConsumeAsync(msgs interface{}) {
// 	ch := alm.asyncConsumer.Errors()
// 	var err error

// 	defer func(begin time.Time) {
// 		message := fmt.Sprintf("Method ConsumeAsync took %s to complete", time.Since(begin))
// 		if err != nil {
// 			alm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
// 			return
// 		}
// 		alm.logger.Info(fmt.Sprintf("%s without errors.", message))
// 	}(time.Now())

// 	go alm.asyncConsumer.ConsumeAsync(msgs)
// 	err = <-ch
// }

// func (alm *asyncLoggingMiddleware) Errors() <-chan error {
// 	return alm.asyncConsumer.Errors()
// }

type syncLoggingMiddleware struct {
	logger       log.Logger
	syncConsumer consumers.SyncConsumer
}

// SyncLoggingMiddleware adds logging facilities to the adapter.
func SyncLoggingMiddleware(syncConsumer consumers.SyncConsumer, logger log.Logger) consumers.SyncConsumer {
	return &syncLoggingMiddleware{
		logger:       logger,
		syncConsumer: syncConsumer,
	}
}

func (slm *syncLoggingMiddleware) ConsumeBlocking(msgs interface{}) (err error) {
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

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/consumers"
)

// var _ consumers.AsyncConsumer = (*asyncMetricsMiddleware)(nil)
var _ consumers.SyncConsumer = (*syncMetricsMiddleware)(nil)

// type asyncMetricsMiddleware struct {
// 	counter       metrics.Counter
// 	latency       metrics.Histogram
// 	asyncConsumer consumers.AsyncConsumer
// }

// // AsyncMetricsMiddleware returns new message repository
// // with Save method wrapped to expose metrics.
// func AsyncMetricsMiddleware(asyncConsumer consumers.AsyncConsumer, counter metrics.Counter, latency metrics.Histogram) consumers.AsyncConsumer {
// 	return &asyncMetricsMiddleware{
// 		counter:       counter,
// 		latency:       latency,
// 		asyncConsumer: asyncConsumer,
// 	}
// }

// func (amm *asyncMetricsMiddleware) ConsumeAsync(msgs interface{}) {
// 	ch := amm.asyncConsumer.Errors()

// 	defer func(begin time.Time) {
// 		amm.counter.With("method", "consume").Add(1)
// 		amm.latency.With("method", "consume").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	go amm.asyncConsumer.ConsumeAsync(msgs)
// 	<-ch
// }

// func (amm *asyncMetricsMiddleware) Errors() <-chan error {
// 	return amm.asyncConsumer.Errors()
// }

type syncMetricsMiddleware struct {
	counter      metrics.Counter
	latency      metrics.Histogram
	syncConsumer consumers.SyncConsumer
}

// SyncMetricsMiddleware returns new message repository
// with Save method wrapped to expose metrics.
func SyncMetricsMiddleware(syncConsumer consumers.SyncConsumer, counter metrics.Counter, latency metrics.Histogram) consumers.SyncConsumer {
	return &syncMetricsMiddleware{
		counter:      counter,
		latency:      latency,
		syncConsumer: syncConsumer,
	}
}

func (smm *syncMetricsMiddleware) ConsumeBlocking(msgs interface{}) error {
	defer func(begin time.Time) {
		smm.counter.With("method", "consume").Add(1)
		smm.latency.With("method", "consume").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return smm.syncConsumer.ConsumeBlocking(msgs)
}

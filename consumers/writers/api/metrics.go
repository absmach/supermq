// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/consumers"
)

var _ consumers.SyncConsumer = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter      metrics.Counter
	latency      metrics.Histogram
	syncConsumer consumers.SyncConsumer
}

// SyncMetricsMiddleware returns new message repository
// with Save method wrapped to expose metrics.
func MetricsMiddleware(syncConsumer consumers.SyncConsumer, counter metrics.Counter, latency metrics.Histogram) consumers.SyncConsumer {
	return &metricsMiddleware{
		counter:      counter,
		latency:      latency,
		syncConsumer: syncConsumer,
	}
}

func (smm *metricsMiddleware) ConsumeBlocking(msgs interface{}) error {
	defer func(begin time.Time) {
		smm.counter.With("method", "consume").Add(1)
		smm.latency.With("method", "consume").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return smm.syncConsumer.ConsumeBlocking(msgs)
}

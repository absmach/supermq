// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/http"
	"github.com/go-kit/kit/metrics"
)

var _ http.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     http.Service
}

// Metrics instruments http adapter by tracking request count and latency.
func Metrics(svc http.Service, counter metrics.Counter, latency metrics.Histogram) http.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

// Subscribe instruments Subscribe method with metrics.
func (mm *metricsMiddleware) Subscribe(ctx context.Context, sessionID, clientKey, domainID, chanID, subtopic string, c *http.Client) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "subscribe").Add(1)
		mm.latency.With("method", "subscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Subscribe(ctx, sessionID, clientKey, domainID, chanID, subtopic, c)
}

// Unsubscribe instruments Unsubscribe method with metrics.
func (mm *metricsMiddleware) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "unsubscribe").Add(1)
		mm.latency.With("method", "unsubscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Unsubscribe(ctx, sessionID, domainID, chanID, subtopic)
}

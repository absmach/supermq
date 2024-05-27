// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/magistrala/activitylog"
	"github.com/go-kit/kit/metrics"
)

var _ activitylog.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	service activitylog.Service
}

// MetricsMiddleware returns new message repository
// with Save method wrapped to expose metrics.
func MetricsMiddleware(service activitylog.Service, counter metrics.Counter, latency metrics.Histogram) activitylog.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		service: service,
	}
}

func (mm *metricsMiddleware) Save(ctx context.Context, activity activitylog.Activity) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "save").Add(1)
		mm.latency.With("method", "save").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.service.Save(ctx, activity)
}

func (mm *metricsMiddleware) RetrieveAll(ctx context.Context, token string, page activitylog.Page) (activitylog.ActivitiesPage, error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "retrieve_all").Add(1)
		mm.latency.With("method", "retrieve_all").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.service.RetrieveAll(ctx, token, page)
}

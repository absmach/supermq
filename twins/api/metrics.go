//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !test

package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/twins"
)

var _ twins.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     twins.Service
}

// MetricsMiddleware instruments core service by tracking request count and
// latency.
func MetricsMiddleware(svc twins.Service, counter metrics.Counter, latency metrics.Histogram) twins.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) Ping(secret string) (response string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "ping").Add(1)
		ms.latency.With("method", "ping").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Ping(secret)
}

func (ms *metricsMiddleware) AddTwin(ctx context.Context, token string, twin twins.Twin) (saved twins.Twin, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "add_twin").Add(1)
		ms.latency.With("method", "add_twin").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.AddTwin(ctx, token, twin)
}

func (ms *metricsMiddleware) UpdateTwin(ctx context.Context, token string, twin twins.Twin) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_twin").Add(1)
		ms.latency.With("method", "update_twin").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateTwin(ctx, token, twin)
}

func (ms *metricsMiddleware) UpdateKey(ctx context.Context, token, id, key string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_key").Add(1)
		ms.latency.With("method", "update_key").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateKey(ctx, token, id, key)
}

func (ms *metricsMiddleware) ViewTwin(ctx context.Context, token, id string) (viewed twins.Twin, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_twin").Add(1)
		ms.latency.With("method", "view_twin").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewTwin(ctx, token, id)
}

func (ms *metricsMiddleware) ListTwins(ctx context.Context, token string, limit uint64, name string, metadata twins.Metadata) (tw twins.TwinsSet, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_twins").Add(1)
		ms.latency.With("method", "list_twins").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListTwins(ctx, token, limit, name, metadata)
}

func (ms *metricsMiddleware) ListTwinsByChannel(ctx context.Context, token, channel string, limit uint64) (tw twins.TwinsSet, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_twins_by_channel").Add(1)
		ms.latency.With("method", "list_twins_by_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListTwinsByChannel(ctx, token, channel, limit)
}

func (ms *metricsMiddleware) RemoveTwin(ctx context.Context, token, id string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "remove_twin").Add(1)
		ms.latency.With("method", "remove_twin").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.RemoveTwin(ctx, token, id)
}

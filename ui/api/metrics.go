// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// +build !test

package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/ui"
)

var _ ui.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     ui.Service
}

// MetricsMiddleware instruments adapter by tracking request count and latency.
func MetricsMiddleware(svc ui.Service, counter metrics.Counter, latency metrics.Histogram) ui.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) Index(ctx context.Context, token string) (b []byte, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "index").Add(1)
		mm.latency.With("method", "index").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Index(ctx, token)
}

func (mm *metricsMiddleware) CreateThings(ctx context.Context, token string, things ...sdk.Thing) (b []byte, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "create_things").Add(1)
		mm.latency.With("method", "create_things").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.CreateThings(ctx, token, things...)
}

func (ms *metricsMiddleware) ViewThing(ctx context.Context, token, id string) (b []byte, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_thing").Add(1)
		ms.latency.With("method", "view_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewThing(ctx, token, id)
}

func (ms *metricsMiddleware) UpdateThing(ctx context.Context, token string, thing sdk.Thing) (b []byte, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_thing").Add(1)
		ms.latency.With("method", "update_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateThing(ctx, token, thing)
}

func (mm *metricsMiddleware) ListThings(ctx context.Context, token string) (b []byte, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "list_things").Add(1)
		mm.latency.With("method", "list_things").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.ListThings(ctx, token)
}

func (ms *metricsMiddleware) RemoveThing(ctx context.Context, token, id string) (b []byte, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "remove_thing").Add(1)
		ms.latency.With("method", "remove_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.RemoveThing(ctx, token, id)
}

func (mm *metricsMiddleware) CreateChannels(ctx context.Context, token string, channels ...sdk.Channel) (b []byte, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "create_channels").Add(1)
		mm.latency.With("method", "create_channels").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.CreateChannels(ctx, token, channels...)
}

func (ms *metricsMiddleware) ViewChannel(ctx context.Context, token, id string) (b []byte, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_channel").Add(1)
		ms.latency.With("method", "view_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewChannel(ctx, token, id)
}

func (ms *metricsMiddleware) UpdateChannel(ctx context.Context, token, id string, channel sdk.Channel) (b []byte, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_channel").Add(1)
		ms.latency.With("method", "update_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateChannel(ctx, token, id, channel)
}

func (mm *metricsMiddleware) ListChannels(ctx context.Context, token string) (b []byte, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "list_channels").Add(1)
		mm.latency.With("method", "list_channels").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.ListChannels(ctx, token)
}

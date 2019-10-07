//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/opc"
)

var _ opc.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     opc.Service
}

// MetricsMiddleware instruments core service by tracking request count and latency.
func MetricsMiddleware(svc opc.Service, counter metrics.Counter, latency metrics.Histogram) opc.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) CreateThing(mfxDevID string, opcID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "create_thing").Add(1)
		mm.latency.With("method", "create_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.CreateThing(mfxDevID, opcID)
}

func (mm *metricsMiddleware) UpdateThing(mfxDevID string, opcID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "update_thing").Add(1)
		mm.latency.With("method", "update_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.UpdateThing(mfxDevID, opcID)
}

func (mm *metricsMiddleware) RemoveThing(mfxDevID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "remove_thing").Add(1)
		mm.latency.With("method", "remove_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.RemoveThing(mfxDevID)
}

func (mm *metricsMiddleware) CreateChannel(mfxChanID string, opcNamespace string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "create_channel").Add(1)
		mm.latency.With("method", "create_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.CreateChannel(mfxChanID, opcNamespace)
}

func (mm *metricsMiddleware) UpdateChannel(mfxChanID string, opcNamespace string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "update_channel").Add(1)
		mm.latency.With("method", "update_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.UpdateChannel(mfxChanID, opcNamespace)
}

func (mm *metricsMiddleware) RemoveChannel(mfxChanID string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "remove_channel").Add(1)
		mm.latency.With("method", "remove_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.RemoveChannel(mfxChanID)
}

func (mm *metricsMiddleware) Publish(ctx context.Context, token string, m opc.Message) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "message_router").Add(1)
		mm.latency.With("method", "message_router").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Publish(ctx, token, m)
}

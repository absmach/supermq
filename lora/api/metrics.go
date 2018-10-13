//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/lora"
	nats "github.com/nats-io/go-nats"
)

var _ lora.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     lora.Service
}

// MetricsMiddleware instruments core service by tracking request count and
// latency.
func MetricsMiddleware(svc lora.Service, counter metrics.Counter, latency metrics.Histogram) lora.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) ProvisionRouter(es lora.EventSourcing) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "mfx_subscribe").Add(1)
		mm.latency.With("method", "mfx_subscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.ProvisionRouter(es)
}

func (mm *metricsMiddleware) MessageRouter(m lora.Message, nc *nats.Conn) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "message_router").Add(1)
		mm.latency.With("method", "message_router").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.MessageRouter(m, nc)
}

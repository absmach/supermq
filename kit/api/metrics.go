//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !test

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/kit"
)

var _ kit.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     kit.Service
}

// MetricsMiddleware instruments core service by tracking request count and
// latency.
func MetricsMiddleware(svc kit.Service, counter metrics.Counter, latency metrics.Histogram) kit.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) Ping(secret string) (response string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "add_thing").Add(1)
		ms.latency.With("method", "add_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Ping(secret)
}

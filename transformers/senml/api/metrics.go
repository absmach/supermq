// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/transformers"
)

var _ transformers.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     transformers.Service
}

// MetricsMiddleware instruments core service by tracking request count and
// latency.
func MetricsMiddleware(svc transformers.Service, counter metrics.Counter, latency metrics.Histogram) transformers.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) Transform(msg mainflux.Message) (interface{}, error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "transform").Add(1)
		mm.latency.With("method", "transform").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Transform(msg)
}

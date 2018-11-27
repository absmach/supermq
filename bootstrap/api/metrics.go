package http

import (
	"nov/bootstrap"
	"time"

	"github.com/go-kit/kit/metrics"
)

var _ bootstrap.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     bootstrap.Service
}

// MetricsMiddleware instruments core service by tracking request count and
// latency.
func MetricsMiddleware(svc bootstrap.Service, counter metrics.Counter, latency metrics.Histogram) bootstrap.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) Add(key string, thing bootstrap.Thing) (saved bootstrap.Thing, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "add").Add(1)
		mm.latency.With("method", "add").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Add(key, thing)
}

func (mm *metricsMiddleware) View(id, key string) (saved bootstrap.Thing, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "view").Add(1)
		mm.latency.With("method", "view").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.View(id, key)
}

func (mm *metricsMiddleware) Remove(id, key string) (err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "remove").Add(1)
		mm.latency.With("method", "remove").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Remove(id, key)
}

func (mm *metricsMiddleware) Bootstrap(externalID string) (cfg bootstrap.Config, err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "bootstrap").Add(1)
		mm.latency.With("method", "bootstrap").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Bootstrap(externalID)
}

func (mm *metricsMiddleware) ChangeStatus(id, key string, status bootstrap.Status) (err error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "change_status").Add(1)
		mm.latency.With("method", "change_status").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.ChangeStatus(id, key, status)
}

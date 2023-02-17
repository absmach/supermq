package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/things/policies"
)

var _ policies.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     policies.Service
}

// MetricsMiddleware returns a new metrics middleware wrapper.
func MetricsMiddleware(svc policies.Service, counter metrics.Counter, latency metrics.Histogram) policies.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) AddPolicy(ctx context.Context, token string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "add_policy").Add(1)
		ms.latency.With("method", "add_policy").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.AddPolicy(ctx, token, p)
}

func (ms *metricsMiddleware) DeletePolicy(ctx context.Context, token string, p policies.Policy) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "delete_policy").Add(1)
		ms.latency.With("method", "delete_policy").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.DeletePolicy(ctx, token, p)
}

func (ms *metricsMiddleware) CanAccessByKey(ctx context.Context, chanID, key string) (id string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "access_by_key").Add(1)
		ms.latency.With("method", "access_by_key").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CanAccessByKey(ctx, chanID, key)
}

func (ms *metricsMiddleware) CanAccessByID(ctx context.Context, chanID, thingID string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "access_by_id").Add(1)
		ms.latency.With("method", "access_by_id").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CanAccessByID(ctx, chanID, thingID)
}

package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/normalizer"
	nats "github.com/nats-io/go-nats"
)

var _ normalizer.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     normalizer.Service
}

// MetricsMiddleware instruments adapter by tracking request count and latency.
func MetricsMiddleware(svc normalizer.Service, counter metrics.Counter, latency metrics.Histogram) normalizer.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) HandleMessage(msg *nats.Msg) {
	defer func(begin time.Time) {
		mm.counter.With("method", "handleMessage").Add(1)
		mm.latency.With("method", "handleMessage").Observe(time.Since(begin).Seconds())
	}(time.Now())
	mm.svc.HandleMessage(msg)
}

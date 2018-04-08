package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux"
)

var _ mainflux.MessagePubSub = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     mainflux.MessagePubSub
}

// MetricsMiddleware instruments adapter by tracking request count and latency.
func MetricsMiddleware(svc mainflux.MessagePubSub, counter metrics.Counter, latency metrics.Histogram) mainflux.MessagePubSub {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) Publish(msg mainflux.RawMessage, cfHandler mainflux.ConnFailHandler) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "publish").Add(1)
		mm.latency.With("method", "publish").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Publish(msg, cfHandler)
}

func (mm *metricsMiddleware) Subscribe(sub mainflux.Subscription, cfHandler mainflux.ConnFailHandler) (mainflux.Unsubscribe, error) {
	defer func(begin time.Time) {
		mm.counter.With("method", "subscribe").Add(1)
		mm.latency.With("method", "subscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Subscribe(sub, cfHandler)
}

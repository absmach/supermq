package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
)

var _ ws.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     ws.Service
}

// MetricsMiddleware instruments adapter by tracking request count and latency.
func MetricsMiddleware(svc ws.Service, counter metrics.Counter, latency metrics.Histogram) ws.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (mm *metricsMiddleware) Publish(msg mainflux.RawMessage) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "publish").Add(1)
		mm.latency.With("method", "publish").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Publish(msg)
}

func (mm *metricsMiddleware) Broadcast(socket ws.Socket, msg mainflux.RawMessage) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "broadcast").Add(1)
		mm.latency.With("method", "broadcast").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Broadcast(socket, msg)
}

func (mm *metricsMiddleware) Listen(socket ws.Socket, sub ws.Subscription, onClose func()) {
	defer func(begin time.Time) {
		mm.counter.With("method", "start_listening").Add(1)
		mm.latency.With("method", "start_listening").Observe(time.Since(begin).Seconds())
	}(time.Now())

	mm.svc.Listen(socket, sub, onClose)
}

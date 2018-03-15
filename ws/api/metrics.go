package api

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
	broker "github.com/nats-io/go-nats"
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

func (mm *metricsMiddleware) HandleMessage(msg *broker.Msg) {
	defer func(begin time.Time) {
		mm.counter.With("method", "handle_message").Add(1)
		mm.latency.With("method", "handle_message").Observe(time.Since(begin).Seconds())
	}(time.Now())

	mm.svc.HandleMessage(msg)
}

func (mm *metricsMiddleware) AddConnection(channelID, publisherID string, conn *websocket.Conn) {
	defer func(begin time.Time) {
		mm.counter.With("method", "add_connection").Add(1)
		mm.latency.With("method", "add_connection").Observe(time.Since(begin).Seconds())
	}(time.Now())

	mm.svc.AddConnection(channelID, publisherID, conn)
}

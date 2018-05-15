package influxdb

import (
	"time"

	"github.com/mainflux/mainflux"

	"github.com/go-kit/kit/metrics"
)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	ef      eventFlow
}

var _ Writer = (*metricsMiddleware)(nil)

func newMetricsMiddleware(ef eventFlow, counter metrics.Counter, latency metrics.Histogram) *metricsMiddleware {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		ef:      ef,
	}
}

func (mm *metricsMiddleware) Save(msg mainflux.Message) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "Save").Add(1)
		mm.latency.With("method", "Save").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mm.ef.Save(msg)
}

func (mm *metricsMiddleware) Close() error {
	return mm.ef.Close()
}

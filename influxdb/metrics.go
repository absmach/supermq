package influxdb

import (
	"time"

	"github.com/mainflux/mainflux"

	"github.com/go-kit/kit/metrics"
)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	w       Writer
}

var _ Writer = (*metricsMiddleware)(nil)

// MetricsMiddleware instruments writer by tracking request count and latency.
func MetricsMiddleware(w Writer, counter metrics.Counter, latency metrics.Histogram) Writer {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		w:       w,
	}
}

func (mm *metricsMiddleware) Save(msg mainflux.Message) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "Save").Add(1)
		mm.latency.With("method", "Save").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mm.w.Save(msg)
}

func (mm *metricsMiddleware) Close() error {
	return mm.w.Close()
}

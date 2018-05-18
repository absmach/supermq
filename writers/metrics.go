package writers

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux"
)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	repo    MessageRepository
}

func newMetricsMiddleware(repo MessageRepository, counter metrics.Counter, latency metrics.Histogram) *metricsMiddleware {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		repo:    repo,
	}
}

func (mm *metricsMiddleware) Save(msg mainflux.Message) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "handleMessage").Add(1)
		mm.latency.With("method", "handleMessage").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return mm.repo.Save(msg)
}

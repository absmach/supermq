// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package middleware

import (
	"context"
	"time"

	smqhttp "github.com/absmach/supermq/http"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/go-kit/kit/metrics"
)

var _ smqhttp.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     smqhttp.Service
}

// NewMetrics instruments adapter by tracking request count and latency.
func NewMetrics(svc smqhttp.Service, counter metrics.Counter, latency metrics.Histogram) smqhttp.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

// Subscribe instruments Subscribe method with metrics.
func (mm *metricsMiddleware) Subscribe(ctx context.Context, sessionID, username, password, domainID, chanID, subtopic string, topicType messaging.TopicType, c *smqhttp.Client) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "subscribe").Add(1)
		mm.latency.With("method", "subscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Subscribe(ctx, sessionID, username, password, domainID, chanID, subtopic, topicType, c)
}

func (mm *metricsMiddleware) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string, topicType messaging.TopicType) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "unsubscribe").Add(1)
		mm.latency.With("method", "unsubscribe").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.svc.Unsubscribe(ctx, sessionID, domainID, chanID, subtopic, topicType)
}

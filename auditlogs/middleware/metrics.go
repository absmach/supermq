// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/auditlogs"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/go-kit/kit/metrics"
)


var _ auditlogs.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     auditlogs.Service
}

// MetricsMiddleware returns a new metrics middleware wrapper.
func MetricsMiddleware(svc auditlogs.Service, counter metrics.Counter, latency metrics.Histogram) auditlogs.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) Save(ctx context.Context, log auditlogs.AuditLog) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "Save").Add(1)
		ms.latency.With("method", "Save").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Save(ctx, log)
}

func (ms *metricsMiddleware) RetrieveByID(ctx context.Context, session authn.Session, id string) (auditlogs.AuditLog, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "RetrieveByID").Add(1)
		ms.latency.With("method", "RetrieveByID").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RetrieveByID(ctx, session, id)
}

func (ms *metricsMiddleware) RetrieveAll(ctx context.Context, session authn.Session, pm auditlogs.Page) (auditlogs.AuditLogPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "RetrieveAll").Add(1)
		ms.latency.With("method", "RetrieveAll").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RetrieveAll(ctx, session, pm)
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"
	"time"

	"github.com/absmach/magistrala/pkg/domains"
	entityRolesAPI "github.com/absmach/magistrala/pkg/entityroles/api"
	"github.com/go-kit/kit/metrics"
)

var _ domains.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     domains.Service
	entityRolesAPI.RolesSvcMetricsMiddleware
}

// MetricsMiddleware instruments core service by tracking request count and latency.
func MetricsMiddleware(svc domains.Service, counter metrics.Counter, latency metrics.Histogram) domains.Service {
	rolesSvcMetricsMiddleware := entityRolesAPI.NewRolesSvcMetricsMiddleware("domains", svc, counter, latency)

	return &metricsMiddleware{
		counter:                   counter,
		latency:                   latency,
		svc:                       svc,
		RolesSvcMetricsMiddleware: rolesSvcMetricsMiddleware,
	}
}

func (ms *metricsMiddleware) CreateDomain(ctx context.Context, token string, d domains.Domain) (domains.Domain, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_domain").Add(1)
		ms.latency.With("method", "create_domain").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreateDomain(ctx, token, d)
}

func (ms *metricsMiddleware) RetrieveDomain(ctx context.Context, token, id string) (domains.Domain, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "retrieve_domain").Add(1)
		ms.latency.With("method", "retrieve_domain").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RetrieveDomain(ctx, token, id)
}

func (ms *metricsMiddleware) UpdateDomain(ctx context.Context, token, id string, d domains.DomainReq) (domains.Domain, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_domain").Add(1)
		ms.latency.With("method", "update_domain").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateDomain(ctx, token, id, d)
}

func (ms *metricsMiddleware) ChangeDomainStatus(ctx context.Context, token, id string, d domains.DomainReq) (domains.Domain, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "change_domain_status").Add(1)
		ms.latency.With("method", "change_domain_status").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ChangeDomainStatus(ctx, token, id, d)
}

func (ms *metricsMiddleware) ListDomains(ctx context.Context, token string, page domains.Page) (domains.DomainsPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_domains").Add(1)
		ms.latency.With("method", "list_domains").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListDomains(ctx, token, page)
}

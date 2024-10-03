// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"time"

	"github.com/absmach/magistrala"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	entityRolesAPI "github.com/absmach/magistrala/pkg/entityroles/api"
	"github.com/absmach/magistrala/things"
	"github.com/go-kit/kit/metrics"
)

var _ things.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     things.Service
	entityRolesAPI.RolesSvcMetricsMiddleware
}

// MetricsMiddleware returns a new metrics middleware wrapper.
func MetricsMiddleware(svc things.Service, counter metrics.Counter, latency metrics.Histogram) things.Service {
	rolesSvcMetricsMiddleware := entityRolesAPI.NewRolesSvcMetricsMiddleware("things", svc, counter, latency)
	return &metricsMiddleware{
		counter:                   counter,
		latency:                   latency,
		svc:                       svc,
		RolesSvcMetricsMiddleware: rolesSvcMetricsMiddleware,
	}
}

func (ms *metricsMiddleware) CreateThings(ctx context.Context, token string, clients ...mgclients.Client) ([]mgclients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "register_things").Add(1)
		ms.latency.With("method", "register_things").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreateThings(ctx, token, clients...)
}

func (ms *metricsMiddleware) ViewClient(ctx context.Context, token, id string) (mgclients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_thing").Add(1)
		ms.latency.With("method", "view_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ViewClient(ctx, token, id)
}

func (ms *metricsMiddleware) ListClients(ctx context.Context, token, reqUserID string, pm mgclients.Page) (mgclients.ClientsPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_things").Add(1)
		ms.latency.With("method", "list_things").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListClients(ctx, token, reqUserID, pm)
}

func (ms *metricsMiddleware) UpdateClient(ctx context.Context, token string, client mgclients.Client) (mgclients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_thing_name_and_metadata").Add(1)
		ms.latency.With("method", "update_thing_name_and_metadata").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClient(ctx, token, client)
}

func (ms *metricsMiddleware) UpdateClientTags(ctx context.Context, token string, client mgclients.Client) (mgclients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_thing_tags").Add(1)
		ms.latency.With("method", "update_thing_tags").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClientTags(ctx, token, client)
}

func (ms *metricsMiddleware) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (mgclients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_thing_secret").Add(1)
		ms.latency.With("method", "update_thing_secret").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)
}

func (ms *metricsMiddleware) EnableClient(ctx context.Context, token, id string) (mgclients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "enable_thing").Add(1)
		ms.latency.With("method", "enable_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.EnableClient(ctx, token, id)
}

func (ms *metricsMiddleware) DisableClient(ctx context.Context, token, id string) (mgclients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "disable_thing").Add(1)
		ms.latency.With("method", "disable_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.DisableClient(ctx, token, id)
}

func (ms *metricsMiddleware) Identify(ctx context.Context, key string) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "identify_thing").Add(1)
		ms.latency.With("method", "identify_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Identify(ctx, key)
}

func (ms *metricsMiddleware) Authorize(ctx context.Context, req *magistrala.AuthorizeReq) (id string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "authorize").Add(1)
		ms.latency.With("method", "authorize").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Authorize(ctx, req)
}

func (ms *metricsMiddleware) DeleteClient(ctx context.Context, token, id string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "delete_client").Add(1)
		ms.latency.With("method", "delete_client").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.DeleteClient(ctx, token, id)
}

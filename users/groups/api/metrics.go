package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/users/groups"
)

var _ groups.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     groups.Service
}

// MetricsMiddleware returns a new metrics middleware wrapper.
func MetricsMiddleware(svc groups.Service, counter metrics.Counter, latency metrics.Histogram) groups.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) CreateGroup(ctx context.Context, token string, g groups.Group) (groups.Group, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_group").Add(1)
		ms.latency.With("method", "create_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreateGroup(ctx, token, g)
}

func (ms *metricsMiddleware) UpdateGroup(ctx context.Context, token string, group groups.Group) (rGroup groups.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_group").Add(1)
		ms.latency.With("method", "update_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateGroup(ctx, token, group)
}

func (ms *metricsMiddleware) ViewGroup(ctx context.Context, token, id string) (g groups.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_group").Add(1)
		ms.latency.With("method", "view_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ViewGroup(ctx, token, id)
}

func (ms *metricsMiddleware) ListGroups(ctx context.Context, token string, gp groups.GroupsPage) (cg groups.GroupsPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_groups").Add(1)
		ms.latency.With("method", "list_groups").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListGroups(ctx, token, gp)
}

func (ms *metricsMiddleware) EnableGroup(ctx context.Context, token string, id string) (g groups.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "enable_group").Add(1)
		ms.latency.With("method", "enable_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.EnableGroup(ctx, token, id)
}

func (ms *metricsMiddleware) DisableGroup(ctx context.Context, token string, id string) (g groups.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "disable_group").Add(1)
		ms.latency.With("method", "disable_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.DisableGroup(ctx, token, id)
}

func (ms *metricsMiddleware) ListMemberships(ctx context.Context, token, clientID string, gp groups.GroupsPage) (mp groups.MembershipsPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_memberships").Add(1)
		ms.latency.With("method", "list_memberships").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListMemberships(ctx, token, clientID, gp)
}

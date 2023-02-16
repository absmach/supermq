package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/clients/clients"
)

var _ clients.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     clients.Service
}

// MetricsMiddleware returns a new metrics middleware wrapper.
func MetricsMiddleware(svc clients.Service, counter metrics.Counter, latency metrics.Histogram) clients.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) CreateThing(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "register_client").Add(1)
		ms.latency.With("method", "register_client").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreateThing(ctx, token, client)
}

func (ms *metricsMiddleware) ViewClient(ctx context.Context, token, id string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_client").Add(1)
		ms.latency.With("method", "view_client").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ViewClient(ctx, token, id)
}

func (ms *metricsMiddleware) ListClients(ctx context.Context, token string, pm clients.Page) (clients.ClientsPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_clients").Add(1)
		ms.latency.With("method", "list_clients").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListClients(ctx, token, pm)
}

func (ms *metricsMiddleware) UpdateClient(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_client_name_and_metadata").Add(1)
		ms.latency.With("method", "update_client_name_and_metadata").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClient(ctx, token, client)
}

func (ms *metricsMiddleware) UpdateClientTags(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_client_tags").Add(1)
		ms.latency.With("method", "update_client_tags").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClientTags(ctx, token, client)
}

func (ms *metricsMiddleware) ShareThing(ctx context.Context, token, id string, actions, userIDs []string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "share_thing").Add(1)
		ms.latency.With("method", "share_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ShareThing(ctx, token, id, actions, userIDs)
}

func (ms *metricsMiddleware) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_client_secret").Add(1)
		ms.latency.With("method", "update_client_secret").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)
}

func (ms *metricsMiddleware) UpdateClientOwner(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_client_owner").Add(1)
		ms.latency.With("method", "update_client_owner").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClientOwner(ctx, token, client)
}

func (ms *metricsMiddleware) EnableClient(ctx context.Context, token string, id string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "enable_client").Add(1)
		ms.latency.With("method", "enable_client").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.EnableClient(ctx, token, id)
}

func (ms *metricsMiddleware) DisableClient(ctx context.Context, token string, id string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "disable_client").Add(1)
		ms.latency.With("method", "disable_client").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.DisableClient(ctx, token, id)
}

func (ms *metricsMiddleware) ListThingsByChannel(ctx context.Context, token, groupID string, pm clients.Page) (mp clients.MembersPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_things_by_channel").Add(1)
		ms.latency.With("method", "list_things_by_channel").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListThingsByChannel(ctx, token, groupID, pm)
}

func (ms *metricsMiddleware) Identify(ctx context.Context, key string) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "identify_thing").Add(1)
		ms.latency.With("method", "identify_thing").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Identify(ctx, key)
}

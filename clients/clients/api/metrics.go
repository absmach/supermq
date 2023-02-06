package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/clients/clients"
	"github.com/mainflux/mainflux/clients/jwt"
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

func (ms *metricsMiddleware) RegisterClient(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "register_client").Add(1)
		ms.latency.With("method", "register_client").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RegisterClient(ctx, token, client)
}

func (ms *metricsMiddleware) IssueToken(ctx context.Context, identity, secret string) (jwt.Token, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "issue_token").Add(1)
		ms.latency.With("method", "issue_token").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.IssueToken(ctx, identity, secret)
}

func (ms *metricsMiddleware) RefreshToken(ctx context.Context, accessToken string) (token jwt.Token, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "refresh_token").Add(1)
		ms.latency.With("method", "refresh_token").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RefreshToken(ctx, accessToken)
}

func (ms *metricsMiddleware) ViewClient(ctx context.Context, token, id string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_client").Add(1)
		ms.latency.With("method", "view_client").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ViewClient(ctx, token, id)
}

func (ms *metricsMiddleware) ViewProfile(ctx context.Context, token string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_profile").Add(1)
		ms.latency.With("method", "view_profile").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ViewProfile(ctx, token)
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

func (ms *metricsMiddleware) UpdateClientIdentity(ctx context.Context, token, id, identity string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_client_identity").Add(1)
		ms.latency.With("method", "update_client_identity").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClientIdentity(ctx, token, id, identity)
}

func (ms *metricsMiddleware) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (clients.Client, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_client_secret").Add(1)
		ms.latency.With("method", "update_client_secret").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)
}

func (ms *metricsMiddleware) GenerateResetToken(ctx context.Context, email, host string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "generate_reset_token").Add(1)
		ms.latency.With("method", "generate_reset_token").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.GenerateResetToken(ctx, email, host)
}

func (ms *metricsMiddleware) ResetSecret(ctx context.Context, token, secret string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "reset_secret").Add(1)
		ms.latency.With("method", "reset_secret").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ResetSecret(ctx, token, secret)
}

func (ms *metricsMiddleware) SendPasswordReset(ctx context.Context, host, email, token string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "send_password_reset").Add(1)
		ms.latency.With("method", "send_password_reset").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.SendPasswordReset(ctx, email, host, token)
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

func (ms *metricsMiddleware) ListMembers(ctx context.Context, token, groupID string, pm clients.Page) (mp clients.MembersPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_members").Add(1)
		ms.latency.With("method", "list_members").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListMembers(ctx, token, groupID, pm)
}

func (ms *metricsMiddleware) Identify(ctx context.Context, token string) (clients.UserIdentity, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "identify").Add(1)
		ms.latency.With("method", "identify").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Identify(ctx, token)
}

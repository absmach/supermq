package tracing

import (
	"context"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/users/clients"
	"github.com/mainflux/mainflux/users/jwt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ clients.Service = (*tracingMiddleware)(nil)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    clients.Service
}

func TracingMiddleware(svc clients.Service, tracer trace.Tracer) clients.Service {
	return &tracingMiddleware{tracer, svc}
}

func (tm *tracingMiddleware) RegisterClient(ctx context.Context, token string, client mfclients.Client) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_register_client", trace.WithAttributes(attribute.String("identity", client.Credentials.Identity)))
	defer span.End()

	return tm.svc.RegisterClient(ctx, token, client)
}

func (tm *tracingMiddleware) IssueToken(ctx context.Context, identity, secret string) (jwt.Token, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_issue_token", trace.WithAttributes(attribute.String("identity", identity)))
	defer span.End()

	return tm.svc.IssueToken(ctx, identity, secret)
}

func (tm *tracingMiddleware) RefreshToken(ctx context.Context, accessToken string) (jwt.Token, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_refresh_token", trace.WithAttributes(attribute.String("access_token", accessToken)))
	defer span.End()

	return tm.svc.RefreshToken(ctx, accessToken)
}
func (tm *tracingMiddleware) ViewClient(ctx context.Context, token string, id string) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_view_client", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()
	return tm.svc.ViewClient(ctx, token, id)
}

func (tm *tracingMiddleware) ListClients(ctx context.Context, token string, pm mfclients.Page) (mfclients.ClientsPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_clients")
	defer span.End()
	return tm.svc.ListClients(ctx, token, pm)
}

func (tm *tracingMiddleware) UpdateClient(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_name_and_metadata", trace.WithAttributes(attribute.String("Name", cli.Name)))
	defer span.End()

	return tm.svc.UpdateClient(ctx, token, cli)
}

func (tm *tracingMiddleware) UpdateClientTags(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_tags", trace.WithAttributes(attribute.StringSlice("Tags", cli.Tags)))
	defer span.End()

	return tm.svc.UpdateClientTags(ctx, token, cli)
}
func (tm *tracingMiddleware) UpdateClientIdentity(ctx context.Context, token, id, identity string) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_identity", trace.WithAttributes(attribute.String("Identity", identity)))
	defer span.End()

	return tm.svc.UpdateClientIdentity(ctx, token, id, identity)

}

func (tm *tracingMiddleware) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_secret")
	defer span.End()

	return tm.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)

}

func (tm *tracingMiddleware) GenerateResetToken(ctx context.Context, email, host string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_generate_reset_token")
	defer span.End()

	return tm.svc.GenerateResetToken(ctx, email, host)

}

func (tm *tracingMiddleware) ResetSecret(ctx context.Context, token, secret string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_reset_secret")
	defer span.End()

	return tm.svc.ResetSecret(ctx, token, secret)

}

func (tm *tracingMiddleware) SendPasswordReset(ctx context.Context, host, email, token string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_send_password_reset")
	defer span.End()

	return tm.svc.SendPasswordReset(ctx, host, email, token)

}

func (tm *tracingMiddleware) ViewProfile(ctx context.Context, token string) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_view_profile")
	defer span.End()

	return tm.svc.ViewProfile(ctx, token)

}

func (tm *tracingMiddleware) UpdateClientOwner(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_owner", trace.WithAttributes(attribute.StringSlice("Tags", cli.Tags)))
	defer span.End()

	return tm.svc.UpdateClientOwner(ctx, token, cli)
}

func (tm *tracingMiddleware) EnableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_enable_client", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()

	return tm.svc.EnableClient(ctx, token, id)
}

func (tm *tracingMiddleware) DisableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_disable_client", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()

	return tm.svc.DisableClient(ctx, token, id)
}

func (tm *tracingMiddleware) ListMembers(ctx context.Context, token, groupID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_members")
	defer span.End()

	return tm.svc.ListMembers(ctx, token, groupID, pm)

}

func (tm *tracingMiddleware) Identify(ctx context.Context, token string) (string, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_identify", trace.WithAttributes(attribute.String("token", token)))
	defer span.End()

	return tm.svc.Identify(ctx, token)
}

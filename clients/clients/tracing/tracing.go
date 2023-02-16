package tracing

import (
	"context"

	"github.com/mainflux/mainflux/clients/clients"
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

func (tm *tracingMiddleware) CreateThing(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_create_client", trace.WithAttributes(attribute.String("identity", client.Credentials.Identity)))
	defer span.End()

	return tm.svc.CreateThing(ctx, token, client)
}

func (tm *tracingMiddleware) ViewClient(ctx context.Context, token string, id string) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_view_client", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()
	return tm.svc.ViewClient(ctx, token, id)
}

func (tm *tracingMiddleware) ListClients(ctx context.Context, token string, pm clients.Page) (clients.ClientsPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_clients")
	defer span.End()
	return tm.svc.ListClients(ctx, token, pm)
}

func (tm *tracingMiddleware) UpdateClient(ctx context.Context, token string, cli clients.Client) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_name_and_metadata", trace.WithAttributes(attribute.String("Name", cli.Name)))
	defer span.End()

	return tm.svc.UpdateClient(ctx, token, cli)
}

func (tm *tracingMiddleware) UpdateClientTags(ctx context.Context, token string, cli clients.Client) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_tags", trace.WithAttributes(attribute.StringSlice("Tags", cli.Tags)))
	defer span.End()

	return tm.svc.UpdateClientTags(ctx, token, cli)
}

func (tm *tracingMiddleware) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_secret")
	defer span.End()

	return tm.svc.UpdateClientSecret(ctx, token, oldSecret, newSecret)

}

func (tm *tracingMiddleware) UpdateClientOwner(ctx context.Context, token string, cli clients.Client) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_client_owner", trace.WithAttributes(attribute.StringSlice("Tags", cli.Tags)))
	defer span.End()

	return tm.svc.UpdateClientOwner(ctx, token, cli)
}

func (tm *tracingMiddleware) EnableClient(ctx context.Context, token, id string) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_enable_client", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()

	return tm.svc.EnableClient(ctx, token, id)
}

func (tm *tracingMiddleware) DisableClient(ctx context.Context, token, id string) (clients.Client, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_disable_client", trace.WithAttributes(attribute.String("ID", id)))
	defer span.End()

	return tm.svc.DisableClient(ctx, token, id)
}

func (tm *tracingMiddleware) ListThingsByChannel(ctx context.Context, token, groupID string, pm clients.Page) (clients.MembersPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_things_by_channel")
	defer span.End()

	return tm.svc.ListThingsByChannel(ctx, token, groupID, pm)

}

func (tm *tracingMiddleware) ShareThing(ctx context.Context, token string, thingID string, actions, userIDs []string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_view_client", trace.WithAttributes(attribute.String("ID", thingID)))
	defer span.End()
	return tm.svc.ShareThing(ctx, token, thingID, actions, userIDs)
}

func (tm *tracingMiddleware) Identify(ctx context.Context, key string) (string, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_view_client", trace.WithAttributes(attribute.String("Key", key)))
	defer span.End()
	return tm.svc.Identify(ctx, key)
}

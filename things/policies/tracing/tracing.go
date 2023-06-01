package tracing

import (
	"context"

	"github.com/mainflux/mainflux/things/policies"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ policies.Service = (*tracingMiddleware)(nil)

type tracingMiddleware struct {
	tracer trace.Tracer
	psvc   policies.Service
}

// TracingMiddleware enriches policies with traces for improved monitoring.
func TracingMiddleware(psvc policies.Service, tracer trace.Tracer) policies.Service {
	return &tracingMiddleware{tracer, psvc}
}

func (tm *tracingMiddleware) Authorize(ctx context.Context, ar policies.AccessRequest, entityType string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_authorize", trace.WithAttributes(attribute.String("subject", p.Subject), attribute.String("object", p.Object), attribute.StringSlice("actions", p.Actions)))
	defer span.End()

	return tm.psvc.Authorize(ctx, ar, entityType, p)
}

func (tm *tracingMiddleware) AuthorizeByKey(ctx context.Context, ar policies.AccessRequest, entityType string) (string, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_authorize_by_key", trace.WithAttributes(attribute.String("subject", ar.Subject), attribute.String("object", ar.Object), attribute.String("action", ar.Action)))
	defer span.End()

	return tm.psvc.AuthorizeByKey(ctx, ar, entityType)
}

func (tm *tracingMiddleware) AddPolicy(ctx context.Context, token string, p policies.Policy) (policies.Policy, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_connect", trace.WithAttributes(attribute.StringSlice("actions", p.Actions)))
	defer span.End()

	return tm.psvc.AddPolicy(ctx, token, p)
}

func (tm *tracingMiddleware) UpdatePolicy(ctx context.Context, token string, p policies.Policy) (policies.Policy, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_policy", trace.WithAttributes(attribute.StringSlice("actions", p.Actions)))
	defer span.End()

	return tm.psvc.UpdatePolicy(ctx, token, p)
}

func (tm *tracingMiddleware) ListPolicies(ctx context.Context, token string, p policies.Page) (policies.PolicyPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_policies", trace.WithAttributes(attribute.String("actions", p.Action)))
	defer span.End()

	return tm.psvc.ListPolicies(ctx, token, p)
}

func (tm *tracingMiddleware) DeletePolicy(ctx context.Context, token string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_disconnect", trace.WithAttributes(attribute.String("subject", p.Subject), attribute.String("object", p.Object)))
	defer span.End()

	return tm.psvc.DeletePolicy(ctx, token, p)
}

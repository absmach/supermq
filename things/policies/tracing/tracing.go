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

func TracingMiddleware(psvc policies.Service, tracer trace.Tracer) policies.Service {
	return &tracingMiddleware{tracer, psvc}
}

func (tm *tracingMiddleware) Authorize(ctx context.Context, entityType string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_authorize", trace.WithAttributes(attribute.String("Subject", p.Subject), attribute.String("Object", p.Object), attribute.StringSlice("Actions", p.Actions)))
	defer span.End()

	return tm.psvc.Authorize(ctx, entityType, p)
}

func (tm *tracingMiddleware) AuthorizeByKey(ctx context.Context, entityType string, p policies.Policy) (string, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_authorize_by_key", trace.WithAttributes(attribute.String("Subject", p.Subject), attribute.String("Object", p.Object), attribute.StringSlice("Actions", p.Actions)))
	defer span.End()

	return tm.psvc.AuthorizeByKey(ctx, entityType, p)
}

func (tm *tracingMiddleware) AddPolicy(ctx context.Context, token string, p policies.Policy) (policies.Policy, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_connect", trace.WithAttributes(attribute.StringSlice("Actions", p.Actions)))
	defer span.End()

	return tm.psvc.AddPolicy(ctx, token, p)
}

func (tm *tracingMiddleware) UpdatePolicy(ctx context.Context, token string, p policies.Policy) (policies.Policy, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_update_policy", trace.WithAttributes(attribute.StringSlice("Actions", p.Actions)))
	defer span.End()

	return tm.psvc.UpdatePolicy(ctx, token, p)
}

func (tm *tracingMiddleware) ListPolicies(ctx context.Context, token string, p policies.Page) (policies.PolicyPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_policies", trace.WithAttributes(attribute.String("Actions", p.Action)))
	defer span.End()

	return tm.psvc.ListPolicies(ctx, token, p)
}

func (tm *tracingMiddleware) DeletePolicy(ctx context.Context, token string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_disconnect", trace.WithAttributes(attribute.String("Subject", p.Subject), attribute.String("Object", p.Object)))
	defer span.End()

	return tm.psvc.DeletePolicy(ctx, token, p)
}

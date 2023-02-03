package tracing

import (
	"context"

	"github.com/mainflux/mainflux/clients/policies"
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

func (tm *tracingMiddleware) Authorize(ctx context.Context, domain string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_authorize", trace.WithAttributes(attribute.StringSlice("Action", p.Actions)))
	defer span.End()

	return tm.psvc.Authorize(ctx, domain, p)
}
func (tm *tracingMiddleware) UpdatePolicy(ctx context.Context, token string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_update_policy", trace.WithAttributes(attribute.StringSlice("Actions", p.Actions)))
	defer span.End()

	return tm.psvc.UpdatePolicy(ctx, token, p)

}

func (tm *tracingMiddleware) AddPolicy(ctx context.Context, token string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_add_policy", trace.WithAttributes(attribute.StringSlice("Actions", p.Actions)))
	defer span.End()

	return tm.psvc.AddPolicy(ctx, token, p)

}

func (tm *tracingMiddleware) DeletePolicy(ctx context.Context, token string, p policies.Policy) error {
	ctx, span := tm.tracer.Start(ctx, "svc_delete_policy", trace.WithAttributes(attribute.String("Subject", p.Subject), attribute.String("Object", p.Object)))
	defer span.End()

	return tm.psvc.DeletePolicy(ctx, token, p)

}

func (tm *tracingMiddleware) ListPolicy(ctx context.Context, token string, pm policies.Page) (policies.PolicyPage, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_list_policy")
	defer span.End()

	return tm.psvc.ListPolicy(ctx, token, pm)

}

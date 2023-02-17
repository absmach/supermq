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

func (tm *tracingMiddleware) CanAccessByID(ctx context.Context, chanID, thingID string) error {
	ctx, span := tm.tracer.Start(ctx, "svc_access_by_id", trace.WithAttributes(attribute.String("channelID", chanID), attribute.String("thingID", thingID)))
	defer span.End()

	return tm.psvc.CanAccessByID(ctx, chanID, thingID)

}
func (tm *tracingMiddleware) CanAccessByKey(ctx context.Context, chanID, thingKey string) (string, error) {
	ctx, span := tm.tracer.Start(ctx, "svc_access_by_key", trace.WithAttributes(attribute.String("channelID", chanID), attribute.String("Key", thingKey)))
	defer span.End()

	return tm.psvc.CanAccessByKey(ctx, chanID, thingKey)

}

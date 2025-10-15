package tracing

import (
	"context"

	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type tracingMiddleware struct {
	tracer trace.Tracer
	co     callout.Callout
}

// New returns a new callout with tracing capabilities.
func New(svc callout.Callout, tracer trace.Tracer) callout.Callout {
	return &tracingMiddleware{tracer, svc}
}

func (tm *tracingMiddleware) Callout(ctx context.Context, op string, pld map[string]any) error {
	{
		ctx, span := tracing.StartSpan(ctx, tm.tracer, "callout", trace.WithAttributes(
			attribute.String("operation", op),
		))
		defer span.End()
		return tm.co.Callout(ctx, op, pld)
	}
}

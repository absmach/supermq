package tracing

import (
	"context"

	"github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
)

var _ http.Service = (*serviceMiddleware)(nil)

const publishOP = "publishOP"

type serviceMiddleware struct {
	tracer opentracing.Tracer
	svc    http.Service
}

// New create new http tracing middleware service
func New(tracer opentracing.Tracer, svc http.Service) http.Service {
	return &serviceMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Publish trace http publish operations
func (sm *serviceMiddleware) Publish(ctx context.Context, token string, msg *messaging.Message) error {
	var spanCtx opentracing.SpanContext = nil
	if httpSpan := opentracing.SpanFromContext(ctx); httpSpan != nil {
		spanCtx = httpSpan.Context()
	}
	span := sm.tracer.StartSpan(publishOP, opentracing.ChildOf(spanCtx))
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return sm.svc.Publish(ctx, token, msg)
}

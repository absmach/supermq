package tracing

import (
	"bytes"
	"context"

	"github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
)

var _ http.Service = (*serviceMiddleware)(nil)

type serviceMiddleware struct {
	tracer opentracing.Tracer
	svc    http.Service
}

func New(tracer opentracing.Tracer, svc http.Service) *serviceMiddleware {
	return &serviceMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Publish implements http.Service
func (sm *serviceMiddleware) Publish(ctx context.Context, token string, msg *messaging.Message) error {
	var spanCtx opentracing.SpanContext = nil

	if coapSpan := opentracing.SpanFromContext(ctx); coapSpan != nil {
		spanCtx = coapSpan.Context()
	}
	span := sm.tracer.StartSpan("http publish", opentracing.ChildOf(spanCtx))
	defer span.Finish()
	dataBuffer := bytes.NewBuffer(msg.Span)

	if err := sm.tracer.Inject(span.Context(), opentracing.Binary, dataBuffer); err != nil {
		return err
	}
	msg.Span = dataBuffer.Bytes()
	return sm.svc.Publish(ctx, token, msg)
}

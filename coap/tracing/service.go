package tracing

import (
	"bytes"
	"context"

	"github.com/mainflux/mainflux/coap"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
)

var _ coap.Service = (*tracingServiceMiddleware)(nil)

type tracingServiceMiddleware struct {
	tracer opentracing.Tracer
	svc    coap.Service
}

func NewTracingMiddleware(tracer opentracing.Tracer, svc coap.Service) *tracingServiceMiddleware {
	return &tracingServiceMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Publish implements coap.Service
func (tm *tracingServiceMiddleware) Publish(ctx context.Context, key string, msg *messaging.Message) error {
	coapSpan := opentracing.SpanFromContext(ctx)

	span := tm.tracer.StartSpan("coap publish", opentracing.ChildOf(coapSpan.Context()))
	defer span.Finish()
	dataBuffer := bytes.NewBuffer(msg.Span)

	if err := tm.tracer.Inject(span.Context(), opentracing.Binary, dataBuffer); err != nil {
		return err
	}
	msg.Span = dataBuffer.Bytes()

	return tm.svc.Publish(ctx, key, msg)
}

// Subscribe implements coap.Service
func (tm *tracingServiceMiddleware) Subscribe(ctx context.Context, key string, chanID string, subtopic string, c coap.Client) error {
	return tm.svc.Subscribe(ctx, key, chanID, subtopic, c)
}

// Unsubscribe implements coap.Service
func (tm *tracingServiceMiddleware) Unsubscribe(ctx context.Context, key string, chanID string, subptopic string, token string) error {
	return tm.svc.Unsubscribe(ctx, key, chanID, subptopic, token)
}

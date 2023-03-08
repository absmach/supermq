package tracing

import (
	"bytes"
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
	"github.com/opentracing/opentracing-go"
)

var _ ws.Service = (*tracingMiddleware)(nil)

type tracingMiddleware struct {
	tracer opentracing.Tracer
	svc    ws.Service
}

func New(tracer opentracing.Tracer, svc ws.Service) *tracingMiddleware {
	return &tracingMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Publish implements ws.Service
func (tm *tracingMiddleware) Publish(ctx context.Context, thingKey string, msg *messaging.Message) error {
	var spanCtx opentracing.SpanContext = nil

	if coapSpan := opentracing.SpanFromContext(ctx); coapSpan != nil {
		spanCtx = coapSpan.Context()
	}
	span := tm.tracer.StartSpan("ws_publish", opentracing.ChildOf(spanCtx))
	defer span.Finish()
	dataBuffer := bytes.NewBuffer(msg.Span)

	if err := tm.tracer.Inject(span.Context(), opentracing.Binary, dataBuffer); err != nil {
		return err
	}
	msg.Span = dataBuffer.Bytes()
	return tm.svc.Publish(ctx, thingKey, msg)
}

// Subscribe implements ws.Service
func (tm *tracingMiddleware) Subscribe(ctx context.Context, thingKey string, chanID string, subtopic string, client *ws.Client) error {
	return tm.svc.Subscribe(ctx, thingKey, chanID, subtopic, client)
}

// Unsubscribe implements ws.Service
func (tm *tracingMiddleware) Unsubscribe(ctx context.Context, thingKey string, chanID string, subtopic string) error {
	return tm.svc.Unsubscribe(ctx, thingKey, chanID, subtopic)
}

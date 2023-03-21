package tracing

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
	"github.com/opentracing/opentracing-go"
)

var _ ws.Service = (*tracingMiddleware)(nil)

const (
	publish_op     = "ws_publish"
	subscribe_op   = "ws_subscribe"
	unsubscribe_op = "ws_unsubscribe"
)

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

	if wsSpan := opentracing.SpanFromContext(ctx); wsSpan != nil {
		spanCtx = wsSpan.Context()
	}
	span := tm.tracer.StartSpan(publish_op, opentracing.ChildOf(spanCtx))
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Publish(ctx, thingKey, msg)
}

// Subscribe implements ws.Service
func (tm *tracingMiddleware) Subscribe(ctx context.Context, thingKey string, chanID string, subtopic string, client *ws.Client) error {
	var spanCtx opentracing.SpanContext = nil

	if wsSpan := opentracing.SpanFromContext(ctx); wsSpan != nil {
		spanCtx = wsSpan.Context()
	}
	span := tm.tracer.StartSpan(subscribe_op, opentracing.ChildOf(spanCtx))
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Subscribe(ctx, thingKey, chanID, subtopic, client)
}

// Unsubscribe implements ws.Service
func (tm *tracingMiddleware) Unsubscribe(ctx context.Context, thingKey string, chanID string, subtopic string) error {
	var spanCtx opentracing.SpanContext = nil

	if wsSpan := opentracing.SpanFromContext(ctx); wsSpan != nil {
		spanCtx = wsSpan.Context()
	}
	span := tm.tracer.StartSpan(unsubscribe_op, opentracing.ChildOf(spanCtx))
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Unsubscribe(ctx, thingKey, chanID, subtopic)
}

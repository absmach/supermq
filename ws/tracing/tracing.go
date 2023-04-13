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

// New creates new ws tracing middleware service
func New(tracer opentracing.Tracer, svc ws.Service) ws.Service {
	return &tracingMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Publish implements ws.Service
func (tm *tracingMiddleware) Publish(ctx context.Context, thingKey string, msg *messaging.Message) error {
	span := tm.createSpan(ctx, publish_op)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Publish(ctx, thingKey, msg)
}

// Subscribe implements ws.Service
func (tm *tracingMiddleware) Subscribe(ctx context.Context, thingKey string, chanID string, subtopic string, client *ws.Client) error {
	span := tm.createSpan(ctx, subscribe_op)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Subscribe(ctx, thingKey, chanID, subtopic, client)
}

// Unsubscribe implements ws.Service
func (tm *tracingMiddleware) Unsubscribe(ctx context.Context, thingKey string, chanID string, subtopic string) error {
	span := tm.createSpan(ctx, unsubscribe_op)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Unsubscribe(ctx, thingKey, chanID, subtopic)
}

func (tm *tracingMiddleware) createSpan(ctx context.Context, opName string) opentracing.Span {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		return tm.tracer.StartSpan(
			opName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	}
	return tm.tracer.StartSpan(opName)
}

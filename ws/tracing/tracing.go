package tracing

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
	"github.com/opentracing/opentracing-go"
)

var _ ws.Service = (*tracingMiddleware)(nil)

const (
	publishOP     = "publish_op"
	subscribeOP   = "subscribe_op"
	unsubscribeOP = "unsubscribe_op"
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

// Publish trace ws publish operations
func (tm *tracingMiddleware) Publish(ctx context.Context, thingKey string, msg *messaging.Message) error {
	span := tm.createSpan(ctx, publishOP)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Publish(ctx, thingKey, msg)
}

// Subscribe trace ws subscribe opertions
func (tm *tracingMiddleware) Subscribe(ctx context.Context, thingKey string, chanID string, subtopic string, client *ws.Client) error {
	span := tm.createSpan(ctx, subscribeOP)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Subscribe(ctx, thingKey, chanID, subtopic, client)
}

// Unsubscribe trace ws unsubscibe operations
func (tm *tracingMiddleware) Unsubscribe(ctx context.Context, thingKey string, chanID string, subtopic string) error {
	span := tm.createSpan(ctx, unsubscribeOP)
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

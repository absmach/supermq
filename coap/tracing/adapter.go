package tracing

import (
	"context"

	"github.com/mainflux/mainflux/coap"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
)

var _ coap.Service = (*tracingServiceMiddleware)(nil)

const (
	publish_op     = "publish"
	subscribe_op   = "subscirbe"
	unsubscribe_op = "unsubscribe"
)

type tracingServiceMiddleware struct {
	tracer opentracing.Tracer
	svc    coap.Service
}

// New create new coap tracing service
func New(tracer opentracing.Tracer, svc coap.Service) coap.Service {
	return &tracingServiceMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Publish implements coap.Service
func (tm *tracingServiceMiddleware) Publish(ctx context.Context, key string, msg *messaging.Message) error {
	span := tm.createSpan(ctx, publish_op)
	defer span.Finish()

	ctx = opentracing.ContextWithSpan(ctx, span)

	return tm.svc.Publish(ctx, key, msg)
}

// Subscribe implements coap.Service
func (tm *tracingServiceMiddleware) Subscribe(ctx context.Context, key string, chanID string, subtopic string, c coap.Client) error {
	span := tm.createSpan(ctx, subscribe_op)
	defer span.Finish()

	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Subscribe(ctx, key, chanID, subtopic, c)
}

// Unsubscribe implements coap.Service
func (tm *tracingServiceMiddleware) Unsubscribe(ctx context.Context, key string, chanID string, subptopic string, token string) error {
	span := tm.createSpan(ctx, unsubscribe_op)
	defer span.Finish()

	ctx = opentracing.ContextWithSpan(ctx, span)
	return tm.svc.Unsubscribe(ctx, key, chanID, subptopic, token)
}

func (tm *tracingServiceMiddleware) createSpan(ctx context.Context, opName string) opentracing.Span {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		return tm.tracer.StartSpan(
			opName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	}
	return tm.tracer.StartSpan(opName)
}

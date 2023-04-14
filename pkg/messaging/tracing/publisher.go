package tracing

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// traced ops
const publishOP = "publish_op"

var _ messaging.Publisher = (*publisherMiddleware)(nil)

type publisherMiddleware struct {
	publisher messaging.Publisher
	tracer    opentracing.Tracer
}

// New creates new messaging publisher tracing middleware
func New(publisher messaging.Publisher, tracer opentracing.Tracer) messaging.Publisher {
	return &publisherMiddleware{
		publisher: publisher,
		tracer:    tracer,
	}
}

// Publish implements messaging.Publisher
func (pm *publisherMiddleware) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	span := createSpan(ctx, publishOP, msg.Subtopic, topic, msg.Publisher, pm.tracer)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return pm.publisher.Publish(ctx, topic, msg)
}

// Close implements messaging.Publisher
func (pm *publisherMiddleware) Close() error {
	return pm.publisher.Close()
}

func createSpan(ctx context.Context, operation, destination, topic, publisher string, tracer opentracing.Tracer) opentracing.Span {
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, tracer, operation)
	if destination != "" {
		ext.MessageBusDestination.Set(span, destination)
	}
	span.SetTag("publisher", publisher)
	span.SetTag("topic", topic)
	return span
}

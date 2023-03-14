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

func New(publisher messaging.Publisher, tracer opentracing.Tracer) *publisherMiddleware {
	return &publisherMiddleware{
		publisher: publisher,
		tracer:    tracer,
	}
}

// Close implements messaging.Publisher
func (pm *publisherMiddleware) Close() error {
	return pm.publisher.Close()
}

// Publish implements messaging.Publisher
func (pm *publisherMiddleware) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, pm.tracer, publishOP)
	ext.MessageBusDestination.Set(span, msg.Subtopic)
	defer span.Finish()

	ctx = opentracing.ContextWithSpan(ctx, span)

	return pm.publisher.Publish(ctx, topic, msg)
}

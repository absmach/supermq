package tracing

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/opentracing/opentracing-go"
)

const (
	publishOP = "publish_op"
)

var _ mainflux.MessagePublisher = (*natsPublisherMiddleware)(nil)

type natsPublisherMiddleware struct {
	tracer    opentracing.Tracer
	publisher mainflux.MessagePublisher
}

// NatsPublisherMiddleware add spans to context
func NatsPublisherMiddleware(publisher mainflux.MessagePublisher, tracer opentracing.Tracer) mainflux.MessagePublisher {
	return natsPublisherMiddleware{
		tracer:    tracer,
		publisher: publisher,
	}
}

func (npm natsPublisherMiddleware) Publish(ctx context.Context, token string, msg mainflux.RawMessage) error {
	span := createSpan(ctx, npm.tracer, publishOP)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return npm.publisher.Publish(ctx, token, msg)
}

func createSpan(ctx context.Context, tracer opentracing.Tracer, opName string) opentracing.Span {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		return tracer.StartSpan(
			opName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	}
	return tracer.StartSpan(opName)
}

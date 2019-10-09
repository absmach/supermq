package tracing

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/coap"
	"github.com/opentracing/opentracing-go"
)

const (
	publishOP = "publish_op"
)

var _ mainflux.MessagePublisher = (*natsPublisherMiddleware)(nil)

type natsPublisherMiddleware struct {
	tracer opentracing.Tracer
	pubsub coap.Broker
}

// NatsPublisherMiddleware add spans to context
func NatsPublisherMiddleware(pubsub coap.Broker, tracer opentracing.Tracer) coap.Broker {
	return natsPublisherMiddleware{
		tracer: tracer,
		pubsub: pubsub,
	}
}

func (npm natsPublisherMiddleware) Publish(ctx context.Context, token string, msg mainflux.RawMessage) error {
	span := createSpan(ctx, npm.tracer, publishOP)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return npm.pubsub.Publish(ctx, token, msg)
}

func (npm natsPublisherMiddleware) Subscribe(chanID, subtopic, obsID string, observer *coap.Observer) error {
	span := createSpan(context.TODO(), npm.tracer, publishOP)
	defer span.Finish()

	return npm.Subscribe(chanID, subtopic, obsID, observer)
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

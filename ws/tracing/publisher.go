package tracing

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
	"github.com/opentracing/opentracing-go"
)

const (
	publishOP = "publish_op"
)

var _ ws.Service = (*natsPublisherMiddleware)(nil)

type natsPublisherMiddleware struct {
	tracer opentracing.Tracer
	svc    ws.Service
}

// NatsPublisherMiddleware add spans to context
func NatsPublisherMiddleware(svc ws.Service, tracer opentracing.Tracer) ws.Service {
	return natsPublisherMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

func (npm natsPublisherMiddleware) Publish(ctx context.Context, token string, msg mainflux.RawMessage) error {
	span := createSpan(ctx, npm.tracer, publishOP)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return npm.svc.Publish(ctx, token, msg)
}

func (npm natsPublisherMiddleware) Subscribe(chanID, subtopic string, channel *ws.Channel) error {
	span := createSpan(context.TODO(), npm.tracer, publishOP)
	defer span.Finish()

	return npm.svc.Subscribe(chanID, subtopic, channel)
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

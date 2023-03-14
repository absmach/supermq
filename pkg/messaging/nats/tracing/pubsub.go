package tracing

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const subscribeOP = "subscribe_op" //traced op

var _ messaging.PubSub = (*pubsubMiddleware)(nil)

type pubsubMiddleware struct {
	pubsub messaging.PubSub
	tracer opentracing.Tracer
}

func NewPubSub(pubsub messaging.PubSub, tracer opentracing.Tracer) messaging.PubSub {
	return &pubsubMiddleware{
		pubsub: pubsub,
		tracer: tracer,
	}
}

// Close implements messaging.PubSub
func (ps *pubsubMiddleware) Close() error {
	return ps.pubsub.Close()
}

// Publish implements messaging.PubSub
func (ps *pubsubMiddleware) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, ps.tracer, publishOP)
	ext.MessageBusDestination.Set(span, msg.Subtopic)
	defer span.Finish()

	ctx = opentracing.ContextWithSpan(ctx, span)

	return ps.pubsub.Publish(ctx, topic, msg)
}

// Subscribe implements messaging.PubSub
func (ps *pubsubMiddleware) Subscribe(id string, topic string, handler messaging.MessageHandler) error {
	return ps.pubsub.Subscribe(id, topic, ps.handle(handler))
}

// Unsubscribe implements messaging.PubSub
func (ps *pubsubMiddleware) Unsubscribe(id string, topic string) error {
	return ps.pubsub.Unsubscribe(id, topic)
}

func (ps *pubsubMiddleware) handle(h messaging.MessageHandler) handleFunc {
	return func(msg *messaging.Message) error {
		span := ps.tracer.StartSpan(subscribeOP, ext.SpanKindConsumer)
		ext.MessageBusDestination.Set(span, msg.Subtopic)
		defer span.Finish()
		return h.Handle(msg)
	}
}

type handleFunc func(msg *messaging.Message) error

func (h handleFunc) Handle(msg *messaging.Message) error {
	return h(msg)
}

func (h handleFunc) Cancel() error {
	return nil
}

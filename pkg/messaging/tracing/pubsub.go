package tracing

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	subscribeOP   = "subscribe_op"
	unsubscribeOp = "unsubscribe_op"
	handleOp      = "handle_op"
)

var _ messaging.PubSub = (*pubsubMiddleware)(nil)

type pubsubMiddleware struct {
	publisherMiddleware
	pubsub messaging.PubSub
	tracer opentracing.Tracer
}

func NewPubSub(pubsub messaging.PubSub, tracer opentracing.Tracer) messaging.PubSub {
	return &pubsubMiddleware{
		publisherMiddleware: publisherMiddleware{
			publisher: pubsub,
			tracer:    tracer,
		},
		pubsub: pubsub,
		tracer: tracer,
	}
}

// Subscribe implements messaging.PubSub
func (pm *pubsubMiddleware) Subscribe(ctx context.Context, id string, topic string, handler messaging.MessageHandler) error {
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, pm.tracer, subscribeOP)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return pm.pubsub.Subscribe(ctx, id, topic, pm.handle(handler))
}

// Unsubscribe implements messaging.PubSub
func (pm *pubsubMiddleware) Unsubscribe(ctx context.Context, id string, topic string) error {
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, pm.tracer, unsubscribeOp)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return pm.pubsub.Unsubscribe(ctx, id, topic)
}

func (ps *pubsubMiddleware) handle(h messaging.MessageHandler) handleFunc {
	return func(msg *messaging.Message) error {
		span := ps.tracer.StartSpan(handleOp, ext.SpanKindConsumer)
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

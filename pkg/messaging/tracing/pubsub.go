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
	ext.MessageBusDestination.Set(span, topic)
	span.SetTag("subscriber", id)
	span.SetTag("topic", topic)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	h := &traceHandler{
		handler: handler,
		tracer:  pm.tracer,
		ctx:     ctx,
	}
	return pm.pubsub.Subscribe(ctx, id, topic, h)
}

// Unsubscribe implements messaging.PubSub
func (pm *pubsubMiddleware) Unsubscribe(ctx context.Context, id string, topic string) error {
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, pm.tracer, unsubscribeOp)
	span.SetTag("topic", topic)
	span.SetTag("subscriber", id)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return pm.pubsub.Unsubscribe(ctx, id, topic)
}

type traceHandler struct {
	handler messaging.MessageHandler
	tracer  opentracing.Tracer
	ctx     context.Context
}

func (h *traceHandler) Handle(msg *messaging.Message) error {
	span, _ := opentracing.StartSpanFromContextWithTracer(h.ctx, h.tracer, handleOp, ext.SpanKindConsumer)
	ext.MessageBusDestination.Set(span, msg.Subtopic)
	span.SetTag("publisher", msg.Publisher)
	span.SetTag("protocol", msg.Protocol)
	span.SetTag("channel", msg.Channel)
	defer span.Finish()
	return h.handler.Handle(msg)
}

func (h *traceHandler) Cancel() error {
	return h.handler.Cancel()
}

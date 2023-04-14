package tracing

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
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

// NewPubSub creates new pubsub tracing middleware service
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
	span := createSpan(ctx, subscribeOP, topic, topic, id, pm.tracer)
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
	span := createSpan(ctx, unsubscribeOp, "", topic, id, pm.tracer)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return pm.pubsub.Unsubscribe(ctx, id, topic)
}

type traceHandler struct {
	handler messaging.MessageHandler
	tracer  opentracing.Tracer
	ctx     context.Context
}

// Handle tracing middleware handle for message handler
func (h *traceHandler) Handle(msg *messaging.Message) error {
	span := createSpan(h.ctx, handleOp, msg.Subtopic, msg.Channel, msg.Publisher, h.tracer)
	defer span.Finish()
	return h.handler.Handle(msg)
}

// Cancel tracing middleware handle for message handler
func (h *traceHandler) Cancel() error {
	return h.handler.Cancel()
}

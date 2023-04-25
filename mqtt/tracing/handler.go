package tracing

import (
	"context"

	"github.com/mainflux/mproxy/pkg/session"
	"github.com/opentracing/opentracing-go"
)

const (
	authConnectOP   = "auth_connect_op"
	authPublishOP   = "auth_publish_op"
	authSubscribeOP = "auth_subscribe_op"
	connectOP       = "connect_op"
	disconnectOP    = "disconnect_op"
	subscribeOP     = "subscribe_op"
	unsubscribeOP   = "unsubscribe_op"
	publishOP       = "publish_op"
)

var _ session.Handler = (*handlerMiddleware)(nil)

type handlerMiddleware struct {
	handler session.Handler
	tracer  opentracing.Tracer
}

func NewHandler(tracer opentracing.Tracer, handler session.Handler) session.Handler {
	return &handlerMiddleware{
		tracer:  tracer,
		handler: handler,
	}
}

// AuthConnect implements session.Handler
func (h *handlerMiddleware) AuthConnect(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, authConnectOP)
	defer span.Finish()
	return h.handler.AuthConnect(ctx)
}

// AuthPublish implements session.Handler
func (h *handlerMiddleware) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, authPublishOP)
	defer span.Finish()
	return h.handler.AuthPublish(ctx, topic, payload)
}

// AuthSubscribe implements session.Handler
func (h *handlerMiddleware) AuthSubscribe(ctx context.Context, topics *[]string) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, authSubscribeOP)
	defer span.Finish()
	return h.handler.AuthSubscribe(ctx, topics)
}

// Connect implements session.Handler
func (h *handlerMiddleware) Connect(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, connectOP)
	defer span.Finish()
	h.handler.Connect(ctx)
}

// Disconnect implements session.Handler
func (h *handlerMiddleware) Disconnect(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, disconnectOP)
	defer span.Finish()
	h.handler.Disconnect(ctx)
}

// Publish implements session.Handler
func (h *handlerMiddleware) Publish(ctx context.Context, topic *string, payload *[]byte) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, publishOP)
	defer span.Finish()
	h.handler.Publish(ctx, topic, payload)
}

// Subscribe implements session.Handler
func (h *handlerMiddleware) Subscribe(ctx context.Context, topics *[]string) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, subscribeOP)
	defer span.Finish()
	h.handler.Subscribe(ctx, topics)
}

// Unsubscribe implements session.Handler
func (h *handlerMiddleware) Unsubscribe(ctx context.Context, topics *[]string) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, h.tracer, unsubscribeOP)
	defer span.Finish()
	h.handler.Unsubscribe(ctx, topics)
}

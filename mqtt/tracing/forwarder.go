package tracing

import (
	"context"

	"github.com/mainflux/mainflux/mqtt"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
)

const forwardOP = "forward_op"

var _ mqtt.Forwarder = (*forwarderMiddleware)(nil)

type forwarderMiddleware struct {
	topic     string
	forwarder mqtt.Forwarder
	tracer    opentracing.Tracer
}

// New creates new mqtt forwarder tracing middleware.
func New(tracer opentracing.Tracer, forwarder mqtt.Forwarder) mqtt.Forwarder {
	return &forwarderMiddleware{
		forwarder: forwarder,
		tracer:    tracer,
	}
}

// Forward traces mqtt forward operations
func (fm *forwarderMiddleware) Forward(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error {
	span, _ := opentracing.StartSpanFromContextWithTracer(ctx, fm.tracer, forwardOP)
	defer span.Finish()
	span.SetTag("subscriber", id)
	span.SetTag("topic", fm.topic)
	ctx = opentracing.ContextWithSpan(ctx, span)
	return fm.forwarder.Forward(ctx, id, sub, pub)
}

package tracing

import (
	notifiers "github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const notifierOp = "notifier_op"

var _ notifiers.Notifier = (*serviceMiddleware)(nil)

type serviceMiddleware struct {
	svc    notifiers.Notifier
	tracer opentracing.Tracer
}

// NewNotifier create new notifier tracing middleware service
func NewNotifier(svc notifiers.Notifier, tracer opentracing.Tracer) notifiers.Notifier {
	return &serviceMiddleware{
		svc:    svc,
		tracer: tracer,
	}
}

// Notify implements notifiers.Notifier
func (sm *serviceMiddleware) Notify(from string, to []string, msg *messaging.Message) error {
	span := sm.tracer.StartSpan(notifierOp, ext.SpanKindConsumer)
	ext.MessageBusDestination.Set(span, msg.Subtopic)
	defer span.Finish()
	return sm.svc.Notify(from, to, msg)
}

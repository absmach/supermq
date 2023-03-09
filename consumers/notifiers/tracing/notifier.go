package tracing

import (
	"bytes"

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

// Notify implements notifiers.Notifier
func (sm *serviceMiddleware) Notify(from string, to []string, msg *messaging.Message) error {
	var sc opentracing.SpanContext = nil
	var err error
	//extract span context from message
	if len(msg.Span) != 0 {
		parentSCBuffer := bytes.NewBuffer(msg.Span)
		sc, err = sm.tracer.Extract(opentracing.Binary, parentSCBuffer)

		if err != nil && err != opentracing.ErrSpanContextNotFound {
			return err
		}
	}

	// start new span
	span := sm.tracer.StartSpan(notifierOp, ext.SpanKindProducer, opentracing.ChildOf(sc))
	ext.MessageBusDestination.Set(span, msg.Subtopic)
	defer span.Finish()
	return sm.svc.Notify(from, to, msg)
}

func NewNotifier(svc notifiers.Notifier, tracer opentracing.Tracer) notifiers.Notifier {
	return &serviceMiddleware{
		svc:    svc,
		tracer: tracer,
	}
}

package tracing

import (
	"bytes"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// traced ops
const (
	publishOP = "publish_op"
)

var _ messaging.Publisher = (*publisherMiddleware)(nil)

type publisherMiddleware struct {
	publisher messaging.Publisher
	tracer    opentracing.Tracer
}

func NewPublisherMiddleware(publisher messaging.Publisher, tracer opentracing.Tracer) *publisherMiddleware {
	return &publisherMiddleware{
		publisher: publisher,
		tracer:    tracer,
	}
}

// Close implements messaging.Publisher
func (pm *publisherMiddleware) Close() error {
	return pm.publisher.Close()
}

// Publish implements messaging.Publisher
func (pm *publisherMiddleware) Publish(topic string, msg *messaging.Message) error {
	//extract span context from message
	parentSCBuffer := bytes.NewBuffer(msg.Span)
	sc, err := pm.tracer.Extract(opentracing.Binary, parentSCBuffer)

	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return err
	}

	// start new span
	span := pm.tracer.StartSpan(publishOP, ext.SpanKindProducer, opentracing.ChildOf(sc))
	ext.MessageBusDestination.Set(span, msg.Subtopic)
	defer span.Finish()

	dataBuffer := bytes.NewBuffer(msg.Span)

	if err := pm.tracer.Inject(span.Context(), opentracing.Binary, dataBuffer); err != nil {
		return err
	}
	msg.Span = dataBuffer.Bytes()
	return pm.publisher.Publish(topic, msg)
}

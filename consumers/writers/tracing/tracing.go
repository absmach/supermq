package tracing

import (
	"github.com/mainflux/mainflux/consumers"
	mfjson "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

var _ consumers.Consumer = (*tracingMiddleware)(nil)

const consume_op = "consume_op"

type tracingMiddleware struct {
	tracer   opentracing.Tracer
	consumer consumers.Consumer
}

// New creates new consumers tracing middleware service
func New(tracer opentracing.Tracer, consumer consumers.Consumer) consumers.Consumer {
	return &tracingMiddleware{
		tracer:   tracer,
		consumer: consumer,
	}
}

// Consume implements consumers.Consumer
func (tm *tracingMiddleware) Consume(message interface{}) error {
	switch m := message.(type) {
	case mfjson.Messages:
		for _, msg := range m.Data {
			span := tm.tracer.StartSpan(consume_op, ext.SpanKindConsumer)
			defer span.Finish()
			ext.MessageBusDestination.Set(span, msg.Subtopic)
		}
	case []senml.Message:
		for _, msg := range m {
			span := tm.tracer.StartSpan(consume_op, ext.SpanKindConsumer)
			defer span.Finish()
			ext.MessageBusDestination.Set(span, msg.Subtopic)
		}
	default:
	}
	return tm.consumer.Consume(message)
}

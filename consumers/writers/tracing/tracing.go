package tracing

import (
	"bytes"

	"github.com/mainflux/mainflux/consumers"
	mfjson "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/opentracing/opentracing-go"
)

var _ consumers.Consumer = (*tracingMiddleware)(nil)

const consume_op = "consume_op"

type tracingMiddleware struct {
	tracer   opentracing.Tracer
	consumer consumers.Consumer
}

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
		for _, mes := range m.Data {
			span, err := tm.getSpanFromBytes(mes.Span)
			if err != nil {
				return err
			}
			defer span.Finish()
		}
	case []senml.Message:
		for _, mes := range m {
			span, err := tm.getSpanFromBytes(mes.Span)
			if err != nil {
				return err
			}
			defer span.Finish()
		}
	default:
	}
	return tm.Consume(message)
}

func (tm *tracingMiddleware) getSpanFromBytes(span []byte) (opentracing.Span, error) {
	buf := bytes.NewBuffer(span)

	sc, err := tm.tracer.Extract(opentracing.Binary, buf)
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, err
	}
	return tm.tracer.StartSpan(consume_op, opentracing.ChildOf(sc)), nil
}

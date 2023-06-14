package tracing

import (
	"context"

	"github.com/mainflux/mainflux/consumers"
	mfjson "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	consumeBlockingOP = "consume_blocking_op"
	consumeAsyncOP    = "consume_async_op"
)

var _ consumers.AsyncConsumer = (*tracingMiddlewareAsync)(nil)
var _ consumers.BlockingConsumer = (*tracingMiddlewareBlock)(nil)

type tracingMiddlewareAsync struct {
	consumer consumers.AsyncConsumer
	tracer   trace.Tracer
}
type tracingMiddlewareBlock struct {
	consumer consumers.BlockingConsumer
	tracer   trace.Tracer
}

// NewAsync creates a new traced consumers.AsyncConsumer service.
func NewAsync(tracer trace.Tracer, consumerAsync consumers.AsyncConsumer) consumers.AsyncConsumer {
	return &tracingMiddlewareAsync{
		consumer: consumerAsync,
		tracer:   tracer,
	}
}

// NewBlocking creates a new traced consumers.BlockingConsumer service.
func NewBlocking(tracer trace.Tracer, consumerBlock consumers.BlockingConsumer) consumers.BlockingConsumer {
	return &tracingMiddlewareBlock{
		consumer: consumerBlock,
		tracer:   tracer,
	}
}

// ConsumeBlocking  traces consume operations for message/s consumed.
func (tm *tracingMiddlewareBlock) ConsumeBlocking(ctx context.Context, messages interface{}) error {
	var span trace.Span
	switch m := messages.(type) {
	case mfjson.Messages:
		if len(m.Data) > 0 {
			firstMsg := m.Data[0]
			ctx, span = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeBlockingOP, len(m.Data))
			defer span.End()
		}
	case []senml.Message:
		if len(m) > 0 {
			firstMsg := m[0]
			ctx, span = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeBlockingOP, len(m))
			defer span.End()
		}
	}
	return tm.consumer.ConsumeBlocking(ctx, messages)
}

// ConsumeAsync traces consume operations for message/s consumed.
func (tm *tracingMiddlewareAsync) ConsumeAsync(ctx context.Context, messages interface{}) {
	var span trace.Span
	switch m := messages.(type) {
	case mfjson.Messages:
		if len(m.Data) > 0 {
			firstMsg := m.Data[0]
			ctx, span = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeAsyncOP, len(m.Data))
			defer span.End()
		}
	case []senml.Message:
		if len(m) > 0 {
			firstMsg := m[0]
			ctx, span = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeAsyncOP, len(m))
			defer span.End()
		}
	}
	tm.consumer.ConsumeAsync(ctx, messages)
}

// Errors traces async consume errors.
func (tm *tracingMiddlewareAsync) Errors() <-chan error {
	return tm.consumer.Errors()
}

func createMessageSpan(ctx context.Context, tracer trace.Tracer, topic, subTopic, publisher, operation string, noMessages int) (context.Context, trace.Span) {
	kvOpts := []attribute.KeyValue{}
	kvOpts = append(kvOpts, attribute.String("topic", topic), attribute.String("publisher", publisher), attribute.Int("number_of_messages", noMessages))
	if subTopic != "" {
		kvOpts = append(kvOpts, attribute.String("subtopic", subTopic))
	}
	return tracer.Start(ctx, operation, trace.WithAttributes(kvOpts...))
}

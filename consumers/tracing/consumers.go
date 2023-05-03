package tracing

import (
	"context"

	"github.com/mainflux/mainflux/consumers"
	mfjson "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	consumeBlockingOP = "consume_blocking_op"
	consumeAsyncOP    = "consume_async_op"
)

var _ consumers.AsyncConsumer = (*tracingMiddlewareAsync)(nil)
var _ consumers.BlockingConsumer = (*tracingMiddlewareBlock)(nil)

type tracingMiddlewareAsync struct {
	consumerAsync consumers.AsyncConsumer
	tracer        opentracing.Tracer
}
type tracingMiddlewareBlock struct {
	consumerBlock consumers.BlockingConsumer
	tracer        opentracing.Tracer
}

// NewAsync creates a new traced consumers.AsyncConsumer service
func NewAsync(tracer opentracing.Tracer, consumerAsync consumers.AsyncConsumer) consumers.AsyncConsumer {
	return &tracingMiddlewareAsync{
		consumerAsync: consumerAsync,
		tracer:        tracer,
	}
}

// NewBlocking creates a new traced consumers.BlockingConsumer service
func NewBlocking(tracer opentracing.Tracer, consumerBlock consumers.BlockingConsumer) consumers.BlockingConsumer {
	return &tracingMiddlewareBlock{
		consumerBlock: consumerBlock,
		tracer:        tracer,
	}
}

// ConsumeBlocking  traces consume operations for message/s consumed.
func (tm *tracingMiddlewareBlock) ConsumeBlocking(ctx context.Context, messages interface{}) error {
	var span opentracing.Span
	switch m := messages.(type) {
	case mfjson.Messages:
		firstMsg := m.Data[0]
		span, ctx = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeBlockingOP, len(m.Data))
		defer span.Finish()
	case []senml.Message:
		firstMsg := m[0]
		span, ctx = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeBlockingOP, len(m))
		defer span.Finish()
	}
	return tm.consumerBlock.ConsumeBlocking(ctx, messages)
}

// ConsumeAsync traces consume operations for message/s consumed.
func (tm *tracingMiddlewareAsync) ConsumeAsync(ctx context.Context, messages interface{}) {
	var span opentracing.Span
	switch m := messages.(type) {
	case mfjson.Messages:
		firstMsg := m.Data[0]
		span, ctx = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeAsyncOP, len(m.Data))
		defer span.Finish()
	case []senml.Message:
		firstMsg := m[0]
		span, ctx = createMessageSpan(ctx, tm.tracer, firstMsg.Channel, firstMsg.Subtopic, firstMsg.Publisher, consumeAsyncOP, len(m))
		defer span.Finish()
	}
	tm.consumerAsync.ConsumeAsync(ctx, messages)
}

// Errors traces async consume errors.
func (tm *tracingMiddlewareAsync) Errors() <-chan error {
	return tm.consumerAsync.Errors()
}

func createMessageSpan(ctx context.Context, tracer opentracing.Tracer, topic, subTopic, publisher, operation string, noMessages int) (opentracing.Span, context.Context) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, operation, ext.SpanKindConsumer)
	span.SetTag("topic", topic)
	if subTopic != "" {
		span.SetTag("sub-topic", subTopic)
	}
	span.SetTag("publisher", publisher)
	span.SetTag("number_of_messages", noMessages)
	return span, ctx
}

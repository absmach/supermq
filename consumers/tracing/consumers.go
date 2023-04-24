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

// NewAsync creates a new traced consumers.Blocking service
func NewAsync(tracer opentracing.Tracer, consumerAsync consumers.AsyncConsumer) consumers.AsyncConsumer {
	return &tracingMiddlewareAsync{
		consumerAsync: consumerAsync,
		tracer:        tracer,
	}
}

// NewBlocking creates a new traced consumers.Blocking service
func NewBlocking(tracer opentracing.Tracer, consumerBlock consumers.BlockingConsumer) consumers.BlockingConsumer {
	return &tracingMiddlewareBlock{
		consumerBlock: consumerBlock,
		tracer:        tracer,
	}
}

// ConsumeBlocking  traces consume operations for each message.
func (tm *tracingMiddlewareBlock) ConsumeBlocking(ctx context.Context, messages interface{}) error {
	switch m := messages.(type) {
	case mfjson.Messages:
		span, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeBlockingOP, ext.SpanKindConsumer)
		span.SetTag("number of messages", len(m.Data))
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		for _, mes := range m.Data {
			mesSpan, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeBlockingOP)
			mesSpan.SetTag("topic", mes.Channel)
			mesSpan.SetTag("sub-topic", mes.Subtopic)
			mesSpan.SetTag("publisher", mes.Publisher)
			ctx = opentracing.ContextWithSpan(ctx, span)
			defer mesSpan.Finish()
		}
	case []senml.Message:
		span, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeBlockingOP, ext.SpanKindConsumer)
		span.SetTag("number of messages", len(m))
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		for _, mes := range m {
			mesSpan, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeBlockingOP)
			mesSpan.SetTag("topic", mes.Channel)
			mesSpan.SetTag("sub-topic", mes.Subtopic)
			mesSpan.SetTag("publisher", mes.Publisher)
			ctx = opentracing.ContextWithSpan(ctx, span)
			defer mesSpan.Finish()
			// NewBlocking creates a new traced consumers.Blocking service mesSpan.Finish()
		}
	}
	return tm.consumerBlock.ConsumeBlocking(ctx, messages)
}

// ConsumeAsync traces consume operations for each message.
func (tm *tracingMiddlewareAsync) ConsumeAsync(ctx context.Context, messages interface{}) {
	switch m := messages.(type) {
	case mfjson.Messages:
		span, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeAsyncOP, ext.SpanKindConsumer)
		span.SetTag("number of messages", len(m.Data))
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		for _, mes := range m.Data {
			mesSpan, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeAsyncOP)
			mesSpan.SetTag("topic", mes.Channel)
			mesSpan.SetTag("sub-topic", mes.Subtopic)
			mesSpan.SetTag("publisher", mes.Publisher)
			ctx = opentracing.ContextWithSpan(ctx, mesSpan)
			defer mesSpan.Finish()
		}
	case []senml.Message:
		span, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeAsyncOP, ext.SpanKindConsumer)
		span.SetTag("number of messages", len(m))
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
		for _, mes := range m {
			mesSpan, _ := opentracing.StartSpanFromContextWithTracer(ctx, tm.tracer, consumeAsyncOP)
			mesSpan.SetTag("topic", mes.Channel)
			mesSpan.SetTag("sub-topic", mes.Subtopic)
			mesSpan.SetTag("publisher", mes.Publisher)
			ctx = opentracing.ContextWithSpan(ctx, mesSpan)
			defer mesSpan.Finish()
		}
	}
	tm.consumerAsync.ConsumeAsync(ctx, messages)
}

// Errors traces async consume errors.
func (tm *tracingMiddlewareAsync) Errors() <-chan error {
	return tm.consumerAsync.Errors()
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"bytes"
	"fmt"

	"github.com/mainflux/mainflux/pkg/messaging"
	broker "github.com/nats-io/nats.go"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/protobuf/proto"
)

// A maximum number of reconnect attempts before NATS connection closes permanently.
// Value -1 represents an unlimited number of reconnect retries, i.e. the client
// will never give up on retrying to re-establish connection to NATS server.
const maxReconnects = -1

var _ messaging.Publisher = (*publisher)(nil)

// traced ops
const (
	publishOP   = "publish_op"
	subscribeOP = "subscribe_op"
)

type publisher struct {
	conn   *broker.Conn
	tracer opentracing.Tracer
}

// Publisher wraps messaging Publisher exposing
// Close() method for NATS connection.

// NewPublisher returns NATS message Publisher.
func NewPublisher(url string, tracer opentracing.Tracer) (messaging.Publisher, error) {
	conn, err := broker.Connect(url, broker.MaxReconnects(maxReconnects))
	if err != nil {
		return nil, err
	}
	ret := &publisher{
		conn:   conn,
		tracer: tracer,
	}
	return ret, nil
}

func (pub *publisher) Publish(topic string, msg *messaging.Message, spanContext opentracing.SpanContext) error {
	if topic == "" {
		return ErrEmptyTopic
	}

	span := pub.tracer.StartSpan(publishOP, ext.SpanKindProducer, opentracing.ChildOf(spanContext))
	ext.MessageBusDestination.Set(span, msg.Subtopic)
	defer span.Finish()

	dataBuffer := bytes.NewBuffer(msg.Span)

	if err := pub.tracer.Inject(span.Context(), opentracing.Binary, dataBuffer); err != nil {
		return err
	}
	msg.Span = dataBuffer.Bytes()

	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	if err := pub.conn.Publish(subject, data); err != nil {
		return err
	}

	return nil
}

func (pub *publisher) Close() error {
	pub.conn.Close()
	return nil
}

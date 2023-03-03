// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
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

func (pub *publisher) Publish(topic string, msg *messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	span := opentracing.StartSpan(publishOP, ext.SpanKindProducer)
	ext.MessageBusDestination.Set(span, topic)
	defer span.Finish()

	pub.tracer.Inject(span.Context(), opentracing.Binary, data)

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

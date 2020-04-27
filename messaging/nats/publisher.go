// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/messaging"
	broker "github.com/nats-io/nats.go"
)

var _ messaging.Publisher = (*pubsliher)(nil)

type pubsliher struct {
	conn *broker.Conn
}

// NewPublisher returns NATS message Publisher.
func NewPublisher(conn *broker.Conn) messaging.Publisher {
	return &pubsub{
		conn: conn,
	}
}

func (pub *pubsliher) Publish(topic string, msg messaging.Message) error {
	data, err := proto.Marshal(&msg)
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

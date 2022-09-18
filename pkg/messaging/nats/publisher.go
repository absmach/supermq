// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	broker "github.com/nats-io/nats.go"
)

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	conn *broker.Conn
}

// Publisher wraps messaging Publisher exposing
// Close() method for NATS connection.

// NewPublisher returns NATS message Publisher.
func NewPublisher(url string) (messaging.Publisher, error) {
	conn, err := broker.Connect(url)
	if err != nil {
		return nil, err
	}
	ret := &publisher{
		conn: conn,
	}
	return ret, nil
}

func (pub *publisher) Publish(topic string, msg messaging.Message) error {
	fmt.Println()
	fmt.Println("Reached pubsub.publish")
	fmt.Println()

	if topic == "" {
		return ErrEmptyTopic
	}
	data, err := proto.Marshal(&msg)
	// fmt.Println()
	// fmt.Println("Marshal error ->", err)
	// fmt.Println()
	if err != nil {
		return err
	}

	// fmt.Println("Data after marshall ->", data)
	// fmt.Println("Data after marshall to string ->", string(data))
	// fmt.Println()

	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}

	fmt.Println()
	fmt.Println("Publishing to -> ", subject)
	fmt.Println()

	if err := pub.conn.Publish(subject, data); err != nil {
		fmt.Println()
		fmt.Println("pub.conn.Publish() error -> ", err)
		fmt.Println()

		return err
	}
	fmt.Println()
	fmt.Println("Published successfully")
	fmt.Println()

	return nil
}

func (pub *publisher) Close() error {
	pub.conn.Close()
	return nil
}

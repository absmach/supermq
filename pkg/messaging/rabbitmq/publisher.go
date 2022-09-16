// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewPublisher returns RabbitMQ message Publisher.
func NewPublisher(url string) (messaging.Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare(exchangeName, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		return nil, err
	}
	ret := &publisher{
		conn: conn,
		ch:   ch,
	}
	return ret, nil
}

func (pub *publisher) Publish(topic string, msg messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	subject := formatTopic(topic, msg)

	err = pub.ch.Publish(
		exchangeName,
		subject,
		false,
		false,
		amqp.Publishing{
			Headers:     amqp.Table{},
			ContentType: "application/octet-stream",
			AppId:       "mainflux-publisher",
			Body:        data,
		})

	if err != nil {
		return err
	}

	return nil
}

func (pub *publisher) Close() error {
	return pub.conn.Close()
}

func formatTopic(topic string, message messaging.Message) string {
	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if message.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, message.Subtopic)
	}
	subject = strings.Replace(subject, ">", "#", -1)
	return subject
}

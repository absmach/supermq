// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"context"
	"fmt"

	"github.com/absmach/supermq/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	options
}

// NewPublisher returns RabbitMQ message Publisher.
func NewPublisher(url string, opts ...messaging.Option) (messaging.Publisher, error) {
	pub := &publisher{
		options: defaultOptions(),
	}

	for _, opt := range opts {
		if err := opt(pub); err != nil {
			return nil, err
		}
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	pub.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare(pub.exchange, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		return nil, err
	}
	pub.channel = ch

	return pub, nil
}

func (pub *publisher) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", pub.prefix, topic)

	err = pub.channel.PublishWithContext(
		ctx,
		pub.exchange,
		subject,
		false,
		false,
		amqp.Publishing{
			Headers:     amqp.Table{},
			ContentType: "application/octet-stream",
			AppId:       "supermq-publisher",
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

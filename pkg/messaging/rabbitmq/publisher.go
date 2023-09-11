// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"context"
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

var _ messaging.Publisher = (*RPublisher)(nil)

type RPublisher struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Prefix  string
}

// NewPublisher returns RabbitMQ message Publisher.
func NewPublisher(url string, opts ...messaging.Option) (messaging.Publisher, error) {
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

	ret := &RPublisher{
		Conn:    conn,
		Channel: ch,
		Prefix:  chansPrefix,
	}

	for _, opt := range opts {
		if err := opt(ret); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func (pub *RPublisher) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", pub.Prefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	subject = formatTopic(subject)

	err = pub.Channel.PublishWithContext(
		ctx,
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

func (pub *RPublisher) Close() error {
	if err := pub.Channel.Close(); err != nil {
		return err
	}

	return pub.Conn.Close()
}

func formatTopic(topic string) string {
	return strings.ReplaceAll(topic, ">", "#")
}

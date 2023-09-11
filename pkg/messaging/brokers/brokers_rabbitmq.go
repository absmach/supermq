// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build rabbitmq
// +build rabbitmq

package brokers

import (
	"context"
	"log"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.#"

func init() {
	log.Println("The binary was build using RabbitMQ as the message broker")
}

func NewPublisher(_ context.Context, url string, opts ...messaging.Option) (messaging.Publisher, error) {
	pb, err := rabbitmq.NewPublisher(url, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func NewPubSub(_ context.Context, url string, logger mflog.Logger, opts ...messaging.Option) (messaging.PubSub, error) {
	pb, err := rabbitmq.NewPubSub(url, logger, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func WithExchange(url, name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) messaging.Option {
	return func(url, _ string) error {
		conn, err := amqp.Dial(url)
		if err != nil {
			return err
		}
		ch, err := conn.Channel()
		if err != nil {
			return err
		}
		if err := ch.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args); err != nil {
			return err
		}

		return nil
	}
}

func WithPrefix(prefix string) messaging.Option {
	return func(_, p string) error {
		p = prefix
		return nil
	}
}

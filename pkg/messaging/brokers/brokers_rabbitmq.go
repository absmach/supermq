// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build rabbitmq
// +build rabbitmq

package brokers

import (
	"context"
	"errors"
	"log"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.#"

// ErrInvalidType is returned when the provided value is not of the expected type.
var ErrInvalidType = errors.New("invalid type")

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

func WithChannel(channel *amqp.Channel) messaging.Option {
	return func(val interface{}) error {
		ch, ok := val.(*rabbitmq.RPublisher)
		if !ok {
			return ErrInvalidType
		}

		ch.Channel = channel

		return nil
	}
}

func WithPrefix(prefix *string) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*rabbitmq.RPublisher)
		if !ok {
			return ErrInvalidType
		}

		p.Prefix = *prefix

		return nil
	}
}

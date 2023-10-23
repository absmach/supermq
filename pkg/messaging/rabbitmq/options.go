// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"errors"

	"github.com/mainflux/mainflux/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ErrInvalidType is returned when the provided value is not of the expected type.
var ErrInvalidType = errors.New("invalid type")

func WithChannel(channel *amqp.Channel) messaging.Option {
	return func(val interface{}) error {
		switch v := val.(type) {
		case *publisher:
			v.channel = channel
		case *pubsub:
			v.channel = channel
		default:
			return ErrInvalidType
		}

		return nil
	}
}

func WithExchange(exchange *string) messaging.Option {
	return func(val interface{}) error {
		switch v := val.(type) {
		case *publisher:
			v.exchange = *exchange
		case *pubsub:
			v.exchange = *exchange
		default:
			return ErrInvalidType
		}

		return nil
	}
}

func WithPrefix(prefix *string) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*publisher)
		if !ok {
			return ErrInvalidType
		}

		p.prefix = *prefix

		return nil
	}
}

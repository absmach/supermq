// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"errors"

	"github.com/absmach/supermq/pkg/messaging"
)

// ErrInvalidType is returned when the provided value is not of the expected type.
var ErrInvalidType = errors.New("invalid type")

const (
	exchangeName = "messages"
	msgPrefix    = "m"
)

type options struct {
	prefix   string
	exchange string
}

func defaultOptions() options {
	return options{
		prefix:   msgPrefix,
		exchange: exchangeName,
	}
}

// Prefix sets the prefix for the publisher.
func Prefix(prefix string) messaging.Option {
	return func(val interface{}) error {
		switch v := val.(type) {
		case *publisher:
			v.prefix = prefix
		case *pubsub:
			v.prefix = prefix
		default:
			return ErrInvalidType
		}
		return nil
	}
}

// Exchange sets the exchange for the publisher or subscriber.
func Exchange(exchange string) messaging.Option {
	return func(val interface{}) error {
		switch v := val.(type) {
		case *publisher:
			v.exchange = exchange
		case *pubsub:
			v.exchange = exchange
		default:
			return ErrInvalidType
		}

		return nil
	}
}

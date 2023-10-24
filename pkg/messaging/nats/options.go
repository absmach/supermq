// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"errors"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/nats-io/nats.go/jetstream"
)

// ErrInvalidType is returned when the provided value is not of the expected type.
var ErrInvalidType = errors.New("invalid type")

// WithPrefix sets the prefix for the publisher.
func WithPrefix(prefix string) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*publisher)
		if !ok {
			return ErrInvalidType
		}

		p.prefix = prefix

		return nil
	}
}

// WithJSStream sets the JetStream for the publisher.
func WithJSStream(stream jetstream.JetStream) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*publisher)
		if !ok {
			return ErrInvalidType
		}

		p.js = stream

		return nil
	}
}

// WithStream sets the Stream for the subscriber.
func WithStream(stream jetstream.Stream) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*pubsub)
		if !ok {
			return ErrInvalidType
		}

		p.stream = stream

		return nil
	}
}

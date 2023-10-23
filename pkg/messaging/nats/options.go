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

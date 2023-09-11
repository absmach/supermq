// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !rabbitmq
// +build !rabbitmq

package brokers

import (
	"context"
	"errors"
	"log"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/nats-io/nats.go/jetstream"
)

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.>"

// ErrInvalidType is returned when the provided value is not of the expected type.
var ErrInvalidType = errors.New("invalid type")

func init() {
	log.Println("The binary was build using Nats as the message broker")
}

func NewPublisher(ctx context.Context, url string, opts ...messaging.Option) (messaging.Publisher, error) {
	pb, err := nats.NewPublisher(ctx, url, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func NewPubSub(ctx context.Context, url string, logger mflog.Logger, opts ...messaging.Option) (messaging.PubSub, error) {
	pb, err := nats.NewPubSub(ctx, url, logger, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func WithPrefix(prefix *string) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*nats.NPublisher)
		if !ok {
			return ErrInvalidType
		}

		p.Prefix = *prefix

		return nil
	}
}

func WithJSStream(stream jetstream.JetStream) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*nats.NPublisher)
		if !ok {
			return ErrInvalidType
		}

		p.JS = stream

		return nil
	}
}

func WithStream(stream jetstream.Stream) messaging.Option {
	return func(val interface{}) error {
		p, ok := val.(*nats.NPubsub)
		if !ok {
			return ErrInvalidType
		}

		p.Stream = stream

		return nil
	}
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !rabbitmq
// +build !rabbitmq

package brokers

import (
	"context"
	"log"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	broker "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	// SubjectAllChannels represents subject to subscribe for all the channels.
	SubjectAllChannels = "channels.>"

	maxReconnects = -1
)

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

func WithStream(ctx context.Context, cfg jetstream.StreamConfig) messaging.Option {
	return func(url, _ string) error {
		conn, err := broker.Connect(url, broker.MaxReconnects(maxReconnects))
		if err != nil {
			return err
		}
		js, err := jetstream.New(conn)
		if err != nil {
			return err
		}
		if _, err := js.CreateStream(ctx, cfg); err != nil {
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

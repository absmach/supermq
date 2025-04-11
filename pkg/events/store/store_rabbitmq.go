// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build es_rabbitmq
// +build es_rabbitmq

package store

import (
	"context"
	"log"
	"log/slog"

	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/events/rabbitmq"
)

// StreamAllEvents represents subject to subscribe for all the events.
const StreamAllEvents = "events.#"

func init() {
	log.Println("The binary was build using RabbitMQ as the events store")
}

func NewPublisher(ctx context.Context, url string) (events.Publisher, error) {
	pb, err := rabbitmq.NewPublisher(ctx, url)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func NewSubscriber(_ context.Context, url string, logger *slog.Logger) (events.Subscriber, error) {
	pb, err := rabbitmq.NewSubscriber(url, logger)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

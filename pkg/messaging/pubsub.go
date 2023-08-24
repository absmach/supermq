// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package messaging

import "context"

// Publisher specifies message publishing API.
type Publisher interface {
	// Publishes message to the stream.
	Publish(ctx context.Context, topic string, msg *Message) error

	// Close gracefully closes message publisher's connection.
	Close() error
}

// MessageHandler represents Message handler for Subscriber.
type MessageHandler interface {
	// Handle handles messages passed by underlying implementation.
	Handle(msg *Message) error

	// Cancel is used for cleanup during unsubscribing and it's optional.
	Cancel() error
}

type SubscriberConfig struct {
	ID      string
	Topic   string
	Handler MessageHandler
}

// Subscriber specifies message subscription API.
type Subscriber interface {
	// Subscribe subscribes to the message stream and consumes messages.
	Subscribe(ctx context.Context, cfg SubscriberConfig) error

	// Unsubscribe unsubscribes from the message stream and
	// stops consuming messages.
	Unsubscribe(ctx context.Context, cfg SubscriberConfig) error

	// Close gracefully closes message subscriber's connection.
	Close() error
}

// PubSub  represents aggregation interface for publisher and subscriber.
type PubSub interface {
	Publisher
	Subscriber
}

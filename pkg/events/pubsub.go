// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"time"
)

const (
	UnpublishedEventsCheckInterval        = 1 * time.Minute
	ConnCheckInterval                     = 100 * time.Millisecond
	MaxUnpublishedEvents           uint64 = 1e6
)

// Event represents event.
type Event interface {
	// Encode encodes event to map.
	Encode() (map[string]interface{}, error)
}

// Publisher specifies events publishing API.
type Publisher interface {
	// Publishes event to stream.
	Publish(ctx context.Context, event Event) error

	// StartPublishingRoutine starts routine that checks for unpublished events
	// and publishes them to stream.
	StartPublishingRoutine(ctx context.Context)

	// Close gracefully closes event publisher's connection.
	Close() error
}

// EventHandler represents event handler for Subscriber.
type EventHandler interface {
	// Handle handles events passed by underlying implementation.
	Handle(event Event) error

	// Cancel is used for cleanup during unsubscribing and it's optional.
	Cancel() error
}

// Subscriber specifies event subscription API.
type Subscriber interface {
	// Subscribe subscribes to the event stream and consumes events.
	Subscribe(ctx context.Context, handler EventHandler) error

	// Unsubscribe unsubscribes from the event stream and
	// stops consuming events.
	Unsubscribe(ctx context.Context) error

	// Close gracefully closes event subscriber's connection.
	Close() error
}

// PubSub  represents aggregation interface for publisher and subscriber.
type PubSub interface {
	Publisher
	Subscriber
}

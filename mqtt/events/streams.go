// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/events/store"
)

const streamID = "supermq.mqtt"

//go:generate mockery --name EventStore --output=../mocks --filename events.go --quiet --note "Copyright (c) Abstract Machines"
type EventStore interface {
	Connect(ctx context.Context, clientID, subscriberID string) error
	Disconnect(ctx context.Context, clientID, subscriberID string) error
	Subscribe(ctx context.Context, clientID, channelID, subscriberID, subtopic string) error
}

// EventStore is a struct used to store event streams in Redis.
type eventStore struct {
	ep       events.Publisher
	instance string
}

// NewEventStore returns wrapper around mProxy service that sends
// events to event store.
func NewEventStore(ctx context.Context, url, instance string) (EventStore, error) {
	publisher, err := store.NewPublisher(ctx, url, streamID)
	if err != nil {
		return nil, err
	}

	return &eventStore{
		instance: instance,
		ep:       publisher,
	}, nil
}

// Connect issues event on MQTT CONNECT.
func (es *eventStore) Connect(ctx context.Context, clientID, subscriberID string) error {
	ev := connectEvent{
		clientID:     clientID,
		operation:    clientConnect,
		subscriberID: subscriberID,
		instance:     es.instance,
	}

	return es.ep.Publish(ctx, ev)
}

// Disconnect issues event on MQTT CONNECT.
func (es *eventStore) Disconnect(ctx context.Context, clientID, subscriberID string) error {
	ev := connectEvent{
		clientID:     clientID,
		operation:    clientDisconnect,
		subscriberID: subscriberID,
		instance:     es.instance,
	}

	return es.ep.Publish(ctx, ev)
}

// Subscribe issues event on MQTT SUBSCRIBE.
func (es *eventStore) Subscribe(ctx context.Context, clientID, channelID, subscriberID, subtopic string) error {
	ev := subscribeEvent{
		operation:    clientSubscribe,
		clientID:     clientID,
		channelID:    channelID,
		subscriberID: subscriberID,
		subtopic:     subtopic,
	}

	return es.ep.Publish(ctx, ev)
}

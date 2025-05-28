// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/absmach/mgate/pkg/session"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/events/store"
	"github.com/absmach/supermq/pkg/messaging"
)

const (
	supermqPrefix    = "supermq."
	subscribeStream  = supermqPrefix + clientSubscribe
	connectStream    = supermqPrefix + clientConnect
	disconnectStream = supermqPrefix + clientDisconnect
)

var errFailedSession = errors.New("failed to obtain session from context")

// EventStore is a struct used to store event streams in Redis.
type eventStore struct {
	ep       events.Publisher
	handler  session.Handler
	instance string
}

// NewEventStoreMiddleware returns middleware around mGate service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, handler session.Handler, url, instance string) (session.Handler, error) {
	publisher, err := store.NewPublisher(ctx, url)
	if err != nil {
		return nil, err
	}

	return &eventStore{
		ep:       publisher,
		handler:  handler,
		instance: instance,
	}, nil
}

func (es *eventStore) AuthConnect(ctx context.Context) error {
	if err := es.handler.AuthConnect(ctx); err != nil {
		return err
	}
	s, ok := session.FromContext(ctx)
	if !ok {
		return errFailedSession
	}

	ev := connectEvent{
		operation:    clientConnect,
		clientID:     s.Username,
		subscriberID: s.ID,
		instance:     es.instance,
	}

	return es.ep.Publish(ctx, connectStream, ev)
}

func (es *eventStore) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	return es.handler.AuthPublish(ctx, topic, payload)
}

func (es *eventStore) AuthSubscribe(ctx context.Context, topics *[]string) error {
	return es.handler.AuthSubscribe(ctx, topics)
}

func (es *eventStore) Connect(ctx context.Context) error {
	return es.handler.Connect(ctx)
}

func (es *eventStore) Publish(ctx context.Context, topic *string, payload *[]byte) error {
	return es.handler.Publish(ctx, topic, payload)
}

func (es *eventStore) Subscribe(ctx context.Context, topics *[]string) error {
	if err := es.handler.Subscribe(ctx, topics); err != nil {
		return err
	}

	s, ok := session.FromContext(ctx)
	if !ok {
		return errFailedSession
	}

	for _, topic := range *topics {
		domainID, channelID, subTopic, err := messaging.ParseSubscribeTopic(topic)
		if err != nil {
			return err
		}
		ev := subscribeEvent{
			operation:    clientSubscribe,
			clientID:     s.Username,
			domainID:     domainID,
			channelID:    channelID,
			subtopic:     subTopic,
			subscriberID: s.ID,
		}

		if err := es.ep.Publish(ctx, subscribeStream, ev); err != nil {
			return err
		}
	}

	return nil
}

func (es *eventStore) Unsubscribe(ctx context.Context, topics *[]string) error {
	return es.handler.Unsubscribe(ctx, topics)
}

func (es *eventStore) Disconnect(ctx context.Context) error {
	if err := es.handler.Disconnect(ctx); err != nil {
		return err
	}

	s, ok := session.FromContext(ctx)
	if !ok {
		return errFailedSession
	}

	ev := connectEvent{
		operation:    clientDisconnect,
		clientID:     s.Username,
		subscriberID: s.ID,
		instance:     es.instance,
	}

	return es.ep.Publish(ctx, disconnectStream, ev)
}

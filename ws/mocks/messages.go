package mocks

import (
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
	broker "github.com/nats-io/go-nats"
)

var _ mainflux.MessagePubSub = (*mockPubSub)(nil)

type mockPubSub struct {
	subscriptions map[string]mockSubscription
}

// NewMessagePubSub returns mock message publisher.
func NewMessagePubSub() mainflux.MessagePubSub {
	return mockPubSub{map[string]mockSubscription{}}
}

func (pubsub mockPubSub) Publish(msg mainflux.RawMessage) error {
	if len(msg.Payload) == 0 {
		return broker.ErrInvalidMsg
	}
	return nil
}

func (pubsub mockPubSub) Subscribe(subscription mainflux.Subscription, write mainflux.WriteMessage, read mainflux.ReadMessage) (func(), error) {
	if _, ok := pubsub.subscriptions[subscription.ChanID]; ok {
		return nil, ws.ErrFailedSubscription
	}
	sub := mockSubscription{subscription.ChanID, write}
	pubsub.subscriptions[subscription.ChanID] = sub
	return func() {
		delete(pubsub.subscriptions, sub.channel)
	}, nil
}

type mockSubscription struct {
	channel string
	write   mainflux.WriteMessage
}

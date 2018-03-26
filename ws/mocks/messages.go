package mocks

import (
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
	broker "github.com/nats-io/go-nats"
)

var _ mainflux.MessagePubSub = (*mockPubSub)(nil)
var _ mainflux.Subscription = (*mockSubscription)(nil)

type mockPubSub struct {
	subscriptions map[string]mainflux.Subscription
}

// NewMessagePubSub returns mock message publisher.
func NewMessagePubSub() mainflux.MessagePubSub {
	return mockPubSub{map[string]mainflux.Subscription{}}
}

func (pubsub mockPubSub) Publish(msg mainflux.RawMessage) error {
	if len(msg.Payload) == 0 {
		return broker.ErrInvalidMsg
	}
	return nil
}

func (pubsub mockPubSub) Subscribe(channel string, onMessage func(mainflux.RawMessage)) (mainflux.Subscription, error) {
	if _, ok := pubsub.subscriptions[channel]; ok {
		return nil, ws.ErrFailedSubscription
	}
	sub := mockSubscription{
		pubsub:    pubsub,
		channel:   channel,
		onMessage: onMessage,
	}
	pubsub.subscriptions[channel] = sub
	return sub, nil
}

type mockSubscription struct {
	pubsub    mockPubSub
	channel   string
	onMessage func(mainflux.RawMessage)
}

func (sub mockSubscription) Unsubscribe() error {
	delete(sub.pubsub.subscriptions, sub.channel)
	return nil
}

package mocks

import (
	"github.com/mainflux/mainflux"
	broker "github.com/nats-io/go-nats"
)

var _ mainflux.MessagePubSub = (*mockPubSub)(nil)

type mockPubSub struct {
	messages []mainflux.RawMessage
}

// NewMessagePubSub returns mock message publisher.
func NewMessagePubSub() mainflux.MessagePubSub {
	return mockPubSub{}
}

func (pubsub mockPubSub) Publish(msg mainflux.RawMessage) error {
	if len(msg.Payload) == 0 {
		return broker.ErrInvalidMsg
	}
	return nil
}

func (pubsub mockPubSub) Subscribe(channel string, onMessage func(mainflux.RawMessage)) (mainflux.Subscription, error) {
	return nil, nil
}

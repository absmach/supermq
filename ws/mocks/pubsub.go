package mocks

import (
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
)

type mockPubSub struct {
	fail bool
}

// NewPubSub returns mock message publisher-subscriber
func NewPubSub() messaging.PubSub {
	return mockPubSub{}
}
func (pubsub mockPubSub) Publish(string, messaging.Message) error {
	if pubsub.fail {
		return ws.ErrFailedMessagePublish
	}
	return nil
}

func (pubsub mockPubSub) Subscribe(string, string, messaging.MessageHandler) error {
	if pubsub.fail {
		return ws.ErrFailedSubscription
	}
	return nil
}

func (pubsub mockPubSub) Unsubscribe(string, string) error {
	if pubsub.fail {
		return ws.ErrFailedUnsubscribe
	}
	return nil
}

func (pb mockPubSub) Close() error {
	return nil
}

package mocks

import "github.com/mainflux/mainflux/pkg/messaging"

type mockPubSub struct{}

// NewPubSub returns mock message publisher-subscriber
func NewPubSub() messaging.PubSub {
	return mockPubSub{}
}

func (pubsub mockPubSub) Publish(string, messaging.Message) error {
	return nil
}

func (pubsub mockPubSub) Subscribe(string, string, messaging.MessageHandler) error {
	return nil
}

func (pubsub mockPubSub) Unsubscribe(string, string) error {
	return nil
}

func (pubsub mockPubSub) Close() error {
	return nil
}

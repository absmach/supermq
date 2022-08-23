// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
)

var _ messaging.PubSub = (*mockPubSub)(nil)

type MockPubSub interface {
	Publish(string, messaging.Message) error
	Subscribe(string, string, messaging.MessageHandler) error
	Unsubscribe(string, string) error
	SetFail(bool)
	Close() error
}

type mockPubSub struct {
	fail bool
}

// NewPubSub returns mock message publisher-subscriber
func NewPubSub() MockPubSub {
	return &mockPubSub{false}
}
func (pubsub *mockPubSub) Publish(string, messaging.Message) error {
	if pubsub.fail {
		return ws.ErrFailedMessagePublish
	}
	return nil
}

func (pubsub *mockPubSub) Subscribe(string, string, messaging.MessageHandler) error {
	if pubsub.fail {
		return ws.ErrFailedSubscription
	}
	return nil
}

func (pubsub *mockPubSub) Unsubscribe(string, string) error {
	if pubsub.fail {
		return ws.ErrFailedUnsubscribe
	}
	return nil
}

func (pubsub *mockPubSub) SetFail(fail bool) {
	pubsub.fail = fail
}

func (pubsub *mockPubSub) Close() error {
	return nil
}

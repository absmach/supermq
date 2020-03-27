// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/broker"
	"github.com/mainflux/mainflux/ws"
	"github.com/nats-io/nats.go"
)

var _ broker.Nats = (*mockPubSub)(nil)

type mockPubSub struct {
	subscriptions map[string]*ws.Channel
}

// New returns mock message publisher.
func New(sub map[string]*ws.Channel) broker.Nats {
	return &mockPubSub{
		subscriptions: sub,
	}
}

func (mp mockPubSub) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	if len(msg.Payload) == 0 {
		return ws.ErrFailedMessagePublish
	}
	return nil
}

func (mp mockPubSub) Subscribe(chanID, subtopic string, f func(*nats.Msg)) (*nats.Subscription, error) {
	if _, ok := mp.subscriptions[chanID+subtopic]; !ok {
		return nil, ws.ErrFailedSubscription
	}

	return nil, nil
}

func (mp mockPubSub) Close() {
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux"
	broker "github.com/mainflux/mainflux/broker/nats"
	"github.com/mainflux/mainflux/ws"
	"github.com/nats-io/nats.go"
)

var _ broker.Publisher = (*mockPub)(nil)
var _ broker.Subscriber = (*mockSub)(nil)

type mockPub struct {
}

type mockSub struct {
	subscriptions map[string]*ws.Channel
}

// NewPublisher returns mock message publisher.
func NewPublisher() broker.Publisher {
	return &mockPub{}
}

// NewSubscriber returns mock message publisher.
func NewSubscriber(subs map[string]*ws.Channel) broker.Subscriber {
	return &mockSub{
		subscriptions: subs,
	}
}

func (mp mockPub) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	if len(msg.Payload) == 0 {
		return ws.ErrFailedMessagePublish
	}
	return nil
}

func (mp mockPub) Conn() *nats.Conn {
	return nil
}

func (mp mockSub) Subscribe(chanID, subtopic string, f func(*nats.Msg)) (*nats.Subscription, error) {
	if _, ok := mp.subscriptions[chanID+subtopic]; !ok {
		return nil, ws.ErrFailedSubscription
	}

	return nil, nil
}

func (mp mockSub) Conn() *nats.Conn {
	return nil
}

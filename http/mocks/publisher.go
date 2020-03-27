// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/broker"
	"github.com/nats-io/nats.go"
)

type mockPublisher struct{}

// NewPublisher returns mock message publisher.
func NewPublisher() broker.Broker {
	return mockPublisher{}
}

func (pub mockPublisher) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	return nil
}

func (pub mockPublisher) Subscribe(chanID, subtopic string, f func(*nats.Msg)) (*nats.Subscription, error) {
	return nil, nil
}

func (pub mockPublisher) Close() {
}

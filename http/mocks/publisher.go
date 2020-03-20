// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/brokers"
	"github.com/nats-io/nats.go"
)

var _ (brokers.NatsPublisher) = (*mockPublisher)(nil)

type mockPublisher struct{}

// NewPublisher returns mock message publisher.
func NewPublisher() brokers.NatsPublisher {
	return mockPublisher{}
}

func (pub mockPublisher) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	return nil
}

func (pub mockPublisher) PubConn() *nats.Conn {
	return nil
}

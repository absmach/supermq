// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux"
	broker "github.com/mainflux/mainflux/broker/nats"
	"github.com/nats-io/nats.go"
)

var _ (broker.Publisher) = (*mockPublisher)(nil)

type mockPublisher struct{}

// NewPublisher returns mock message publisher.
func NewPublisher() broker.Publisher {
	return mockPublisher{}
}

func (pub mockPublisher) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	return nil
}

func (pub mockPublisher) Conn() *nats.Conn {
	return nil
}

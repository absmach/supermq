// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/brokers"
)

var _ (brokers.MessagePublisher) = (*mockPublisher)(nil)

type mockPublisher struct{}

// NewPublisher returns mock message publisher.
func NewPublisher() brokers.MessagePublisher {
	return mockPublisher{}
}

func (pub mockPublisher) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	return nil
}

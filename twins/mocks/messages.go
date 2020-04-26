// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/errors"
)

var _ mainflux.Publisher = (*mockBroker)(nil)

type mockBroker struct {
	subscriptions map[string]string
}

// New returns mock message publisher.
func New(sub map[string]string) mainflux.Publisher {
	return &mockBroker{
		subscriptions: sub,
	}
}

func (mb mockBroker) Publish(topic string, msg mainflux.Message) error {
	if len(msg.Payload) == 0 {
		return errors.New("failed to publish")
	}
	return nil
}

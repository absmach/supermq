// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/mainflux/mainflux"
)

type mockPublisher struct{}

// NewPublisher returns mock message publisher.
func NewPublisher() mainflux.Publisher {
	return mockPublisher{}
}

func (pub mockPublisher) Publish(mainflux.Message) error {
	return nil
}

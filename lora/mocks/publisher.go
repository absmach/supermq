// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
)

type mockPublisher struct{}

// NewPublisher returns mock message publisher.
func NewPublisher() messaging.Publisher {
	return mockPublisher{}
}

func (pub mockPublisher) Publish(topic string, msg *messaging.Message, spanContext opentracing.SpanContext) error {
	return nil
}

func (pub mockPublisher) Close() error {
	return nil
}

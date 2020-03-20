// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package brokers

import (
	context "context"

	"github.com/mainflux/mainflux"
	"github.com/nats-io/nats.go"
)

// MessagePublisher specifies a message publishing API.
type MessagePublisher interface {
	// Publish publishes message to the msessage broker.
	Publish(context.Context, string, mainflux.Message) error
}

// MessageSubscriber specifies a message subscribing API.
type MessageSubscriber interface {
	// Subscribe subscribes to the message broker for a given channel ID and subtopic.
	Subscribe(string, string, func(msg *nats.Msg)) (*nats.Subscription, error)
}

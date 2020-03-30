// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package broker

import (
	"context"

	"github.com/nats-io/nats.go"
)

// SubjectAllChannels allows to subscribe to all subjects of all channels
const SubjectAllChannels = "channel.>"

// Nats specifies a NATS message API.
type Nats interface {
	// Publish publishes message to the msessage broker.
	Publish(context.Context, string, Message) error

	// Subscribe subscribes to the message broker for a given channel ID and subtopic.
	Subscribe(string, string, func(msg *nats.Msg)) (*nats.Subscription, error)

	// Close closes NATS connection.
	Close()
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"strings"

	"github.com/mainflux/mainflux/messaging"
)

const (
	channels = "channels"
	messages = "messages"
)

// Forwarder specifies MQTT forwarder interface API.
type Forwarder interface {
	// Forward subscribes to the Subscriber and
	// publishes messages using provided Publisher.
	Forward(sub messaging.Subscriber, pub messaging.Publisher) error
}

type forwarder struct {
	topic string
}

// NewForwarder returns new Forwarder implementation.
func NewForwarder(topic string) Forwarder {
	return forwarder{topic}
}

func (f forwarder) Forward(sub messaging.Subscriber, pub messaging.Publisher) error {
	return sub.Subscribe(f.topic, handle(pub))
}

func handle(pub messaging.Publisher) messaging.MessageHandler {
	return func(msg messaging.Message) error {
		if msg.Protocol == protocol {
			return nil
		}
		// Use concatenation instead of mft.Sprintf for the
		// sake of simplicity and performance.
		topic := channels + "/" + msg.Channel + "/" + messages
		if msg.Subtopic != "" {
			topic += "/" + strings.ReplaceAll(msg.Subtopic, ".", "/")
		}
		return pub.Publish(topic, msg)
	}
}

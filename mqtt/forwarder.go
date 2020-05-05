// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/messaging"
)

const (
	channels = "channels"
	messages = "messages"
)

// Forward subscribes to the Subscriber and publishes all
// the messages from the given topic using provided Publisher.
func Forward(topic string, sub messaging.Subscriber, pub messaging.Publisher) error {
	return sub.Subscribe(topic, handle(pub))
}

func handle(pub messaging.Publisher) messaging.MessageHandler {
	return func(msg messaging.Message) error {
		if msg.Protocol == protocol {
			return nil
		}
		topic := fmt.Sprintf("%s.%s.%s.%s", channels, msg.Channel, messages, msg.Subtopic)
		topic = strings.ReplaceAll(topic, ".", "/")
		return pub.Publish(topic, msg)
	}
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/messaging"
)

const (
	channels = "channels"
	messages = "messages"
	id       = "mqtt-adapter"
)

type forwarder struct {
	client  mqtt.Client
	timeout time.Duration
}

// NewForwarder returns a new MQTT message forwarder.
func NewForwarder(publisher messaging.Publisher) messaging.MessageHandler {
	return func(msg messaging.Message) error {
		topic := fmt.Sprintf("%s.%s.%s.%s", channels, msg.Channel, messages, msg.Subtopic)
		topic = strings.ReplaceAll(topic, ".", "/")
		return publisher.Publish(topic, msg)
	}
}

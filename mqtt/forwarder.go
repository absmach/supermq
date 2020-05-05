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
func NewForwarder(address string, timeout time.Duration) (messaging.MessageHandler, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(address).
		SetUsername(id).
		SetPassword(id).
		SetClientID(id).
		SetCleanSession(false)
	client := mqtt.NewClient(opts)
	tkn := client.Connect()
	if tkn.Wait() && tkn.Error() != nil {
		return nil, tkn.Error()
	}
	f := forwarder{
		client:  client,
		timeout: timeout,
	}
	return f.Forward, nil
}

func (f forwarder) Forward(msg messaging.Message) error {
	if msg.Protocol == protocol {
		return nil
	}
	topic := fmt.Sprintf("%s.%s.%s.%s", channels, msg.Channel, messages, msg.Subtopic)
	topic = strings.ReplaceAll(topic, ".", "/")
	tkn := f.client.Publish(topic, 1, false, msg.Payload)
	if tkn.WaitTimeout(f.timeout) && tkn.Error() != nil {
		return tkn.Error()
	}
	return nil
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/messaging"
)

var errPublish = errors.New("unable to publish the message to the MQTT broker")

type forwarder struct {
	client  mqtt.Client
	timeout time.Duration
}

func NewForwarder(address string) (messaging.MessageHandler, error) {
	id := "mqtt-adapter"
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://127.0.0.1:1884").
		SetUsername(id).
		SetPassword(id).
		SetClientID(id).
		SetCleanSession(false)
	client := mqtt.NewClient(opts)
	tkn := client.Connect()
	if tkn.Wait() && tkn.Error() != nil {
		return nil, tkn.Error()
	}
	f := forwarder{client: client, timeout: time.Second * 30}
	return f.Forward, nil
}

func (f forwarder) Forward(msg messaging.Message) error {
	if msg.Protocol == protocol {
		return nil
	}
	fmt.Println("Publishing...", msg.Protocol, msg.Channel, msg.Subtopic)
	topic := strings.ReplaceAll("channels."+msg.Channel+".messages."+msg.Subtopic, ".", "/")
	fmt.Println(topic)
	tkn := f.client.Publish(topic, 2, false, msg.Payload)
	if tkn.WaitTimeout(f.timeout) && tkn.Error() != nil {
		return tkn.Error()
	}
	return nil
}

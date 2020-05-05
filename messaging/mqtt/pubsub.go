// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"errors"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	channels = "channels"
	messages = "messages"
	protocol = "mqtt"
	id       = "mqtt-adapter"
)

var errConnect = errors.New("unable to connect to MQTT broker")

func newClient(address string, timeout time.Duration) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(address).
		SetUsername(id).
		SetPassword(id).
		SetClientID(id).
		SetCleanSession(false)
	client := mqtt.NewClient(opts)
	tkn := client.Connect()
	to := tkn.WaitTimeout(timeout)
	if to && tkn.Error() != nil {
		return nil, tkn.Error()
	}
	if !to {
		return nil, errConnect
	}
	return client, nil
}

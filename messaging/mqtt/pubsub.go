// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"errors"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	protocol = "mqtt"
	id       = "mqtt-publisher"
	qos      = 1
)

var errConnect = errors.New("failed to connect to MQTT broker")

func newClient(address string, timeout time.Duration) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(address).
		SetUsername(id).
		SetPassword(id).
		SetClientID(id).
		SetCleanSession(false)
	client := mqtt.NewClient(opts)
	tkn := client.Connect()
	ok := tkn.WaitTimeout(timeout)
	if ok && tkn.Error() != nil {
		return nil, tkn.Error()
	}
	if !ok {
		return nil, errConnect
	}

	return client, nil
}

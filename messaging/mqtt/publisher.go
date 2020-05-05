// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"errors"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/messaging"
)

var _ messaging.Publisher = (*publisher)(nil)

var errPublish = errors.New("failed to publish")

type publisher struct {
	client  mqtt.Client
	timeout time.Duration
}

// NewPublisher returns a new MQTT message publisher.
func NewPublisher(address string, timeout time.Duration) (messaging.Publisher, error) {
	client, err := newClient(address, timeout)
	if err != nil {
		return nil, err
	}

	ret := publisher{
		client:  client,
		timeout: timeout,
	}
	return ret, nil
}

func (pub publisher) Publish(topic string, msg messaging.Message) error {
	if msg.Protocol == protocol {
		return nil
	}
	tkn := pub.client.Publish(topic, 1, false, msg.Payload)
	ok := tkn.WaitTimeout(pub.timeout)
	if ok && tkn.Error() != nil {
		return tkn.Error()
	}
	if !ok {
		return errPublish
	}

	return nil
}

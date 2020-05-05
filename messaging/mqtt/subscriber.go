// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"fmt"

	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/messaging"
)

var _ messaging.Publisher = (*publisher)(nil)

type subscriber struct {
	client  mqtt.Client
	timeout time.Duration
	logger  log.Logger
}

// NewSubscriber returns a new MQTT message subscriber.
func NewSubscriber(address string, timeout time.Duration, logger log.Logger) (messaging.Subscriber, error) {
	client, err := newClient(address, timeout)
	if err != nil {
		return nil, err
	}

	ret := subscriber{
		client:  client,
		timeout: timeout,
		logger:  logger,
	}
	return ret, nil
}

func (sub subscriber) Subscribe(topic string, handler messaging.MessageHandler) error {
	tkn := sub.client.Subscribe(topic, 2, sub.mqttHandler(handler))
	if tkn.WaitTimeout(sub.timeout) && tkn.Error() != nil {
		return tkn.Error()
	}
	return nil
}

func (sub subscriber) Unsubscribe(topic string) error {
	tkn := sub.client.Unsubscribe(topic)
	if tkn.WaitTimeout(sub.timeout) && tkn.Error() != nil {
		return tkn.Error()
	}
	return nil
}

func (sub subscriber) mqttHandler(h messaging.MessageHandler) mqtt.MessageHandler {
	return func(c mqtt.Client, m mqtt.Message) {
		var msg messaging.Message
		if err := proto.Unmarshal(m.Payload(), &msg); err != nil {
			sub.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h(msg); err != nil {
			sub.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
		}
	}
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"errors"
	"time"

	"github.com/absmach/supermq/pkg/messaging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var errPublishTimeout = errors.New("failed to publish due to timeout reached")

var _ messaging.Publisher = (*publisher)(nil)

type publisher struct {
	client  mqtt.Client
	timeout time.Duration
}

const (
	qosKey      = "qos"
	publisherID = "smq-mqtt-publisher"
)

// NewPublisher returns a new MQTT message publisher.
func NewPublisher(address, username, password string, timeout time.Duration) (messaging.Publisher, error) {
	client, err := newClient(address, username, password, publisherID, timeout)
	if err != nil {
		return nil, err
	}

	ret := publisher{
		client:  client,
		timeout: timeout,
	}
	return ret, nil
}

func (pub publisher) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	if topic == "" {
		return ErrEmptyTopic
	}

	// Publish only the payload and not the whole message.
	token := pub.client.Publish(topic, qos(ctx), false, msg.GetPayload())
	if token.Error() != nil {
		return token.Error()
	}

	if ok := token.WaitTimeout(pub.timeout); !ok {
		return errPublishTimeout
	}

	return nil
}

func (pub publisher) Close() error {
	pub.client.Disconnect(uint(pub.timeout))
	return nil
}

func WithQoS(ctx context.Context, qos byte) context.Context {
	return context.WithValue(ctx, qosKey, qos)
}

// Return QoS defaulting to 1 for compatibility reasons.
func qos(ctx context.Context) byte {
	val := ctx.Value(qosKey)
	if val != nil {
		if ret, ok := val.(byte); ok {
			return ret
		}
	}
	return 1
}

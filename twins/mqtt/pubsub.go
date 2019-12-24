// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/logger"
)

// Mqtt stores mqtt client and topic
type Mqtt struct {
	client mqtt.Client
	topic  string
}

// New instantiates the mqtt service.
func New(mc mqtt.Client, topic string) Mqtt {
	return Mqtt{
		client: mc,
		topic:  topic,
	}
}

// Connect to MQTT broker
func Connect(mqttURL, id, key string, logger logger.Logger) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttURL)
	opts.SetClientID("twins")
	opts.SetUsername(id)
	opts.SetPassword(key)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		logger.Info("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		logger.Error(fmt.Sprintf("MQTT connection lost: %s", err.Error()))
		os.Exit(1)
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Error(fmt.Sprintf("Failed to connect to MQTT broker: %s", token.Error()))
		os.Exit(1)
	}

	return client
}

func (mqtt *Mqtt) Topic() string {
	return mqtt.topic
}

func (mqtt *Mqtt) publish(id, op string, payload *[]byte) error {
	topic := fmt.Sprintf("channels/%s/messages/%s/%s", mqtt.topic, id, op)
	if len(id) < 1 {
		topic = fmt.Sprintf("channels/%s/messages/%s", mqtt.topic, op)
	}

	token := mqtt.client.Publish(topic, 0, false, *payload)
	token.Wait()

	return token.Error()
}

// Publish sends mqtt message to a predefined topic
func (mqtt *Mqtt) Publish(id *string, err *error, succOp, failOp string, payload *[]byte) error {
	op := succOp
	if *err != nil {
		op = failOp
		esb := []byte((*err).Error())
		payload = &esb
	}

	mqttErr := mqtt.publish(*id, op, payload)
	if mqttErr != nil {
		return mqttErr
	}

	return nil
}

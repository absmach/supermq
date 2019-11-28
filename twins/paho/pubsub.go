package paho

import (
	"fmt"
	"os"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/logger"
)

// Mqtt stores mqtt client and topic
type Mqtt struct {
	client paho.Client
	topic  string
}

// New instantiates the paho service.
func New(mc paho.Client, topic string) Mqtt {
	return Mqtt{
		client: mc,
		topic:  topic,
	}
}

// Connect to MQTT broker
func Connect(mqttURL, id, key string, logger logger.Logger) paho.Client {
	opts := paho.NewClientOptions()
	opts.AddBroker(mqttURL)
	opts.SetClientID("twins")
	opts.SetUsername(id)
	opts.SetPassword(key)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(c paho.Client) {
		logger.Info("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(c paho.Client, err error) {
		logger.Error(fmt.Sprintf("MQTT connection lost: %s", err.Error()))
		os.Exit(1)
	})

	client := paho.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logger.Error(fmt.Sprintf("Failed to connect to MQTT broker: %s", token.Error()))
		os.Exit(1)
	}

	return client
}

// Publish sends mqtt message to a predefined topic
func (mqtt *Mqtt) Publish(id, op string, payload *[]byte) error {
	topic := fmt.Sprintf("channels/%s/messages/%s/%s", mqtt.topic, id, op)
	if len(id) < 1 {
		topic = fmt.Sprintf("channels/%s/messages/%s", mqtt.topic, op)
	}

	token := mqtt.client.Publish(topic, 0, false, *payload)
	token.Wait()

	return token.Error()
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/mqtt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic = "topic"
	// chansPrefix = "channels"
	channel  = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic = "engine"
	clientID = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
)

var (
	msgChan = make(chan messaging.Message)
	data    = []byte("payload")
)

// Tests only the publisher
func TestPublisher(t *testing.T) {
	// Subscribing with topic, and with subtopic, so that we can publish messages
	client, err := newClient(address, "", 30*time.Second)

	token := client.Subscribe(topic, qos, mqttHandler(handler{false}))
	// token := client.Subscribe(fmt.Sprintf("%s.%s", chansPrefix, topic), qos, mqttHandler(handler{false}))
	require.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", err))
	token = client.Subscribe(fmt.Sprintf("%s.%s", topic, subtopic), qos, mqttHandler(handler{false}))
	// token = client.Subscribe(fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic), qos, mqttHandler(handler{false}))
	require.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc     string
		channel  string
		subtopic string
		payload  []byte
	}{
		{
			desc:    "publish message with nil payload",
			payload: nil,
		},
		{
			desc:    "publish message with string payload",
			payload: data,
		},
		{
			desc:    "publish message with channel",
			payload: data,
			channel: channel,
		},
		{
			desc:     "publish message with subtopic",
			payload:  data,
			subtopic: subtopic,
		},
		{
			desc:     "publish message with channel and subtopic",
			payload:  data,
			channel:  channel,
			subtopic: subtopic,
		},
	}

	for _, tc := range cases {
		expectedMsg := messaging.Message{
			Channel:  tc.channel,
			Subtopic: tc.subtopic,
			Payload:  tc.payload,
		}
		err := publisher.Publish(topic, expectedMsg)
		assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s", tc.desc, err))

		receivedMsg := <-msgChan
		fmt.Printf("%s: received payload: %s\n\n", tc.desc, receivedMsg.Payload)
		assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
	}
}

// Tests only the Subscriber
func TestSubscribe(t *testing.T) {
	client, err := newClient(address, "id", 30*time.Second)
	require.Nil(t, err, fmt.Sprintf("got unexpected error while creating client: %s", err.Error()))

	cases := []struct {
		desc         string
		topic        string
		clientID     string
		errorMessage error
		handler      messaging.MessageHandler
	}{
		{
			desc:         "Subscribe to a topic with an ID",
			topic:        topic,
			clientID:     "clientid1",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to the same topic with a different ID",
			topic:        topic,
			clientID:     "clientid2",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:  "Subscribe to an already subscribed topic with an ID",
			topic: topic,
			// topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid1",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to the same topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid1",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid1",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to an empty topic with an ID",
			topic:        "",
			clientID:     "clientid1",
			errorMessage: mqtt.ErrEmptyTopic,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to a topic with empty id",
			topic:        topic,
			clientID:     "",
			errorMessage: mqtt.ErrEmptyID,
			handler:      handler{},
		},
	}

	for _, pc := range cases {
		t.Run(pc.desc, func(t *testing.T) {
			err := pubsub.Subscribe(pc.clientID, pc.topic, pc.handler)
			switch err {
			case nil:
				// if no error, publish message, and receive after subscribing
				expectedMsg := messaging.Message{
					Channel:  channel,
					Subtopic: subtopic,
					Payload:  data,
				}
				client.Publish(topic, qos, false, expectedMsg)

				receivedMsg := <-msgChan
				assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", pc.desc, expectedMsg, receivedMsg))
			default:
				switch pc.errorMessage {
				case nil:
					t.Error("got unexpected error: ", err.Error())
				default:
					assert.Equal(t, err.Error(), pc.errorMessage)
				}
			}
		})
	}
}

// Tests only the unsubscriber
func TestUnsubscribe(t *testing.T) {
	client, err := newClient(address, "", 30*time.Second)
	require.Nil(t, err, fmt.Sprintf("got unexpected error while creating client: %s", err.Error()))

	subs := []struct {
		desc         string
		topic        string
		clientID     string
		errorMessage error
		handler      messaging.MessageHandler
	}{
		{
			desc:         "Subscribe to a topic with an ID",
			topic:        topic,
			clientID:     "clientid3",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to the same topic with a different ID",
			topic:        topic,
			clientID:     "clientid4",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid3",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to an already subscribed topic with a subtopic with a different ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid4",
			errorMessage: nil,
			handler:      handler{},
		},
		{
			desc:         "Subscribe to a topic with an ID set to fail",
			topic:        fmt.Sprintf("%s%s", topic, "2"),
			clientID:     "clientid5",
			errorMessage: nil,
			handler:      handler{true},
		},
		{
			desc:         "Subscribe to a topic with a subtopic with an ID set to fail",
			topic:        fmt.Sprintf("%s%s.%s", topic, "3", subtopic),
			clientID:     "clientid5",
			errorMessage: nil,
			handler:      handler{true},
		},
	}

	for _, sub := range subs {
		token := client.Subscribe(sub.topic, qos, mqttHandler(handler{false}))
		require.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", err))
	}

	cases := []struct {
		desc         string
		topic        string
		clientID     string
		errorMessage error
	}{
		{
			desc:         "Unsubscribe from a topic with an ID",
			topic:        topic,
			clientID:     "clientid3",
			errorMessage: nil,
		},
		{
			desc:         "Unsubscribe from a non-existent topic with an ID",
			topic:        "h",
			clientID:     "clientid3",
			errorMessage: mqtt.ErrNotSubscribed,
		},
		{
			desc:         "Unsubscribe from the same topic with a different ID not subscribed",
			topic:        topic,
			clientID:     "clientidd",
			errorMessage: mqtt.ErrNotSubscribed,
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with an ID",
			topic:        topic,
			clientID:     "clientid3",
			errorMessage: mqtt.ErrNotSubscribed,
		},
		{
			desc:         "Unsubscribe from a topic with a subtopic with an ID",
			topic:        topic,
			clientID:     "clientid3",
			errorMessage: nil,
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid3",
			errorMessage: mqtt.ErrNotSubscribed,
		},
		{
			desc:         "Unsubscribe from an empty topic with an ID",
			topic:        "",
			clientID:     "clientid3",
			errorMessage: mqtt.ErrEmptyTopic,
		},
		{
			desc:         "Unsubscribe from a topic with empty id",
			topic:        topic,
			clientID:     "",
			errorMessage: mqtt.ErrEmptyID,
		},
		{
			desc:         "Unsubscribe from a topic with an ID with failing handler",
			topic:        fmt.Sprintf("%s%s", topic, "2"),
			clientID:     "clientid5",
			errorMessage: mqtt.ErrFailed,
		},
		{
			desc:         "Unsubscribe from a topic with an ID with failing handler",
			topic:        fmt.Sprintf("%s%s.%s", topic, "2", subtopic),
			clientID:     "clientid5",
			errorMessage: mqtt.ErrFailed,
		},
	}

	for _, pc := range cases {
		t.Run(pc.desc, func(t *testing.T) {
			err := pubsub.Unsubscribe(pc.clientID, pc.topic)

			switch err {
			case nil:
				if pc.errorMessage != nil {
					t.Errorf("expected an error: %s, did not receive any", pc.errorMessage.Error())
				}
			default:
				switch pc.errorMessage {
				case nil:
					t.Error("got unexpected error: ", err.Error())
				default:
					assert.Equal(t, err.Error(), pc.errorMessage)
				}
			}
		})
	}
}

// Tests both publisher and subscriber
func TestPubSub(t *testing.T) {
	// Subscribing with topic, and with subtopic, so that we can publish messages
	err := pubsub.Subscribe(clientID, topic, handler{false})
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	err = pubsub.Subscribe(clientID, fmt.Sprintf("%s.%s", topic, subtopic), handler{false})
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc     string
		channel  string
		subtopic string
		payload  []byte
	}{
		{
			desc:    "publish message with nil payload",
			payload: nil,
		},
		{
			desc:    "publish message with string payload",
			payload: data,
		},
		{
			desc:    "publish message with channel",
			payload: data,
			channel: channel,
		},
		{
			desc:     "publish message with subtopic",
			payload:  data,
			subtopic: subtopic,
		},
		{
			desc:     "publish message with channel and subtopic",
			payload:  data,
			channel:  channel,
			subtopic: subtopic,
		},
	}

	for _, tc := range cases {
		expectedMsg := messaging.Message{
			Channel:  tc.channel,
			Subtopic: tc.subtopic,
			Payload:  tc.payload,
		}
		err := publisher.Publish(topic, expectedMsg)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		receivedMsg := <-msgChan
		assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
	}
}

// Tests both Subscribe and Unsubscribe
func TestSubUnsub(t *testing.T) {
	client, err := newClient(address, "", 30*time.Second)
	require.Nil(t, err, fmt.Sprintf("got unexpected error while creating client: %s", err.Error()))
	// Test Subscribe and Unsubscribe
	cases := []struct {
		desc         string
		topic        string
		clientID     string
		errorMessage error
		pubsub       bool // true for subscribe and false for unsubscribe
		handler      messaging.MessageHandler
	}{
		{
			desc:         "Subscribe to a topic with an ID",
			topic:        topic,
			clientID:     "clientid7",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to the same topic with a different ID",
			topic:        topic,
			clientID:     "clientid8",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an already subscribed topic with an ID",
			topic:        topic,
			clientID:     "clientid7",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with an ID",
			topic:        topic,
			clientID:     "clientid7",
			errorMessage: nil,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a non-existent topic with an ID",
			topic:        "h",
			clientID:     "clientid7",
			errorMessage: mqtt.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from the same topic with a different ID",
			topic:        topic,
			clientID:     "clientid8",
			errorMessage: mqtt.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from the same topic with a different ID not subscribed",
			topic:        topic,
			clientID:     "clientid9",
			errorMessage: mqtt.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with an ID",
			topic:        topic,
			clientID:     "clientid7",
			errorMessage: mqtt.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid7",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid7",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid7",
			errorMessage: nil,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s", topic, subtopic),
			clientID:     "clientid7",
			errorMessage: mqtt.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an empty topic with an ID",
			topic:        "",
			clientID:     "clientid7",
			errorMessage: mqtt.ErrEmptyTopic,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an empty topic with an ID",
			topic:        "",
			clientID:     "clientid7",
			errorMessage: mqtt.ErrEmptyTopic,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to a topic with empty id",
			topic:        topic,
			clientID:     "",
			errorMessage: mqtt.ErrEmptyID,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with empty id",
			topic:        topic,
			clientID:     "",
			errorMessage: mqtt.ErrEmptyID,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to another topic with an ID",
			topic:        fmt.Sprintf("%s%s", topic, "1"),
			clientID:     "clientid9",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Subscribe to another already subscribed topic with an ID with Unsubscribe failing",
			topic:        fmt.Sprintf("%s%s", topic, "1"),
			clientID:     "clientid9",
			errorMessage: mqtt.ErrFailed,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Subscribe to a new topic with an ID",
			topic:        fmt.Sprintf("%s%s", topic, "2"),
			clientID:     "clientid10",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Unsubscribe from a topic with an ID with failing handler",
			topic:        fmt.Sprintf("%s%s", topic, "2"),
			clientID:     "clientid10",
			errorMessage: mqtt.ErrFailed,
			pubsub:       false,
			handler:      handler{true},
		},
	}

	for _, pc := range cases {

		t.Run(pc.desc, func(t *testing.T) {
			switch pc.pubsub {
			case true:
				// err := pubsub.Subscribe(pc.clientID, pc.topic, pc.handler)
				token := client.Subscribe(pc.topic, qos, mqttHandler(pc.handler))

				if pc.errorMessage == nil && token.Error() != nil {
					t.Error("got unexpected error: ", err.Error())
				} else if err.Error() != pc.errorMessage.Error() {
					t.Errorf("expected %s, got %s", pc.errorMessage.Error(), err.Error())
				}
			default:
				// err := pubsub.Unsubscribe(pc.clientID, pc.topic)
				token := client.Unsubscribe(pc.topic)

				if pc.errorMessage == nil && token.Error() != nil {
					t.Error("got unexpected error: ", err.Error())
				} else if err.Error() != pc.errorMessage.Error() {
					t.Errorf("expected %s, got %s", pc.errorMessage.Error(), err.Error())
				}
				if pc.errorMessage == nil {
					require.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", pc.desc, err))
				} else {
					assert.Equal(t, err, pc.errorMessage)
				}
			}
		})
	}
}

type handler struct {
	fail bool
}

func (h handler) Handle(msg messaging.Message) error {
	msgChan <- msg
	return nil
}

func (h handler) Cancel() error {
	if h.fail {
		return mqtt.ErrFailed
	}
	return nil
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/pkg/messaging"
	mqtt_pubsub "github.com/mainflux/mainflux/pkg/messaging/mqtt"
	"github.com/stretchr/testify/assert"
)

const (
	topic       = "topic"
	chansPrefix = "channels"
	channel     = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic    = "engine"
)

var (
	// msgChan = make(chan messaging.Message)
	data = []byte("payload")
)

func TestPublisher(t *testing.T) {
	msgChan := make(chan messaging.Message)

	// Subscribing with topic, and with subtopic, so that we can publish messages
	client, err := newClient(address, "clientID1", 30*time.Second)
	assert.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	token := client.Subscribe(topic, qos, mqttHandler(handler{false, "clientID1", &msgChan}))
	assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))
	token = client.Subscribe(fmt.Sprintf("%s.%s", topic, subtopic), qos, mqttHandler(handler{false, "clientID1", &msgChan}))
	assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))

	t.Cleanup(func() {
		token := client.Unsubscribe(topic, fmt.Sprintf("%s.%s", topic, subtopic))
		token.Wait()
		assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))

		client.Disconnect(5)
	})

	// publish with empty topic
	err = pubsub.Publish("", messaging.Message{Payload: data})
	assert.Equal(t, err, mqtt_pubsub.ErrEmptyTopic, fmt.Sprintf("Publish with empty topic: expected: %s, got: %s", mqtt_pubsub.ErrEmptyTopic, err))

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
			Publisher: "clientID11",
			Channel:   tc.channel,
			Subtopic:  tc.subtopic,
			Payload:   tc.payload,
		}
		err := pubsub.Publish(topic, expectedMsg)
		assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s\n", tc.desc, err))

		receivedMsg := <-msgChan
		if tc.payload == nil {
			assert.Equal(t, 0, len(receivedMsg.Payload), fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, tc.payload, receivedMsg.Payload))
		} else {
			assert.Equal(t, tc.payload, receivedMsg.Payload, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, tc.payload, receivedMsg.Payload))
		}
	}
}

func TestSubscribe(t *testing.T) {
	msgChan := make(chan messaging.Message)

	// Creating client to Publish messages to subscribed topic.
	client, err := newClient(address, "mainflux", 30*time.Second)
	assert.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	t.Cleanup(func() {
		client.Unsubscribe()
		client.Disconnect(5)
	})

	cases := []struct {
		desc     string
		topic    string
		clientID string
		err      error
		handler  messaging.MessageHandler
	}{
		{
			desc:     "Subscribe to a topic with an ID",
			topic:    topic,
			clientID: "clientid1",
			err:      nil,
			handler:  handler{false, "clientid1", &msgChan},
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    topic,
			clientID: "clientid2",
			err:      nil,
			handler:  handler{false, "clientid2", &msgChan},
		},
		{
			desc:     "Subscribe to an already subscribed topic with an ID",
			topic:    topic,
			clientID: "clientid1",
			err:      nil,
			handler:  handler{false, "clientid1", &msgChan},
		},
		{
			desc:     "Subscribe to a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s", topic, subtopic),
			clientID: "clientid1",
			err:      nil,
			handler:  handler{false, "clientid1", &msgChan},
		},
		{
			desc:     "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s", topic, subtopic),
			clientID: "clientid1",
			err:      nil,
			handler:  handler{false, "clientid1", &msgChan},
		},
		{
			desc:     "Subscribe to an empty topic with an ID",
			topic:    "",
			clientID: "clientid1",
			err:      mqtt_pubsub.ErrEmptyTopic,
			handler:  handler{false, "clientid1", &msgChan},
		},
		{
			desc:     "Subscribe to a topic with empty id",
			topic:    topic,
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			handler:  handler{false, "", &msgChan},
		},
	}
	for _, tc := range cases {
		err = pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
		assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, err, tc.err))

		if tc.err == nil {
			token := client.Subscribe(tc.topic, qos, mqttHandler(handler{false, "clientID1", &msgChan}))
			assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))

			expectedMsg := messaging.Message{
				Publisher: "clientIDD",
				Channel:   channel,
				Subtopic:  subtopic,
				Payload:   data,
			}
			token = client.Publish(tc.topic, qos, false, expectedMsg.Payload)
			assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))
			token.Wait()

			receivedMsg := <-msgChan
			assert.Equal(t, expectedMsg.Payload, receivedMsg.Payload, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
		}
	}
}

func TestPubSub(t *testing.T) {
	msgChan := make(chan messaging.Message)

	cases := []struct {
		desc     string
		topic    string
		clientID string
		err      error
		handler  messaging.MessageHandler
	}{
		{
			desc:     "Subscribe to a topic with an ID",
			topic:    topic,
			clientID: "clientid7",
			err:      nil,
			handler:  handler{false, "clientid7", &msgChan},
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    topic,
			clientID: "clientid8",
			err:      nil,
			handler:  handler{false, "clientid8", &msgChan},
		},
		{
			desc:     "Subscribe to a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s", topic, subtopic),
			clientID: "clientid7",
			err:      nil,
			handler:  handler{false, "clientid7", &msgChan},
		},
		{
			desc:     "Subscribe to an empty topic with an ID",
			topic:    "",
			clientID: "clientid7",
			err:      mqtt_pubsub.ErrEmptyTopic,
			handler:  handler{false, "clientid7", &msgChan},
		},
		{
			desc:     "Subscribe to a topic with empty id",
			topic:    topic,
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			handler:  handler{false, "", &msgChan},
		},
	}
	for _, tc := range cases {
		err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
		switch err {
		case nil:
			// use pubsub to subscribe to a topic, and then publish messages to that topic.
			expectedMsg := messaging.Message{
				Publisher: "clientID",
				Channel:   channel,
				Subtopic:  subtopic,
				Payload:   data,
			}

			// publish message, and then receive it on message Channel
			err := pubsub.Publish(topic, expectedMsg)
			assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s\n", tc.desc, err))

			receivedMsg := <-msgChan
			assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
		default:
			assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, err, tc.err))
		}
	}
}

func TestUnsubscribe(t *testing.T) {
	msgChan := make(chan messaging.Message)

	cases := []struct {
		desc     string
		topic    string
		clientID string
		err      error
		pubsub   bool //true for subscribe and false for unsubscribe
		handler  messaging.MessageHandler
	}{
		{
			desc:     "Subscribe to a topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid4",
			err:      nil,
			pubsub:   true,
			handler:  handler{false, "clientid4", &msgChan},
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid9",
			err:      nil,
			pubsub:   true,
			handler:  handler{false, "clientid9", &msgChan},
		},
		{
			desc:     "Unsubscribe from a topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid4",
			err:      nil,
			pubsub:   false,
			handler:  handler{false, "clientid4", &msgChan},
		},
		{
			desc:     "Unsubscribe from same topic with different ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid9",
			err:      nil,
			pubsub:   false,
			handler:  handler{false, "clientid9", &msgChan},
		},
		{
			desc:     "Unsubscribe from a non-existent topic with an ID",
			topic:    "h",
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false, "clientid4", &msgChan},
		},
		{
			desc:     "Unsubscribe from an already unsubscribed topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false, "clientid4", &msgChan},
		},
		{
			desc:     "Subscribe to a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "clientidd4",
			err:      nil,
			pubsub:   true,
			handler:  handler{false, "clientidd4", &msgChan},
		},
		{
			desc:     "Unsubscribe from a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "clientidd4",
			err:      nil,
			pubsub:   false,
			handler:  handler{false, "clientidd4", &msgChan},
		},
		{
			desc:     "Unsubscribe from an already unsubscribed topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false, "clientid4", &msgChan},
		},
		{
			desc:     "Unsubscribe from an empty topic with an ID",
			topic:    "",
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrEmptyTopic,
			pubsub:   false,
			handler:  handler{false, "clientid4", &msgChan},
		},
		{
			desc:     "Unsubscribe from a topic with empty ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			pubsub:   false,
			handler:  handler{false, "", &msgChan},
		},
		{
			desc:     "Subscribe to a new topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic+"2"),
			clientID: "clientid55",
			err:      nil,
			pubsub:   true,
			handler:  handler{true, "clientid5", &msgChan},
		},
		{
			desc:     "Unsubscribe from a topic with an ID with failing handler",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic+"2"),
			clientID: "clientid55",
			err:      mqtt_pubsub.ErrFailedHandleMessage,
			pubsub:   false,
			handler:  handler{true, "clientid5", &msgChan},
		},
		{
			desc:     "Subscribe to a new topic with subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic+"2", subtopic),
			clientID: "clientid55",
			err:      nil,
			pubsub:   true,
			handler:  handler{true, "clientid5", &msgChan},
		},
		{
			desc:     "Unsubscribe from a topic with subtopic with an ID with failing handler",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic+"2", subtopic),
			clientID: "clientid55",
			err:      mqtt_pubsub.ErrFailedHandleMessage,
			pubsub:   false,
			handler:  handler{true, "clientid5", &msgChan},
		},
	}
	for _, tc := range cases {
		switch tc.pubsub {
		case true:
			err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
			assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.err, err))
		default:
			err := pubsub.Unsubscribe(tc.clientID, tc.topic)
			assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.err, err))
		}
	}
}

type handler struct {
	fail      bool
	publisher string
	msgChan   *chan (messaging.Message)
}

func (h handler) Handle(msg messaging.Message) error {
	if msg.Publisher != h.publisher {
		*h.msgChan <- msg
	}
	return nil
}

func (h handler) Cancel() error {
	if h.fail {
		return mqtt_pubsub.ErrFailedHandleMessage
	}
	return nil
}

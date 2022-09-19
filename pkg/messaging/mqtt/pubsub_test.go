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
	msgChan = make(chan messaging.Message)
	data    = []byte("payload")
)

func TestPublisher(t *testing.T) {
	// Subscribing with topic, and with subtopic, so that we can publish messages
	client, err := newClient(address, "clientID1", 30*time.Second)
	t.Cleanup(func() {
		client.Unsubscribe()
		client.Disconnect(5)
	})
	assert.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	token := client.Subscribe(topic, qos, nil)
	assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))
	token = client.Subscribe(fmt.Sprintf("%s.%s", topic, subtopic), qos, nil)
	assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))

	// publish with empty topic
	err = pubsub.Publish("", messaging.Message{Payload: data})
	assert.Equal(t, err, mqtt_pubsub.ErrEmptyTopic, fmt.Sprintf("Publish with empty topic: expected: %s, got: %s", mqtt_pubsub.ErrEmptyTopic, err))

	cases := []struct {
		desc     string
		topic    string
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
		err := pubsub.Publish(topic, expectedMsg)
		assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s\n", tc.desc, err))

		receivedMsg := <-msgChan
		if tc.payload == nil {
			assert.Equal(t, 0, len(receivedMsg.Payload), fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, tc.payload, receivedMsg.Payload))
		} else {
			assert.Equal(t, tc.payload, receivedMsg.Payload, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, tc.payload, receivedMsg.Payload))
		}
	}

	token = client.Unsubscribe(topic)
	token.Wait()
	assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))
	token = client.Unsubscribe(fmt.Sprintf("%s.%s", topic, subtopic))
	token.Wait()
	assert.Nil(t, token.Error(), fmt.Sprintf("got unexpected error: %s", token.Error()))
}

func TestSubscribe(t *testing.T) {
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
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    topic,
			clientID: "clientid2",
			err:      nil,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to an already subscribed topic with an ID",
			topic:    topic,
			clientID: "clientid1",
			err:      nil,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s", topic, subtopic),
			clientID: "clientid1",
			err:      nil,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s", topic, subtopic),
			clientID: "clientid1",
			err:      nil,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to an empty topic with an ID",
			topic:    "",
			clientID: "clientid1",
			err:      mqtt_pubsub.ErrEmptyTopic,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with empty id",
			topic:    topic,
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			handler:  handler{false},
		},
	}
	for _, tc := range cases {
		err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
		assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, err, tc.err))
	}
}

func TestPubSub(t *testing.T) {
	client, err := newClient(address, "", 30*time.Second)
	t.Cleanup(func() {
		client.Unsubscribe()
		client.Disconnect(5)
	})
	assert.Nil(t, err, fmt.Sprintf("got unexpected error while creating client: %s", err))

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
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    topic,
			clientID: "clientid8",
			err:      nil,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s", topic, subtopic),
			clientID: "clientid7",
			err:      nil,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to an empty topic with an ID",
			topic:    "",
			clientID: "clientid7",
			err:      mqtt_pubsub.ErrEmptyTopic,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with empty id",
			topic:    topic,
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			handler:  handler{false},
		},
	}
	for _, tc := range cases {
		err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
		switch err {
		case nil:
			// if no error, publish message, and receive after subscribing
			expectedMsg := messaging.Message{
				Channel:  channel,
				Subtopic: subtopic,
				Payload:  data,
			}
			err := pubsub.Publish(topic, expectedMsg)
			assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s\n", tc.desc, err))

			receivedMsg := <-msgChan
			assert.Equal(t, expectedMsg.Payload, receivedMsg.Payload, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
		default:
			assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, err, tc.err))
		}
	}
}

// Tests the subscriber interface, the flow between pubsub.Subscribe and pubsub.Unsubscribe
func TestSubUnsub(t *testing.T) {
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
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to an already subscribed topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid4",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid9",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
		},
		{
			desc:     "Unsubscribe from a topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid4",
			err:      nil,
			pubsub:   false,
			handler:  handler{false},
		},
		{
			desc:     "Unsubscribe from a non-existent topic with an ID",
			topic:    "h",
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false},
		},
		{
			desc:     "Unsubscribe from an already unsubscribed topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "clientidd4",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "clientidd4",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
		},
		{
			desc:     "Unsubscribe from a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "clientidd4",
			err:      nil,
			pubsub:   false,
			handler:  handler{false},
		},
		{
			desc:     "Unsubscribe from an already unsubscribed topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to an empty topic with an ID",
			topic:    "",
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrEmptyTopic,
			pubsub:   true,
			handler:  handler{false},
		},
		{
			desc:     "Unsubscribe from an empty topic with an ID",
			topic:    "",
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrEmptyTopic,
			pubsub:   false,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with empty ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			pubsub:   true,
			handler:  handler{false},
		},
		{
			desc:     "Unsubscribe from a topic with empty ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			pubsub:   false,
			handler:  handler{false},
		},
	}
	for _, tc := range cases {
		switch tc.pubsub {
		case true:
			err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
			assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.err, err))
		default:
			err := pubsub.Unsubscribe(tc.clientID, tc.topic)
			assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.err, err))
		}
	}
}

func TestAll(t *testing.T) {
	client, err := newClient(address, "", 30*time.Second)
	t.Cleanup(func() {
		client.Unsubscribe()
		client.Disconnect(5)
	})
	assert.Nil(t, err, fmt.Sprintf("got unexpected error while creating client: %s", err))
	cases := []struct {
		desc     string
		topic    string
		clientID string
		err      error
		pubsub   bool //true for subscribe and false for unsubscribe
		tryPub   bool
		handler  messaging.MessageHandler
	}{
		{
			desc:     "Subscribe to a topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "1",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "2",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Subscribe to an already subscribed topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "1",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Subscribe to the same topic with a different ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "3",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Unsubscribe from a topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "1",
			err:      nil,
			pubsub:   false,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Unsubscribe from a non-existent topic with an ID",
			topic:    "h",
			clientID: "7",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Unsubscribe from an already unsubscribed topic with an ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "1",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Subscribe to a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "1",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Subscribe to a topic with a subtopic with diff ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "2",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Subscribe to a topic with a subtopic with more diff ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "3",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "1",
			err:      nil,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   true,
		},
		{
			desc:     "Unsubscribe from a topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "1",
			err:      nil,
			pubsub:   false,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Unsubscribe from an already unsubscribed topic with a subtopic with an ID",
			topic:    fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID: "1",
			err:      mqtt_pubsub.ErrNotSubscribed,
			pubsub:   false,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Subscribe to an empty topic with an ID",
			topic:    "",
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrEmptyTopic,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Unsubscribe from an empty topic with an ID",
			topic:    "",
			clientID: "clientid4",
			err:      mqtt_pubsub.ErrEmptyTopic,
			pubsub:   false,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Subscribe to a topic with empty ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			pubsub:   true,
			handler:  handler{false},
			tryPub:   false,
		},
		{
			desc:     "Unsubscribe from a topic with empty ID",
			topic:    fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID: "",
			err:      mqtt_pubsub.ErrEmptyID,
			pubsub:   false,
			handler:  handler{false},
			tryPub:   false,
		},
	}
	for _, tc := range cases {
		switch tc.pubsub {
		case true:
			err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
			assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.err, err))
		default:
			err := pubsub.Unsubscribe(tc.clientID, tc.topic)
			assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.err, err))
		}

		if tc.tryPub {
			expectedMsg := messaging.Message{
				Channel:  channel,
				Subtopic: subtopic,
				Payload:  data,
			}
			token := client.Publish(tc.topic, qos, false, expectedMsg.Payload)
			token.Wait()

			receivedMsg := <-msgChan
			assert.Equal(t, expectedMsg.Payload, receivedMsg.Payload, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
		}
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
		return mqtt_pubsub.ErrFailedHandleMessage
	}
	return nil
}

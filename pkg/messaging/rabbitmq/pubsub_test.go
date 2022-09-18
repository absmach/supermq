// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
	"github.com/stretchr/testify/assert"
)

const (
	topic        = "topic"
	chansPrefix  = "channels"
	channel      = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic     = "engine"
	clientID     = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	exchangeName = "mainflux-exchange"
)

var (
	msgChan = make(chan messaging.Message)
	data    = []byte("payload")
)

func TestPublisher(t *testing.T) {
	// Subscribing with topic, and with subtopic, so that we can publish messages
	conn, ch, err := newConn()
	assert.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	// rabbitTopic := fmt.Sprintf("%s.%s", chansPrefix, topic)
	topicChan := subscribe(t, ch, fmt.Sprintf("%s.%s", chansPrefix, topic))
	subtopicChan := subscribe(t, ch, fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic))

	go rabbitHandler(topicChan, handler{})
	go rabbitHandler(subtopicChan, handler{})

	t.Cleanup(func() {
		conn.Close()
		ch.Close()
	})

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
		fmt.Println("###")
		fmt.Println("TestCase: ", tc.desc)
		fmt.Println("###")
		expectedMsg := messaging.Message{
			Channel:  tc.channel,
			Subtopic: tc.subtopic,
			Payload:  tc.payload,
		}
		err = pubsub.Publish(topic, expectedMsg)
		assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error: %s", tc.desc, err))

		fmt.Println("... receiving message from msgChan ...")
		receivedMsg := <-msgChan
		assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
	}
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
			err:      rabbitmq.ErrEmptyTopic,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with empty id",
			topic:    topic,
			clientID: "",
			err:      rabbitmq.ErrEmptyID,
			handler:  handler{false},
		},
	}
	for _, tc := range cases {
		err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
		assert.Equal(t, err, tc.err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, err, tc.err))
	}
}

func TestSubUnsub(t *testing.T) {
	// Test Subscribe and Unsubscribe
	subcases := []struct {
		desc         string
		topic        string
		clientID     string
		errorMessage error
		pubsub       bool //true for subscribe and false for unsubscribe
		handler      messaging.MessageHandler
	}{
		{
			desc:         "Subscribe to a topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to the same topic with a different ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid2",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an already subscribed topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: nil,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a non-existent topic with an ID",
			topic:        "h",
			clientID:     "clientid1",
			errorMessage: rabbitmq.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from the same topic with a different ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientidd2",
			errorMessage: rabbitmq.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from the same topic with a different ID not subscribed",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientidd3",
			errorMessage: rabbitmq.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "clientid1",
			errorMessage: rabbitmq.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientidd1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an already subscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientidd1",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientidd1",
			errorMessage: nil,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an already unsubscribed topic with a subtopic with an ID",
			topic:        fmt.Sprintf("%s.%s.%s", chansPrefix, topic, subtopic),
			clientID:     "clientid1",
			errorMessage: rabbitmq.ErrNotSubscribed,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to an empty topic with an ID",
			topic:        "",
			clientID:     "clientid1",
			errorMessage: rabbitmq.ErrEmptyTopic,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from an empty topic with an ID",
			topic:        "",
			clientID:     "clientid1",
			errorMessage: rabbitmq.ErrEmptyTopic,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to a topic with empty id",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "",
			errorMessage: rabbitmq.ErrEmptyID,
			pubsub:       true,
			handler:      handler{false},
		},
		{
			desc:         "Unsubscribe from a topic with empty id",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic),
			clientID:     "",
			errorMessage: rabbitmq.ErrEmptyID,
			pubsub:       false,
			handler:      handler{false},
		},
		{
			desc:         "Subscribe to another topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"1"),
			clientID:     "clientid3",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Subscribe to another already subscribed topic with an ID with Unsubscribe failing",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"1"),
			clientID:     "clientid3",
			errorMessage: rabbitmq.ErrFailed,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Subscribe to a new topic with an ID",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"2"),
			clientID:     "clientid4",
			errorMessage: nil,
			pubsub:       true,
			handler:      handler{true},
		},
		{
			desc:         "Unsubscribe from a topic with an ID with failing handler",
			topic:        fmt.Sprintf("%s.%s", chansPrefix, topic+"2"),
			clientID:     "clientid4",
			errorMessage: rabbitmq.ErrFailed,
			pubsub:       false,
			handler:      handler{true},
		},
	}

	for _, tc := range subcases {
		switch tc.pubsub {
		case true:
			err := pubsub.Subscribe(tc.clientID, tc.topic, tc.handler)
			assert.Equal(t, err, tc.errorMessage, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.errorMessage, err))
		default:
			err := pubsub.Unsubscribe(tc.clientID, tc.topic)
			assert.Equal(t, err, tc.errorMessage, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, tc.errorMessage, err))
		}
	}
}

func TestPubSub(t *testing.T) {
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
			err:      rabbitmq.ErrEmptyTopic,
			handler:  handler{false},
		},
		{
			desc:     "Subscribe to a topic with empty id",
			topic:    topic,
			clientID: "",
			err:      rabbitmq.ErrEmptyID,
			handler:  handler{false},
		},
	}
	for _, tc := range cases {
		fmt.Println()
		fmt.Println("####")
		fmt.Println("Test Case: ", tc.desc)
		fmt.Println("####")
		fmt.Println()

		subject := ""
		if tc.topic != "" {
			subject = fmt.Sprintf("%s.%s", chansPrefix, tc.topic)
		}
		err := pubsub.Subscribe(tc.clientID, subject, tc.handler)
		switch tc.err {
		case nil:
			fmt.Println()
			fmt.Println("pubsub.Subscribe() error ->", err)
			fmt.Println()

			assert.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", tc.desc, err))

			// if no error, publish message, and receive after subscribing
			expectedMsg := messaging.Message{
				Channel: channel,
				Payload: data,
			}
			fmt.Println()
			fmt.Println("expectedMsg -> ", expectedMsg)
			fmt.Println()

			err = pubsub.Publish(tc.topic, expectedMsg)
			fmt.Println()
			fmt.Println("pubsub.Publish() error ->", err)
			fmt.Println()
			assert.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", tc.desc, err))

			fmt.Println("Now, waiting for message on channel")
			receivedMsg := <-msgChan
			fmt.Println()
			fmt.Println("receivedMsg -> ", receivedMsg)
			fmt.Println()

			assert.Equal(t, expectedMsg.Payload, receivedMsg.Payload, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))

			err = pubsub.Unsubscribe(tc.clientID, fmt.Sprintf("%s.%s", chansPrefix, tc.topic))
			assert.Nil(t, err, fmt.Sprintf("%s got unexpected error: %s", tc.desc, err))
		default:
			assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected: %s, but got: %s", tc.desc, err, tc.err))
		}
	}
}

type handler struct {
	fail bool
}

func (h handler) Handle(msg messaging.Message) error {
	fmt.Println()
	fmt.Println("Inside pubsub_test.go -> Handle")
	fmt.Println("msg => ", msg)
	fmt.Println()
	msgChan <- msg
	fmt.Println("msg sent to msgChan")
	return nil
}

func (h handler) Cancel() error {
	if h.fail {
		return rabbitmq.ErrFailed
	}
	return nil
}

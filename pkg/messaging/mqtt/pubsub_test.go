// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic    = "topic"
	data     = "payload"
	channel  = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic = "engine"
	wait     = 15
)

var (
	msgChan = make(chan messaging.Message)
)

func TestPubsub(t *testing.T) {
	err := subscriber.Subscribe(topic, handler)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	time.Sleep(time.Second * wait)

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
			payload: []byte(data),
		},
		{
			desc:    "publish message with channel",
			payload: []byte(data),
			channel: channel,
		},
		{
			desc:     "publish message with subtopic",
			payload:  []byte(data),
			subtopic: subtopic,
		},
		{
			desc:     "publish message with channel and subtopic",
			payload:  []byte(data),
			channel:  channel,
			subtopic: subtopic,
		},
	}

	for _, tc := range cases {
		expectedMsg := message(tc.channel, tc.subtopic, tc.payload)
		payload, err := payload(expectedMsg)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		err = publisher.Publish(topic, messaging.Message{Payload: payload})
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		receivedMsg := <-msgChan
		// _ = receivedMsg
		assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
	}
}

func handler(msg messaging.Message) error {
	msgChan <- msg
	return nil
}

func message(channel, subtopic string, payload []byte) messaging.Message {
	return messaging.Message{
		Channel:  channel,
		Subtopic: subtopic,
		Payload:  payload,
	}
}
func payload(m messaging.Message) ([]byte, error) {
	protoMsg, err := proto.Marshal(&m)
	if err != nil {
		return nil, err
	}
	return protoMsg, nil
}

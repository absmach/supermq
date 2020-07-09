// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic    = "topic"
	channel  = "9b7b1b3f-b1b0-46a8-a717-b8213f9eda3b"
	subtopic = "engine"
)

var (
	msgChan = make(chan messaging.Message)
	data    = []byte("payload")
)

func TestPubsub(t *testing.T) {
	err := subscriber.Subscribe(topic, handler)
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
			desc:    "publish message with non-nil payload",
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
		expectedMsg := message(tc.channel, tc.subtopic, tc.payload)
		payload, err := payload(expectedMsg)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		err = publisher.Publish(topic, messaging.Message{Payload: payload})
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		receivedMsg := <-msgChan
		assert.Equal(t, expectedMsg, receivedMsg, fmt.Sprintf("%s: expected %+v got %+v\n", tc.desc, expectedMsg, receivedMsg))
	}
}

var createds []int64
var created int64

func contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func handler(msg messaging.Message) error {
	if !contains(createds, msg.Created) {
		createds = append(createds, created)
		msgChan <- msg
	}
	return nil
}

func message(channel, subtopic string, payload []byte) messaging.Message {
	created++
	return messaging.Message{
		Channel:  channel,
		Subtopic: subtopic,
		Payload:  payload,
		Created:  created,
	}
}
func payload(m messaging.Message) ([]byte, error) {
	protoMsg, err := proto.Marshal(&m)
	if err != nil {
		return nil, err
	}
	return protoMsg, nil
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	topic = "topic"
)

var (
	msgChan = make(chan messaging.Message)
)

func TestSubscribe(t *testing.T) {
	err := subscriber.Subscribe(topic, handler)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc    string
		payload []byte
	}{
		// {
		// 	desc:    "publish empty payload message",
		// 	payload: []byte{},
		// },
		{
			desc:    "publish message with string payload",
			payload: []byte("e03a31c4-5f53-483f-8f25-f7e570aa7101"),
		},
	}

	for _, tc := range cases {
		err = publisher.Publish(topic, messaging.Message{Payload: tc.payload})
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		msg := <-msgChan
		assert.Equal(t, true, assert.ObjectsAreEqualValues(tc.payload, msg.Payload), fmt.Sprintf("%s: expected payload %s got %s\n", tc.desc, string(tc.payload), string(msg.Payload)))
	}
}

func handler(msg messaging.Message) error {
	msgChan <- msg
	return nil
}

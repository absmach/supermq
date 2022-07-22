// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package ws_test

import (
	"context"
	"testing"

	"github.com/mainflux/mainflux"
	httpmock "github.com/mainflux/mainflux/http/mocks"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
	"github.com/mainflux/mainflux/ws/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	chanID   = "1"
	id       = "1"
	thingKey = "thing_key"
	subTopic = "subtopic"
	protocol = "ws"
)

var msg = messaging.Message{
	Channel:   chanID,
	Publisher: id,
	Subtopic:  "",
	Protocol:  protocol,
	Payload:   []byte(`[{"n":"current","t":-5,"v":1.2}]`),
}

func newService(cc mainflux.ThingsServiceClient) ws.Service {
	pubsub := mocks.NewPubSub()
	return ws.New(cc, pubsub)
}

func NewThingsClient() mainflux.ThingsServiceClient {
	return httpmock.NewThingsClient(map[string]string{thingKey: chanID})
}

func TestPublish(t *testing.T) {
	thingsClient := NewThingsClient()
	svc := newService(thingsClient)

	cases := []struct {
		name     string
		thingKey string
		msg      messaging.Message
		err      error
	}{
		{
			name:     "publish a valid message with valid thingKey",
			thingKey: thingKey,
			msg:      msg,
			err:      nil,
		},
		{
			name:     "publish a valid message with empty thingKey",
			thingKey: "",
			msg:      msg,
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "publish a valid message with invalid thingKey",
			thingKey: "invalid",
			msg:      msg,
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "publish an empty message with valid thingKey",
			thingKey: thingKey,
			msg:      messaging.Message{},
			err:      ws.ErrFailedMessagePublish,
		},
		{
			name:     "publish an empty message with empty thingKey",
			thingKey: "",
			msg:      messaging.Message{},
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "publish an empty message with invalid thingKey",
			thingKey: "invalid",
			msg:      messaging.Message{},
			err:      ws.ErrUnauthorizedAccess,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.err, svc.Publish(context.Background(), tt.thingKey, tt.msg))
		})
	}
}

func TestSubscribe(t *testing.T) {
	thingsClient := NewThingsClient()
	svc := newService(thingsClient)

	c := ws.NewClient(nil, thingKey)

	cases := []struct {
		name     string
		thingKey string
		chanID   string
		subtopic string
		err      error
	}{
		{
			name:     "subscribe to channel with valid thingKey, chanID, subtopic",
			thingKey: thingKey,
			chanID:   chanID,
			subtopic: subTopic,
			err:      nil,
		},
		{
			name:     "subscribe to channel with invalid chanID and invalid thingKey",
			thingKey: "invalid",
			chanID:   "0",
			subtopic: subTopic,
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "subscribe to channel with empty channel",
			thingKey: thingKey,
			chanID:   "",
			subtopic: subTopic,
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "subscribe to channel with empty thingKey",
			thingKey: "",
			chanID:   chanID,
			subtopic: subTopic,
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "subscribe to channel with empty thingKey and empty channel",
			thingKey: "",
			chanID:   "",
			subtopic: subTopic,
			err:      ws.ErrUnauthorizedAccess,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.err, svc.Subscribe(context.Background(), tt.thingKey, tt.chanID, tt.subtopic, c))
		})
	}
}

func TestUnsubscribe(t *testing.T) {
	thingsClient := NewThingsClient()
	svc := newService(thingsClient)

	cases := []struct {
		name     string
		thingKey string
		chanID   string
		subtopic string
		err      error
	}{
		{
			name:     "unsubscribe from channel with valid thingKey, chanID, subtopic",
			thingKey: thingKey,
			chanID:   chanID,
			subtopic: subTopic,
			err:      nil,
		},
		{
			name:     "unsubscribe from channel with valid thingKey, chanID, and empty subtopic",
			thingKey: thingKey,
			chanID:   chanID,
			subtopic: "",
			err:      nil,
		},
		{
			name:     "unsubscribe from channel with empty channel",
			thingKey: thingKey,
			chanID:   "",
			subtopic: subTopic,
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "unsubscribe from channel with empty thingKey",
			thingKey: "",
			chanID:   chanID,
			subtopic: subTopic,
			err:      ws.ErrUnauthorizedAccess,
		},
		{
			name:     "unsubscribe from channel with empty thingKey and empty channel",
			thingKey: "",
			chanID:   "",
			subtopic: subTopic,
			err:      ws.ErrUnauthorizedAccess,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.err, svc.Unsubscribe(context.Background(), tt.thingKey, tt.chanID, tt.subtopic))
		})
	}
}

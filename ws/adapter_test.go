// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package ws_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
	"github.com/mainflux/mainflux/ws/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	chanID   = "1"
	pubID    = "1"
	protocol = "ws"
)

var msg = messaging.Message{
	Channel:   chanID,
	Subtopic:  "",
	Publisher: pubID,
	Protocol:  protocol,
	Payload:   []byte(`[{"n":"current","t":5,"v":1.2}]`),
	Created:   time.Now().Unix(),
}

var (
	conn   = websocket.Conn{}
	client = ws.NewClient(&conn, "", nil)
)

func newService(client *ws.Connclient) ws.Service {
	subs := map[string]*ws.Connclient{chanID: client}
	pubsub := mocks.New(subs)

	return ws.New(nil, pubsub)
}

func TestPublish(t *testing.T) {
	svc := newService(client)

	cases := []struct {
		desc string
		msg  messaging.Message

		err error
	}{
		{
			desc: "publish valid message",
			msg:  msg,
			err:  nil,
		},
		{
			desc: "publish with empty message",
			msg:  msg,
			err:  ws.ErrFailedMessagePublish,
		},
	}

	for _, tc := range cases {
		// Check if message was sent
		go func(desc string, tcMsg messaging.Message) {
			receivedMsg := <-client.Messages
			assert.Equal(t, tcMsg, receivedMsg, fmt.Sprintf("%s: expected %v got %v\n", desc, tcMsg, receivedMsg))
		}(tc.desc, tc.msg)

		// Check if publish succeeded.
		err := svc.Publish(context.Background(), "", tc.msg)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expect %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestSubscribe(t *testing.T) {
	go func(client *ws.Connclient) {
		receivedMsg := <-client.Messages
		assert.Equal(t, msg, receivedMsg, fmt.Sprintf("send message to channel: expected %v got %v\n", msg, receivedMsg))
	}(client)

	client.Send(msg)
}

func TestClose(t *testing.T) {
	go func() {
		closed := <-client.Closed
		assert.True(t, closed, "channel closed statyed open")
	}()

	client.Close()
}

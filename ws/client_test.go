// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package ws_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/ws"
	"github.com/stretchr/testify/assert"
)

var (
	msgChan     = make(chan []byte)
	c           *ws.Client
	receivedMsg []byte
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			msgChan <- []byte("empty-string")
			break
		}
		msgChan <- message
	}
}
func TestHandle(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// Connect to the server
	wsConn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer wsConn.Close()

	c = ws.NewClient(wsConn)

	cases := []struct {
		desc            string
		publisher       string
		expectedPayload []byte
	}{
		{
			desc:            "handling with different id from ws.Client",
			publisher:       msg.Publisher,
			expectedPayload: msg.Payload,
		},
		{
			desc:            "handling with same id as ws.Client",
			publisher:       "",
			expectedPayload: []byte{},
		},
	}

	for _, tc := range cases {
		msg.Publisher = tc.publisher
		err = c.Handle(msg)
		assert.Nil(t, err, fmt.Sprintf("expected nil error from handle, got: %s", err))

		select {
		case rec := <-msgChan:
			receivedMsg = rec
		case <-time.After(time.Duration(5) * time.Second):
			receivedMsg = []byte{}
		}
		assert.Equal(t, tc.expectedPayload, receivedMsg, fmt.Sprintf("%s: expected %+v, got %+v", tc.desc, msg, receivedMsg))
		// assert.Equal(t, 0, len(receivedMsg), fmt.Sprintf("expected empty message, got %+v", receivedMsg))
	}
}

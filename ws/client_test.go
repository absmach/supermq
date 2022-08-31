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
	msgChan = make(chan []byte)
	c       *ws.Client
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
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer conn.Close()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			msgChan <- []byte{}
			break
		}
		msgChan <- message
	}
}
func TestHandle(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := strings.Replace(s.URL, "http", "ws", 1)

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
		flagWait        bool
		expectedPayload []byte
	}{
		{
			desc:            "handling with different id from ws.Client",
			publisher:       msg.Publisher,
			flagWait:        true,
			expectedPayload: msg.Payload,
		},
		{
			desc:            "handling with same id as ws.Client (empty as default)",
			publisher:       "",
			flagWait:        false,
			expectedPayload: []byte{},
		},
	}

	for _, tc := range cases {
		msg.Publisher = tc.publisher
		err = c.Handle(msg)
		assert.Nil(t, err, fmt.Sprintf("expected nil error from handle, got: %s", err))

		receivedMsg := []byte{}
		if tc.flagWait {
			receivedMsg = <-msgChan
		}
		assert.Equal(t, tc.expectedPayload, receivedMsg, fmt.Sprintf("%s: expected %+v, got %+v", tc.desc, msg, receivedMsg))
	}
}

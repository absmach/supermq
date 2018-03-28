package ws_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/ws"
	"github.com/mainflux/mainflux/ws/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux"
)

var (
	validMsg = mainflux.RawMessage{
		Channel:   "1",
		Publisher: "1",
		Protocol:  "ws",
		Payload:   []byte(`{"n":"current","t":-5,"v":1.2}`),
	}
	emptyMsg = mainflux.RawMessage{}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type socketPair struct {
	server ws.Socket
	client ws.Socket
}

func newService() ws.Service {
	logger := log.NewNopLogger()
	pub := mocks.NewMessagePubSub()

	return ws.New(pub, logger)
}

func newSocket(t *testing.T) (socketPair, error) {
	var serverSocket ws.Socket
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatal(fmt.Sprintf("unexpected error: %s", err))
		}
		serverSocket = ws.NewSocket(conn)
	}))
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return socketPair{}, err
	}
	clientSocket := ws.NewSocket(conn)
	return socketPair{serverSocket, clientSocket}, nil
}

func TestPublish(t *testing.T) {
	svc := newService()

	cases := map[string]struct {
		msg mainflux.RawMessage
		err error
	}{
		"publish valid message": {validMsg, nil},
		"publish empty message": {emptyMsg, ws.ErrFailedMessagePublish},
	}

	for desc, tc := range cases {
		err := svc.Publish(tc.msg)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestBroadcast(t *testing.T) {
	svc := newService()
	sockets, err := newSocket(t)
	if err != nil {
		t.Fatal(fmt.Sprintf("unexpected error: %s", err))
	}

	cases := map[string]struct {
		msg mainflux.RawMessage
		err error
	}{
		"broadcast valid message":                 {validMsg, nil},
		"broadcast failed for unexpected reasons": {validMsg, ws.ErrFailedMessageBroadcast},
	}

	for desc, tc := range cases {
		err := svc.Broadcast(tc.msg, func(msg mainflux.RawMessage) error {
			if err := sockets.server.Write(msg); err != nil {
				return err
			}
			return tc.err
		})
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		_, bytes, err := sockets.client.ReadMessage()
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, tc.msg.Payload, bytes, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.msg.Payload, bytes))
	}
}

func TestSubscribe(t *testing.T) {
	svc := newService()

	cases := []struct {
		desc    string
		channel string
		err     error
	}{
		{"subscription to valid channel", "1", nil},
		{"subscription to channel that should fail", "1", ws.ErrFailedSubscription},
	}

	for _, tc := range cases {
		_, err := svc.Subscribe(tc.channel, func(mainflux.RawMessage) {})
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestListen(t *testing.T) {
	svc := newService()

	cases := map[string]struct {
		sub   ws.Subscription
		msg   mainflux.RawMessage
		close bool
	}{
		"send message and close connection":       {ws.Subscription{"pubID", "chanID"}, validMsg, true},
		"send message and don't close connection": {ws.Subscription{"pubID", "chanID"}, validMsg, false},
		"close connection":                        {ws.Subscription{"pubID", "chanID"}, emptyMsg, true},
		"don't close connection":                  {ws.Subscription{"pubID", "chanID"}, emptyMsg, false},
	}

	for desc, tc := range cases {
		sockets, err := newSocket(t)
		if err != nil {
			t.Fatal(fmt.Sprintf("unexpected error: %s", err))
		}

		close := false
		go svc.Listen(sockets.server, tc.sub, func() {
			close = true
		})
		// Send message if not empty.
		if len(tc.msg.Payload) == 0 {
			sockets.client.WriteMessage(websocket.TextMessage, tc.msg.Payload)
		}

		// Close connection if needed.
		if tc.close {
			sockets.client.Close()
		}

		// Wait for connection to close.
		time.Sleep(1 * time.Second)
		assert.Equal(t, tc.close, close, fmt.Sprintf("%s: expected %t got %t\n", desc, tc.close, close))
	}
}

package ws_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
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

func newService() mainflux.MessagePubSub {
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
		err := svc.Publish(tc.msg, nil)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestSubscribe(t *testing.T) {
	svc := newService()
	sockets, err := newSocket(t)
	if err != nil {
		t.Fatal(fmt.Sprintf("unexpected error: %s\n", err))
	}

	cases := []struct {
		desc      string
		sub       mainflux.Subscription
		readCount uint32
		err       error
	}{
		{"subscription to valid channel", mainflux.Subscription{PubID: "1", ChanID: "1"}, 5, nil},
		{"subscription to channel that should fail", mainflux.Subscription{PubID: "2", ChanID: "1"}, 0, ws.ErrFailedSubscription},
	}

	for _, tc := range cases {
		var readCount uint32
		tc.sub.Write = func(mainflux.RawMessage) error {
			return nil
		}
		tc.sub.Read = func() ([]byte, error) {
			_, payload, err := sockets.server.ReadMessage()
			atomic.AddUint32(&readCount, 1) // read message was called
			return payload, err
		}
		_, err := svc.Subscribe(tc.sub, nil)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))

		// Write messages that should be read.
		for i := 0; i < int(tc.readCount); i++ {
			sockets.client.Write(validMsg)
		}

		// Wait for message to be received.
		time.Sleep(1 * time.Second)
		assert.Equal(t, tc.readCount, readCount, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.readCount, readCount))
	}
}

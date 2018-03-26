package ws_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

func newService() ws.Service {
	logger := log.NewNopLogger()
	pub := mocks.NewMessagePubSub()

	return ws.New(pub, logger)
}

func newSocket(t *testing.T) (ws.Socket, error) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := upgrader.Upgrade(w, r, nil); err != nil {
			t.Fatal(fmt.Sprintf("unexpected error: %s", err))
		}
	}))
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return ws.Socket{}, err
	}
	socket := ws.NewSocket(conn)
	return socket, nil
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
	socket, err := newSocket(t)
	if err != nil {
		t.Fatal(fmt.Sprintf("unexpected error: %s", err))
	}

	cases := map[string]struct {
		msg mainflux.RawMessage
		err error
	}{
		"broadcast valid message": {validMsg, nil},
	}

	for desc, tc := range cases {
		err := svc.Broadcast(socket, tc.msg)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
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

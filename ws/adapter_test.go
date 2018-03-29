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
		err := svc.Publish(tc.msg)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestSubscribe(t *testing.T) {
	svc := newService()

	cases := []struct {
		desc string
		sub  mainflux.Subscription
		err  error
	}{
		{"subscription to valid channel", mainflux.Subscription{"1", "1"}, nil},
		{"subscription to channel that should fail", mainflux.Subscription{"2", "1"}, ws.ErrFailedSubscription},
	}

	for _, tc := range cases {
		_, err := svc.Subscribe(
			tc.sub,
			func(mainflux.RawMessage) error {
				return nil
			},
			func() ([]byte, error) {
				return []byte{}, nil
			},
		)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

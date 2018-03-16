package ws_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/mainflux/mainflux/ws"
	"github.com/mainflux/mainflux/ws/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux"
)

var (
	errFailedMessagePublish = errors.New("failed to publish message")
)

var (
	validMsg = mainflux.RawMessage{
		Channel:   "1",
		Publisher: "1",
		Protocol:  "ws",
		Payload:   []byte(`{"n":"current","t":-5,"v":1.2}`),
	}
	emptyMsg = mainflux.RawMessage{}
)

func newService() ws.Service {
	logger := log.NewNopLogger()
	pub := mocks.NewMessagePublisher()

	return ws.New(pub, logger)
}

func TestPublish(t *testing.T) {
	svc := newService()

	cases := map[string]struct {
		msg mainflux.RawMessage
		err error
	}{
		"publish valid message": {validMsg, nil},
		"publish empty message": {emptyMsg, errFailedMessagePublish},
	}

	for desc, tc := range cases {
		err := svc.Publish(tc.msg)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

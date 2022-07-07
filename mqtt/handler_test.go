package mqtt

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt/mocks"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	thingID   = "thingID"
	username  = "username"
	invalidID = "invalidID"
	password  = "password"
)

var buf bytes.Buffer

func TestAuthConnect(t *testing.T) {
	handler := newHandler()
	cases := []struct {
		desc    string
		err     error
		session *session.Client
	}{
		{
			desc:    "connect without active session",
			err:     errInvalidConnect,
			session: nil,
		},
		//{
		//	desc: "connect with valid id",
		//	err:  errors.ErrAuthentication,
		//	session: &session.Client{
		//		ID:       thingID,
		//		Username: username,
		//		Password: []byte(""),
		//	},
		//},
		//{
		//	desc: "connect with valid password and invalid username",
		//	err:  errors.ErrAuthentication,
		//	session: &session.Client{
		//		ID:       thingID,
		//		Username: invalidID,
		//		Password: []byte(password),
		//	},
		//},
		{
			desc: "connect with valid username and password",
			err:  nil,
			session: &session.Client{
				ID:       thingID,
				Username: username,
				Password: []byte(password),
			},
		},
	}

	for _, tc := range cases {
		err := handler.AuthConnect(tc.session)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestAuthPublish(t *testing.T) {
	handler := newHandler()
	var topic string
	var payload = []byte("[{'n':'test-name', 'v': 1.2}]")

	sessionClient := session.Client{
		ID:       thingID,
		Username: username,
		Password: []byte(password),
	}

	cases := []struct {
		desc    string
		client  *session.Client
		err     error
		topic   *string
		payload []byte
	}{
		{
			desc:    "publish without active session",
			client:  nil,
			err:     errInactiveSession,
			topic:   &topic,
			payload: payload,
		},
		{
			desc:    "publish without topic",
			client:  &sessionClient,
			err:     errNilTopicPub,
			topic:   nil,
			payload: payload,
		},
		{
			desc:    "publish with active session, valid topic and payload",
			client:  &sessionClient,
			err:     errMalformedTopic,
			topic:   &topic,
			payload: payload,
		},
	}

	for _, tc := range cases {
		err := handler.AuthPublish(tc.client, tc.topic, &tc.payload)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestAuthSubscribe(t *testing.T) {
	handler := newHandler()
	var idProvider = uuid.NewMock()

	chID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	var topics = []string{"channels/" + chID + "/messages/2/ct/3"}
	var invalidTopics = []string{"topic"}

	sessionClient := session.Client{
		ID:       thingID,
		Username: username,
		Password: []byte(password),
	}

	cases := []struct {
		desc   string
		client *session.Client
		err    error
		topic  *[]string
	}{
		{
			desc:   "subscribe without active session",
			client: nil,
			err:    errInactiveSession,
			topic:  &topics,
		},
		{
			desc:   "subscribe without topics",
			client: &sessionClient,
			err:    errNilTopicSub,
			topic:  nil,
		},
		{
			desc:   "subscribe with invalid channel",
			client: &sessionClient,
			err:    errMalformedTopic,
			topic:  &invalidTopics,
		},
		{
			desc:   "subscribe with active session and valid topics",
			client: &sessionClient,
			err:    nil,
			topic:  &topics,
		},
	}

	for _, tc := range cases {
		err := handler.AuthSubscribe(tc.client, tc.topic)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestPublish(t *testing.T) {
	handler := newHandler()
	buf.Reset()
	sessionClient := session.Client{
		ID:       thingID,
		Username: username,
		Password: []byte(password),
	}

	cases := []struct {
		desc     string
		client   *session.Client
		topic    string
		payload  []byte
		expected string
	}{
		{
			desc:     "publish without active session",
			client:   nil,
			topic:    "topic",
			payload:  []byte("payload"),
			expected: "Nil client publish",
		},
		{
			desc:     "publish with invalid channel parts",
			client:   &sessionClient,
			topic:    "topic",
			payload:  []byte("payload"),
			expected: "Publish - client ID thingID to the topic: topic",
		},
	}
	for _, tc := range cases {
		handler.Publish(tc.client, &tc.topic, &tc.payload)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func newHandler() session.Handler {
	logger, err := logger.New(&buf, "debug")
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}

	authClient := mocks.NewClient(map[string]string{password: username})
	eventStore := mocks.NewEventStore()

	return NewHandler([]messaging.Publisher{mocks.NewPublisher()}, eventStore, logger, authClient)
}

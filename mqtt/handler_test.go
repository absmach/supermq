package mqtt

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	rdb "github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt/mocks"
	"github.com/mainflux/mainflux/mqtt/redis"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	id              = "123"
	username        = "username"
	invalidUsername = "invalidUsername"
	password        = "password"
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
		{
			desc: "connect with valid id",
			err:  errors.ErrAuthentication,
			session: &session.Client{
				ID:       id,
				Username: username,
				Password: []byte(""),
			},
		},
		{
			desc: "connect with valid password and invalid username",
			err:  errors.ErrAuthentication,
			session: &session.Client{
				ID:       id,
				Username: invalidUsername,
				Password: []byte(password),
			},
		},
		{
			desc: "connect with valid username and password",
			err:  nil,
			session: &session.Client{
				ID:       id,
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
	var payload = []byte("payload")

	sessionClient := session.Client{
		ID:       id,
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
			err:     errNilClient,
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
		ID:       id,
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
			err:    errNilClient,
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
		ID:       id,
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
			expected: "Publish - client ID 123 to the topic: topic",
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
		panic(err)
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("redis", "5.0-alpine", nil)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	redisClient := rdb.NewClient(&rdb.Options{
		Addr:     fmt.Sprintf("localhost:%s", container.GetPort("6379/tcp")),
		Password: "",
		DB:       0,
	})

	eventStore := redis.NewEventStore(redisClient, "")
	return NewHandler([]messaging.Publisher{mocks.NewPublisher()}, eventStore, logger, mocks.NewClient(map[string]string{password: username}))
}

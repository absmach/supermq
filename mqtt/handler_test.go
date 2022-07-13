package mqtt

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt/mocks"
	mqttredis "github.com/mainflux/mainflux/mqtt/redis"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	thingID              = "thingID"
	clientID             = "clientID"
	invalidThingID       = "invalidThingID"
	password             = "password"
	subscribeMsg         = "subscribed with client_id: %s to topics: %s"
	unsubscribeMsg       = "unsubscribe client_id: %s from topics: %s"
	failedSubscribeMsg   = "failed to subscribe: %s"
	failedUnsubscribeMsg = "failed to unsubscribe: %s"
	connectMsg           = "connected with client_id: %s"
	disconnectMsg        = "disconnected client_id: %s and username: %s"
	failedDisconnectMsg  = "failed to disconnect: %s"
	publishMsg           = "published with client_id: %s to the topic: %s"
	failedConnectMsg     = "failed to connect: %s"
)

var buf bytes.Buffer
var sessionClient = session.Client{
	ID:       clientID,
	Username: thingID,
	Password: []byte(password),
}

func TestAuthConnect(t *testing.T) {
	handler := newHandler()

	cases := []struct {
		desc    string
		err     error
		session *session.Client
	}{
		{
			desc:    "connect without active session",
			err:     errClientNotInitialized,
			session: nil,
		},
		{
			desc: "connect without clientID",
			err:  nil,
			session: &session.Client{
				ID:       "",
				Username: thingID,
				Password: []byte("password"),
			},
		},
		{
			desc: "connect with invalid password",
			err:  errors.ErrAuthentication,
			session: &session.Client{
				ID:       clientID,
				Username: thingID,
				Password: []byte(""),
			},
		},
		{
			desc: "connect with valid password and invalid username",
			err:  errors.ErrAuthentication,
			session: &session.Client{
				ID:       clientID,
				Username: invalidThingID,
				Password: []byte(password),
			},
		},
		{
			desc: "connect with valid username and password",
			err:  nil,
			session: &session.Client{
				ID:       clientID,
				Username: thingID,
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
	var invalidTopic = "invalidTopic"
	var payload = []byte("[{'n':'test-name', 'v': 1.2}]")
	var idProvider = uuid.NewMock()

	chID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	var topic = "channels/" + chID + "/messages"

	cases := []struct {
		desc    string
		client  *session.Client
		err     error
		topic   *string
		payload []byte
	}{
		{
			desc:    "publish with inactive client",
			client:  nil,
			err:     errClientNotInitialized,
			topic:   &topic,
			payload: payload,
		},
		{
			desc:    "publish without topic",
			client:  &sessionClient,
			err:     errMissingTopicPub,
			topic:   nil,
			payload: payload,
		},
		{
			desc:    "publish with malformed topic",
			client:  &sessionClient,
			err:     errMalformedTopic,
			topic:   &invalidTopic,
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
	var invalidTopics = []string{"topic"}

	chID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	var topics = []string{"channels/" + chID + "/messages"}

	cases := []struct {
		desc   string
		client *session.Client
		err    error
		topic  *[]string
	}{
		{
			desc:   "subscribe without active session",
			client: nil,
			err:    errClientNotInitialized,
			topic:  &topics,
		},
		{
			desc:   "subscribe without topics",
			client: &sessionClient,
			err:    errMissingTopicSub,
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

func TestConnect(t *testing.T) {
	handler := newHandler()
	buf.Reset()
	cases := []struct {
		desc     string
		client   *session.Client
		expected string
	}{
		{
			desc:     "connect without active session",
			client:   nil,
			expected: fmt.Sprintf(failedConnectMsg, errClientNotInitialized.Error()),
		},
		{
			desc:     "connect with valid session",
			client:   &sessionClient,
			expected: fmt.Sprintf(connectMsg, clientID),
		},
	}

	for _, tc := range cases {
		handler.Connect(tc.client)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestPublish(t *testing.T) {
	handler := newHandler()
	var idProvider = uuid.NewMock()
	var invalidTopic = "invalidTopic"
	buf.Reset()

	chID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	var topic = "channels/" + chID + "/messages"

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
			topic:    topic,
			payload:  []byte("payload"),
			expected: errClientNotInitialized.Error(),
		},
		{
			desc:     "publish with invalid topic",
			client:   &sessionClient,
			topic:    invalidTopic,
			payload:  []byte("payload"),
			expected: fmt.Sprintf(publishMsg, clientID, invalidTopic),
		},
	}

	for _, tc := range cases {
		handler.Publish(tc.client, &tc.topic, &tc.payload)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestSubscribe(t *testing.T) {
	handler := newHandler()
	var idProvider = uuid.NewMock()
	buf.Reset()

	chID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	var topics = []string{"channels/" + chID + "/messages"}

	cases := []struct {
		desc     string
		client   *session.Client
		topic    []string
		expected string
	}{
		{
			desc:     "subscribe without active session",
			client:   nil,
			topic:    topics,
			expected: fmt.Sprintf(failedSubscribeMsg, errClientNotInitialized.Error()),
		},
		{
			desc:     "subscribe with valid session and topics",
			client:   &sessionClient,
			topic:    topics,
			expected: fmt.Sprintf(subscribeMsg, clientID, topics[0]),
		},
	}

	for _, tc := range cases {
		handler.Subscribe(tc.client, &tc.topic)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestUnsubscribe(t *testing.T) {
	handler := newHandler()
	var idProvider = uuid.NewMock()
	buf.Reset()

	chID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	var topics = []string{"channels/" + chID + "/messages"}

	cases := []struct {
		desc     string
		client   *session.Client
		topic    []string
		expected string
	}{
		{
			desc:     "unsubscribe without active session",
			client:   nil,
			topic:    topics,
			expected: fmt.Sprintf(failedUnsubscribeMsg, errClientNotInitialized.Error()),
		},
		{
			desc:     "unsubscribe with valid session and topics",
			client:   &sessionClient,
			topic:    topics,
			expected: fmt.Sprintf(unsubscribeMsg, clientID, topics[0]),
		},
	}

	for _, tc := range cases {
		handler.Unsubscribe(tc.client, &tc.topic)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestDisconnect(t *testing.T) {
	handler := newHandler()
	var idProvider = uuid.NewMock()
	buf.Reset()

	chID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	var topics = []string{"channels/" + chID + "/messages"}

	cases := []struct {
		desc     string
		client   *session.Client
		topic    []string
		expected string
	}{
		{
			desc:     "disconect without active session",
			client:   nil,
			topic:    topics,
			expected: fmt.Sprintf(failedDisconnectMsg, errClientNotInitialized.Error()),
		},
		{
			desc:     "disconect with valid session",
			client:   &sessionClient,
			topic:    topics,
			expected: fmt.Sprintf(disconnectMsg, clientID, thingID),
		},
	}

	for _, tc := range cases {
		handler.Disconnect(tc.client)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func newHandler() session.Handler {
	logger, err := logger.New(&buf, "debug")
	if err != nil {
		log.Fatalf("failed to create logger: %s", err)
	}

	authClient := mocks.NewClient(map[string]string{password: thingID})

	var redisClient *redis.Client

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("redis", "5.0-alpine", nil)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	if err := pool.Retry(func() error {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("localhost:%s", container.GetPort("6379/tcp")),
			Password: "",
			DB:       0,
		})

		return redisClient.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	eventStore := mqttredis.NewEventStore(redisClient, "1")

	return NewHandler([]messaging.Publisher{mocks.NewPublisher()}, eventStore, logger, authClient)
}

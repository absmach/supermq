package mqtt_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt"
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
	thingID        = "thingID"
	chanID         = "123e4567-e89b-12d3-a456-000000000001"
	invalidChanID  = "1"
	clientID       = "clientID"
	invalidThingID = "invalidThingID"
	password       = "password"
)

var (
	invalidTopic  = "invalidTopic"
	idProvider    = uuid.NewMock()
	buf           = bytes.Buffer{}
	sessionClient = session.Client{
		ID:       clientID,
		Username: thingID,
		Password: []byte(password),
	}
	invalidThingSessionClient = session.Client{
		ID:       clientID,
		Username: invalidThingID,
		Password: []byte(password),
	}
)

func TestAuthConnect(t *testing.T) {
	handler := newHandler()

	cases := []struct {
		desc    string
		err     error
		session *session.Client
	}{
		{
			desc:    "connect without active session",
			err:     mqtt.ErrClientNotInitialized,
			session: nil,
		},
		{
			desc: "connect without clientID",
			err:  mqtt.ErrMissingClientID,
			session: &session.Client{
				ID:       "",
				Username: thingID,
				Password: []byte(password),
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
	var payload = []byte("[{'n':'test-name', 'v': 1.2}]")

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
			err:     mqtt.ErrClientNotInitialized,
			topic:   &topic,
			payload: payload,
		},
		{
			desc:    "publish without topic",
			client:  &sessionClient,
			err:     mqtt.ErrMissingTopicPub,
			topic:   nil,
			payload: payload,
		},
		{
			desc:    "publish with malformed topic",
			client:  &sessionClient,
			err:     mqtt.ErrMalformedTopic,
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
	var invalidTopics = []string{invalidTopic}
	var topics = []string{"channels/" + chanID + "/messages"}
	invalidChanIDTopics := []string{"channels/" + invalidChanID + "/messages"}

	cases := []struct {
		desc   string
		client *session.Client
		err    error
		topic  *[]string
	}{
		{
			desc:   "subscribe without active session",
			client: nil,
			err:    mqtt.ErrClientNotInitialized,
			topic:  &topics,
		},
		{
			desc:   "subscribe without topics",
			client: &sessionClient,
			err:    mqtt.ErrMissingTopicSub,
			topic:  nil,
		},
		{
			desc:   "subscribe with invalid topics",
			client: &sessionClient,
			err:    mqtt.ErrMalformedTopic,
			topic:  &invalidTopics,
		},
		{
			desc:   "subscribe with invalid channel ID",
			client: &sessionClient,
			err:    mqtt.ErrAuthentication,
			topic:  &invalidChanIDTopics,
		},
		{
			desc:   "subscribe with invalid thing ID",
			client: &invalidThingSessionClient,
			err:    mqtt.ErrAuthentication,
			topic:  &topics,
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
			expected: fmt.Sprint(errors.Wrap(mqtt.ErrFailedConnect, mqtt.ErrClientNotInitialized).Error()),
		},
		{
			desc:     "connect with valid session",
			client:   &sessionClient,
			expected: fmt.Sprintf(mqtt.InfoConnected, clientID),
		},
	}

	for _, tc := range cases {
		handler.Connect(tc.client)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestPublish(t *testing.T) {
	handler := newHandler()
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
			expected: mqtt.ErrClientNotInitialized.Error(),
		},
		{
			desc:     "publish with invalid topic",
			client:   &sessionClient,
			topic:    invalidTopic,
			payload:  []byte("payload"),
			expected: fmt.Sprintf(mqtt.InfoPublished, clientID, invalidTopic),
		},
	}

	for _, tc := range cases {
		handler.Publish(tc.client, &tc.topic, &tc.payload)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestSubscribe(t *testing.T) {
	handler := newHandler()
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
			expected: fmt.Sprint(errors.Wrap(mqtt.ErrFailedSubscribe, mqtt.ErrClientNotInitialized).Error()),
		},
		{
			desc:     "subscribe with valid session and topics",
			client:   &sessionClient,
			topic:    topics,
			expected: fmt.Sprintf(mqtt.InfoSubscribed, clientID, topics[0]),
		},
	}

	for _, tc := range cases {
		handler.Subscribe(tc.client, &tc.topic)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestUnsubscribe(t *testing.T) {
	handler := newHandler()
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
			expected: fmt.Sprint(errors.Wrap(mqtt.ErrFailedUnsubscribe, mqtt.ErrClientNotInitialized).Error()),
		},
		{
			desc:     "unsubscribe with valid session and topics",
			client:   &sessionClient,
			topic:    topics,
			expected: fmt.Sprintf(mqtt.InfoUnsubscribed, clientID, topics[0]),
		},
	}

	for _, tc := range cases {
		handler.Unsubscribe(tc.client, &tc.topic)
		assert.Contains(t, buf.String(), tc.expected)
	}
}

func TestDisconnect(t *testing.T) {
	handler := newHandler()
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
			expected: fmt.Sprint(errors.Wrap(mqtt.ErrFailedDisconnect, mqtt.ErrClientNotInitialized).Error()),
		},
		{
			desc:     "disconect with valid session",
			client:   &sessionClient,
			topic:    topics,
			expected: fmt.Sprintf(mqtt.InfoDisconnected, clientID, thingID),
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

	authClient := mocks.NewClient(map[string]string{password: thingID}, map[string]interface{}{chanID: thingID})

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

	return mqtt.NewHandler([]messaging.Publisher{mocks.NewPublisher()}, eventStore, logger, authClient)
}

package mqtt

import (
	"fmt"
	"github.com/mainflux/mainflux/pkg/uuid"
	"log"
	"testing"

	"github.com/mainflux/mainflux/pkg/errors"

	rdb "github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt/mocks"
	"github.com/mainflux/mainflux/mqtt/redis"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

func TestAuthConnect(t *testing.T) {
	cases := []struct {
		desc    string
		err     error
		session *session.Client
	}{
		{
			desc:    "Handle message without active session",
			err:     errInvalidConnect,
			session: nil,
		},
		{
			desc: "Handle message when thing id is invalid",
			err:  errors.New("thing identify error"),
			session: &session.Client{
				ID:       "123",
				Username: "testUsername",
				Password: []byte(""),
			},
		},
		{
			desc: "Handle message when username is wrong",
			err:  errors.ErrAuthentication,
			session: &session.Client{
				ID:       "123",
				Username: "testUsername",
				Password: []byte("password"),
			},
		},
		{
			desc: "Handle message correctly",
			err:  nil,
			session: &session.Client{
				ID:       "123",
				Username: "ok",
				Password: []byte("password"),
			},
		},
	}

	for _, tc := range cases {
		err := setupTest().AuthConnect(tc.session)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestAuthPublish(t *testing.T) {
	var topic string
	var payload = []byte("payload")

	sessionClient := session.Client{
		ID:       "123",
		Username: "username",
		Password: []byte("password"),
	}

	cases := []struct {
		desc    string
		client  *session.Client
		err     error
		topic   *string
		payload []byte
	}{
		{
			desc:    "Handle publish authorization on client without active session",
			client:  nil,
			err:     errNilClient,
			topic:   &topic,
			payload: payload,
		},
		{
			desc:    "Handle publish authorization on client without topic",
			client:  &sessionClient,
			err:     errNilTopicPub,
			topic:   nil,
			payload: payload,
		},
		{
			desc:    "Handle publish authorization on client correctly",
			client:  &sessionClient,
			err:     errMalformedTopic,
			topic:   &topic,
			payload: payload,
		},
	}

	for _, tc := range cases {
		err := setupTest().AuthPublish(tc.client, tc.topic, &tc.payload)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestAuthSubscribe(t *testing.T) {
	var topics = []string{"channels/1/messages/2/ct/3"}
	var invalidTopics = []string{"topic"}

	sessionClient := session.Client{
		ID:       "123",
		Username: "username",
		Password: []byte("password"),
	}

	cases := []struct {
		desc   string
		client *session.Client
		err    error
		topic  *[]string
	}{
		{
			desc:   "Handle subscribe authorization on client without active session",
			client: nil,
			err:    errNilClient,
			topic:  &topics,
		},
		{
			desc:   "Handle subscribe authorization on client without topics",
			client: &sessionClient,
			err:    errNilTopicSub,
			topic:  nil,
		},
		{
			desc:   "Handle subscribe authorization on client with invalid channel",
			client: &sessionClient,
			err:    errMalformedTopic,
			topic:  &invalidTopics,
		},
		{
			desc:   "Handle subscribe authorization on client correctly",
			client: &sessionClient,
			err:    nil,
			topic:  &topics,
		},
	}

	for _, tc := range cases {
		err := setupTest().AuthSubscribe(tc.client, tc.topic)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func setupTest() session.Handler {
	idProvider := uuid.NewMock()
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
	return NewHandler([]messaging.Publisher{mocks.NewPublisher()}, eventStore, logger.NewMock(), mocks.NewClient(), idProvider)
}

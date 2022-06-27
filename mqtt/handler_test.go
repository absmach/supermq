package mqtt

import (
	"fmt"
	"log"
	"testing"

	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/mainflux/mproxy/pkg/session"
	"github.com/stretchr/testify/require"

	rdb "github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt/mocks"
	"github.com/mainflux/mainflux/mqtt/redis"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

func TestHandler_AuthConnect(t *testing.T) {

	cases := []struct {
		desc    string
		err     error
		session *session.Client
	}{
		{
			desc:    "should return error if client is nil",
			err:     errInvalidConnect,
			session: nil,
		},
		{
			desc: "should return error if thing id is invalid",
			err:  errors.New("thing identify error"),
			session: &session.Client{
				ID:       "123",
				Username: "testUsername",
				Password: []byte(""),
			},
		},
		{
			desc: "should return error if username is wrong",
			err:  errors.ErrAuthentication,
			session: &session.Client{
				ID:       "123",
				Username: "testUsername",
				Password: []byte("password"),
			},
		},
		{
			desc: "should not return error",
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

func TestHandler_AuthPublish(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var topic = "topic"
	var payload = []byte("payload")

	sessionClient := session.Client{
		ID:       "123",
		Username: "username",
		Password: []byte("password"),
	}

	t.Run("should return error if client is nil", func(t *testing.T) {
		err := setupTest().AuthPublish(nil, &topic, &payload)
		assert.Equal(err, errNilClient)
	})

	t.Run("should return error if topic is nil", func(t *testing.T) {
		err := setupTest().AuthPublish(&sessionClient, nil, &payload)
		assert.Equal(err, errNilTopicPub)
	})

	t.Run("authpublish should run ok", func(t *testing.T) {
		require.NotNil(&sessionClient)
		require.NotNil(&topic)
		err := setupTest().AuthPublish(&sessionClient, &topic, &payload)
		assert.Error(err)

	})

}

func TestHandler_AuthSubscribe(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	var topics = []string{"channels/1/messages/2/ct/3"}
	var invalidTopics = []string{"topic"}

	sessionClient := session.Client{
		ID:       "123",
		Username: "username",
		Password: []byte("password"),
	}

	t.Run("should return error if client is nil", func(t *testing.T) {
		err := setupTest().AuthSubscribe(nil, &topics)
		assert.Equal(err, errNilClient)
	})

	t.Run("should return error if topics is nil", func(t *testing.T) {
		require.NotNil(&sessionClient)
		err := setupTest().AuthSubscribe(&sessionClient, nil)
		assert.Equal(err, errNilTopicSub)
	})

	t.Run("should return auth acess error", func(t *testing.T) {
		require.NotNil(&sessionClient)
		require.NotNil(&topics)
		err := setupTest().AuthSubscribe(&sessionClient, &invalidTopics)
		assert.Error(err)
	})

	t.Run("auth subscribe should run ok", func(t *testing.T) {
		require.NotNil(&sessionClient)
		require.NotNil(&topics)
		require.NotNil(topics)
		err := setupTest().AuthSubscribe(&sessionClient, &topics)
		assert.Nil(err)
	})
}

func setupTest() session.Handler {
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
	return NewHandler([]messaging.Publisher{mocks.NewPublisher()}, eventStore, logger.NewMock(), mocks.NewClient())

}

package redis_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/things/redis"
	"github.com/mainflux/mainflux/things/uuid"
	"github.com/stretchr/testify/assert"
)

func TestThingSave(t *testing.T) {
	thingCache := redis.NewThingCache(client)
	key := uuid.New().ID()
	id := u

	err := thingCache.Save(key, id)
	assert.Nil(t, err, fmt.Sprintf("save new thing to cache: expected no error got %s\n", err))
}

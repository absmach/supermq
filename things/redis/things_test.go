//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package redis_test

import (
	"fmt"
	"testing"

	r "github.com/go-redis/redis"
	"github.com/mainflux/mainflux/things/redis"
	"github.com/mainflux/mainflux/things/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThingSave(t *testing.T) {

	thingCache := redis.NewThingCache(cacheClient)
	key := uuid.New().ID()
	id := uint64(123)

	// show that the save works the same for both non-cached and cached thing
	for i := 0; i < 2; i++ {
		err := thingCache.Save(key, id)
		require.Nil(t, err, fmt.Sprintf("#%d: save thing to cache: expected no error got %s\n", i, err))
	}
}

func TestThingID(t *testing.T) {
	thingCache := redis.NewThingCache(cacheClient)

	key := uuid.New().ID()
	id := uint64(123)
	thingCache.Save(key, id)

	cases := []struct {
		desc string
		ID   uint64
		key  string
		err  error
	}{
		{desc: "Get ID by existing thing-key", ID: id, key: key, err: nil},
		{desc: "Get ID by non-existing thing-key", ID: 0, key: wrongValue, err: r.Nil},
	}

	for _, tc := range cases {
		cacheID, err := thingCache.ID(tc.key)
		assert.Equal(t, tc.ID, cacheID, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.ID, cacheID))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestThingRemove(t *testing.T) {
	thingCache := redis.NewThingCache(cacheClient)

	key := uuid.New().ID()
	id := uint64(123)
	thingCache.Save(key, id)

	cases := []struct {
		desc string
		ID   uint64
		err  error
	}{
		{desc: "Remove existing thing from cache", ID: id, err: nil},
		{desc: "Remove removed thing from cache", ID: id, err: r.Nil},
	}

	for _, tc := range cases {
		err := thingCache.Remove(tc.ID)
		require.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}

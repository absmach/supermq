// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/mocks"
	"github.com/mainflux/mainflux/twins/redis"
	"github.com/mainflux/mainflux/twins/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var names = []string{"temperature", "humidity", "speed"}
var subtopics = []string{"engine", "chassis", "wheel_2"}
var channels = []string{"01ec3c3e-0e66-4e69-9751-a0545b44e08f", "48061e4f-7c23-4f5c-9012-0f9b7cd9d18d", "5b2180e4-e96b-4469-9dc1-b6745078d0b6"}

func TestTwinSave(t *testing.T) {
	redisClient.FlushAll()
	twinCache := redis.NewTwinCache(redisClient)

	twin1, err := createTwin(names[0:2], channels[0:2], subtopics[0:2])
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	twin2, err := createTwin(names[1:3], channels[1:3], subtopics[1:3])
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := []struct {
		desc string
		twin twins.Twin
		err  error
	}{
		{
			desc: "Save twin to cache",
			twin: twin1,
			err:  nil,
		},
		{
			desc: "Save already cached twin to cache",
			twin: twin1,
			err:  nil,
		},
		{
			desc: "Save another twin to cache",
			twin: twin2,
			err:  nil,
		},
		{
			desc: "Save already cached twin to cache",
			twin: twin2,
			err:  nil,
		},
	}

	for _, tc := range cases {
		ctx := context.Background()
		err := twinCache.Save(ctx, tc.twin)
		assert.Nil(t, err, fmt.Sprintf("%s: expected %s got %s", tc.desc, tc.err, err))

		def := tc.twin.Definitions[len(tc.twin.Definitions)-1]
		for _, attr := range def.Attributes {
			ids, err := twinCache.IDs(ctx, attr)
			require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
			assert.Contains(t, ids, tc.twin.ID, fmt.Sprintf("%s: id %s not found in %v", tc.desc, tc.twin.ID, ids))
		}
	}
}

func TestTwinIDs(t *testing.T) {
	redisClient.FlushAll()
	twinCache := redis.NewTwinCache(redisClient)
	ctx := context.Background()

	var tws []twins.Twin
	for i := range names {
		tw, err := createTwin(names[i:i+1], channels[0:1], subtopics[0:1])
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		err = twinCache.Save(ctx, tw)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		tws = append(tws, tw)
	}
	for i := range names {
		tw, err := createTwin(names[i:i+1], channels[1:2], subtopics[1:2])
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		err = twinCache.Save(ctx, tw)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		tws = append(tws, tw)
	}

	nonExistAttr := twins.Attribute{
		Name:         names[0],
		Channel:      channels[2],
		Subtopic:     subtopics[0],
		PersistState: true,
	}

	cases := []struct {
		desc string
		ids  []string
		attr twins.Attribute
		err  error
	}{
		{
			desc: "Get twin IDs from cache for subset of ids",
			ids:  []string{tws[0].ID, tws[1].ID, tws[2].ID},
			attr: tws[0].Definitions[0].Attributes[0],
			err:  nil,
		},
		{
			desc: "Get twin IDs from cache for subset of ids",
			ids:  []string{tws[3].ID, tws[4].ID, tws[5].ID},
			attr: tws[3].Definitions[0].Attributes[0],
			err:  nil,
		},
		{
			desc: "Get twin IDs from cache for non existing attribute",
			ids:  []string{},
			attr: nonExistAttr,
			err:  nil,
		},
	}

	for _, tc := range cases {
		ids, err := twinCache.IDs(ctx, tc.attr)
		assert.Nil(t, err, fmt.Sprintf("%s: expected %s got %s", tc.desc, tc.err, err))
		assert.ElementsMatch(t, ids, tc.ids, fmt.Sprintf("%s: expected ids %v got ids %v", tc.desc, tc.ids, ids))
	}
}

func TestTwinRemove(t *testing.T) {
	redisClient.FlushAll()
	twinCache := redis.NewTwinCache(redisClient)
	ctx := context.Background()

	var tws []twins.Twin
	for i := range names {
		tw, err := createTwin(names[i:i+1], channels[i:i+1], subtopics[i:i+1])
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		err = twinCache.Save(ctx, tw)
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		tws = append(tws, tw)
	}

	cases := []struct {
		desc string
		twin twins.Twin
		err  error
	}{
		{
			desc: "Remove twin from cache",
			twin: tws[0],
			err:  nil,
		},
		{
			desc: "Remove already removed twin from cache",
			twin: tws[0],
			err:  nil,
		},
		{
			desc: "Remove another twin from cache",
			twin: tws[1],
			err:  nil,
		},
	}

	for _, tc := range cases {
		err := twinCache.Remove(ctx, tc.twin)
		assert.Nil(t, err, fmt.Sprintf("%s: expected %s got %s", tc.desc, tc.err, err))

		def := tc.twin.Definitions[len(tc.twin.Definitions)-1]
		for _, attr := range def.Attributes {
			ids, err := twinCache.IDs(ctx, attr)
			require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
			assert.NotContains(t, ids, tc.twin.ID, fmt.Sprintf("%s: id %s not found in %v", tc.desc, tc.twin.ID, ids))
		}
	}
}

func createTwin(names []string, channels []string, subtopics []string) (twins.Twin, error) {
	id, err := uuid.New().ID()
	if err != nil {
		return twins.Twin{}, err
	}
	return twins.Twin{
		ID:          id,
		Definitions: []twins.Definition{mocks.CreateDefinition(names, channels, subtopics)},
	}, nil
}

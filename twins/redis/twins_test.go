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

const (
	attrName1     = "temperature"
	attrSubtopic1 = "engine"
	attrName2     = "humidity"
	attrSubtopic2 = "chassis"
	attrName3     = "speed"
	attrSubtopic3 = "wheel_2"
)

func TestTwinSave(t *testing.T) {
	twinCache := redis.NewTwinCache(redisClient)

	twin1, err := createTwin([]string{attrName1, attrName2}, []string{attrSubtopic1, attrSubtopic2})
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	twin2, err := createTwin([]string{attrName2, attrName3}, []string{attrSubtopic2, attrSubtopic3})
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

func createTwin(attrNames []string, subtopics []string) (twins.Twin, error) {
	id, err := uuid.New().ID()
	if err != nil {
		return twins.Twin{}, err
	}
	return twins.Twin{
		ID:          id,
		Definitions: []twins.Definition{mocks.CreateDefinition(attrNames, subtopics)},
	}, nil
}

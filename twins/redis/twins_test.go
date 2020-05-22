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
	"github.com/stretchr/testify/assert"
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

	def := mocks.CreateDefinition([]string{attrName1, attrName2}, []string{attrSubtopic1, attrSubtopic2})
	twin := twins.Twin{
		Definitions: []twins.Definition{def},
	}

	cases := []struct {
		desc string
		twin twins.Twin
		err  error
	}{
		{
			desc: "Save twin to cache",
			twin: twin,
			err:  nil,
		},
		{
			desc: "Save already cached twin to cache",
			twin: twin,
			err:  nil,
		},
	}

	for _, tc := range cases {
		err := twinCache.Save(context.Background(), tc.twin)
		assert.Nil(t, err, fmt.Sprintf("%s: expected %s got %s", tc.desc, tc.err, err))

	}
}

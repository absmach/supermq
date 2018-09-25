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

	"github.com/mainflux/mainflux/things/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect(t *testing.T) {
	channelCache := redis.NewChannelCache(cacheClient)

	cid := uint64(123)
	tid := uint64(321)

	err := channelCache.Connect(cid, tid)
	assert.Nil(t, err, fmt.Sprintf("connect new thing to channel: expected no error got %s\n", err))
}

func TestHasThing(t *testing.T) {
	channelCache := redis.NewChannelCache(cacheClient)

	cid := uint64(123)
	tid := uint64(321)

	channelCache.Connect(cid, tid)

	cases := []struct {
		desc      string
		cid       uint64
		tid       uint64
		hasAccess bool
	}{
		{desc: "access check for thing that has access", cid: cid, tid: tid, hasAccess: true},
		{desc: "access check for thing without access", cid: cid, tid: cid, hasAccess: false},
		{desc: "access check for non-existing channel", cid: tid, tid: tid, hasAccess: false},
	}

	for _, tc := range cases {
		hasAccess := channelCache.HasThing(tc.cid, tc.tid)
		assert.Equal(t, tc.hasAccess, hasAccess, fmt.Sprintf("%s: expected %t got %t\n", tc.desc, tc.hasAccess, hasAccess))
	}
}
func TestDisconnect(t *testing.T) {
	channelCache := redis.NewChannelCache(cacheClient)

	cid := uint64(123)
	tid := uint64(321)

	channelCache.Disconnect(cid, tid)
	hasAccess := channelCache.HasThing(cid, tid)
	assert.Equal(t, false, hasAccess, fmt.Sprintf("Check access after disconnect unconnected thing from channel: expected false got %t\n", hasAccess))

	channelCache.Connect(cid, tid)
	hasAccess = channelCache.HasThing(cid, tid)
	assert.Equal(t, true, hasAccess, fmt.Sprintf("Check access after connecting thing to channel: expected true got %t\n", hasAccess))

	channelCache.Disconnect(cid, tid)
	hasAccess = channelCache.HasThing(cid, tid)
	assert.Equal(t, false, hasAccess, fmt.Sprintf("Check access after disconnecting thing from channel: expected false got %t\n", hasAccess))
}

func TestRemove(t *testing.T) {
	channelCache := redis.NewChannelCache(cacheClient)

	cid := uint64(123)
	tid := uint64(321)

	channelCache.Connect(cid, tid)
	hasAccess := channelCache.HasThing(cid, tid)
	require.Equal(t, true, hasAccess, fmt.Sprintf("Check access after connecting thing to channel: expected true got %t\n", hasAccess))

	channelCache.Remove(cid)
	hasAccess = channelCache.HasThing(cid, tid)
	require.Equal(t, false, hasAccess, fmt.Sprintf("Check access after removing channel: expected false got %t\n", hasAccess))
}

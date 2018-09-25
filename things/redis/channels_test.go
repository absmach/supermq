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

	_ = channelCache.Connect(cid, tid)

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

	/*
		_ = channelCache.Connect(cid, tid)

		cases := []struct {
			desc string
			cid  uint64
			tid  uint64
			err  error
		}{
			{desc: "disconnect connected thing", cid: cid, tid: tid, err: nil},
			{desc: "disconnect non-connected thing", cid: cid, tid: cid, err: nil},
			{desc: "disconnect non-existing channel", cid: tid, tid: tid, err: nil}, //some error
		}

		for _, tc := range cases {
			err := channelCache.Disconnect(tc.cid, tc.tid)
			assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		}
	*/

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

	/*
		cases := []struct {
			desc string
			cid  uint64
			err  error
		}{
			{desc: "Remove existing thing from cache", cid: cid, err: nil},
			{desc: "Remove removed thing from cache", cid: cid, err: r.Nil},
		}

		for _, tc := range cases {
			err := channelCache.Remove(tc.cid)
			require.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		}
	*/

}

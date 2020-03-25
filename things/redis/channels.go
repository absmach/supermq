// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/things"
)

const chanPrefix = "channel"

// ErrRedisConnectChannel indicates error while adding connection in redis cache
var ErrRedisConnectChannel = errors.New("add connection to redis cache error")

// ErrRedisDisconnectChannel indicates error while removing connection from redis cache
var ErrRedisDisconnectChannel = errors.New("remove connection from redis cache error")

// ErrRedisRemoveChannel indicates error while removing channel from redis cache
var ErrRedisRemoveChannel = errors.New("remove channel from redis cache error")

var _ things.ChannelCache = (*channelCache)(nil)

type channelCache struct {
	client *redis.Client
}

// NewChannelCache returns redis channel cache implementation.
func NewChannelCache(client *redis.Client) things.ChannelCache {
	return channelCache{client: client}
}

func (cc channelCache) Connect(_ context.Context, chanID, thingID string) errors.Error {
	cid, tid := kv(chanID, thingID)
	err := cc.client.SAdd(cid, tid).Err()
	return errors.Wrap(ErrRedisConnectChannel, err)
}

func (cc channelCache) HasThing(_ context.Context, chanID, thingID string) bool {
	cid, tid := kv(chanID, thingID)
	return cc.client.SIsMember(cid, tid).Val()
}

func (cc channelCache) Disconnect(_ context.Context, chanID, thingID string) errors.Error {
	cid, tid := kv(chanID, thingID)
	err := cc.client.SRem(cid, tid).Err()
	return errors.Wrap(ErrRedisDisconnectChannel, err)
}

func (cc channelCache) Remove(_ context.Context, chanID string) errors.Error {
	cid, _ := kv(chanID, "0")
	err := cc.client.Del(cid).Err()
	return errors.Wrap(ErrRedisRemoveChannel, err)
}

// Generates key-value pair
func kv(chanID, thingID string) (string, string) {
	cid := fmt.Sprintf("%s:%s", chanPrefix, chanID)
	return cid, thingID
}

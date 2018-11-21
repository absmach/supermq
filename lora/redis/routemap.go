//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package redis

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/lora"
)

const (
	loraAppPrefix   = "lora_app"
	mfxChanelPrefix = "mfx_channel"
)

var _ lora.RouteMapRepository = (*routerMap)(nil)

type routerMap struct {
	client *redis.Client
}

// NewRouteMapRepository returns redis thing cache implementation.
func NewRouteMapRepository(client *redis.Client) lora.RouteMapRepository {
	return &routerMap{
		client: client,
	}
}

func (mr *routerMap) Save(app string, channel uint64) error {
	tkey := fmt.Sprintf("%s:%s", loraAppPrefix, app)
	ch := strconv.FormatUint(channel, 10)
	if err := mr.client.Set(tkey, ch, 0).Err(); err != nil {
		return err
	}

	tid := fmt.Sprintf("%s:%s", mfxChanelPrefix, ch)
	return mr.client.Set(tid, app, 0).Err()
}

func (mr *routerMap) Channel(loraApp string) (uint64, error) {
	laKey := fmt.Sprintf("%s:%s", loraAppPrefix, loraApp)
	mfxChannel, err := mr.client.Get(laKey).Result()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseUint(mfxChannel, 10, 64)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (mr *routerMap) Remove(channel uint64) error {
	tid := fmt.Sprintf("%s:%d", mfxChanelPrefix, channel)
	key, err := mr.client.Get(tid).Result()
	if err != nil {
		return err
	}

	tkey := fmt.Sprintf("%s:%s", loraAppPrefix, key)

	return mr.client.Del(tkey, tid).Err()
}

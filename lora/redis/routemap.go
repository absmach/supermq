//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package redis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/lora"
)

const (
	loraMapPrefix = "lora:mfx"
	mfxMapPrefix  = "mfx:lora"
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

func (mr *routerMap) Save(mfxID string, loraID string) error {
	tkey := fmt.Sprintf("%s:%s", mfxMapPrefix, mfxID)
	if err := mr.client.Set(tkey, loraID, 0).Err(); err != nil {
		return err
	}
	lkey := fmt.Sprintf("%s:%s", loraMapPrefix, loraID)
	if err := mr.client.Set(lkey, mfxID, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (mr *routerMap) Get(mfxID string) (string, error) {
	lKey := fmt.Sprintf("%s:%s", loraMapPrefix, mfxID)
	mval, err := mr.client.Get(lKey).Result()
	if err != nil {
		return "", err
	}

	return mval, nil
}

func (mr *routerMap) Remove(mfxID string) error {
	mkey := fmt.Sprintf("%s:%s", mfxMapPrefix, mfxID)
	lval, err := mr.client.Get(mkey).Result()
	if err != nil {
		return err
	}

	lkey := fmt.Sprintf("%s:%s", loraMapPrefix, lval)
	return mr.client.Del(mkey, lkey).Err()
}

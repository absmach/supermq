// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/twins"
)

const (
	prefix = "twin"
)

// ErrRedisTwinSave indicates error while saving Twin in redis cache
var ErrRedisTwinSave = errors.New("saving twin in redis cache error")

// ErrRedisTwinIDs indicates error while geting Twin IDs from redis cache
var ErrRedisTwinIDs = errors.New("get twin id from redis cache error")

// ErrRedisTwinRemove indicates error while removing Twin from redis cache
var ErrRedisTwinRemove = errors.New("remove twin from redis cache error")

var _ twins.TwinCache = (*twinCache)(nil)

type twinCache struct {
	client *redis.Client
}

// NewTwinCache returns redis twin cache implementation.
func NewTwinCache(client *redis.Client) twins.TwinCache {
	return &twinCache{
		client: client,
	}
}

func (tc *twinCache) Save(_ context.Context, twin twins.Twin) error {
	def := twin.Definitions[len(twin.Definitions)-1]
	for _, attr := range def.Attributes {
		if err := tc.client.SAdd(attrKey(attr), twin.ID).Err(); err != nil {
			return errors.Wrap(ErrRedisTwinSave, err)
		}
		if err := tc.client.SAdd(twinKey(twin.ID), attrVal(attr)).Err(); err != nil {
			return errors.Wrap(ErrRedisTwinSave, err)
		}
	}
	return nil
}

func (tc *twinCache) IDs(_ context.Context, attr twins.Attribute) ([]string, error) {
	ids, err := tc.client.SMembers(attrKey(attr)).Result()
	if err != nil {
		return nil, errors.Wrap(ErrRedisTwinIDs, err)
	}
	return ids, nil
}

func (tc *twinCache) Remove(_ context.Context, twin twins.Twin) error {
	if err := tc.client.Del(twinKey(twin.ID)).Err(); err != nil {
		return errors.Wrap(ErrRedisTwinRemove, err)
	}
	def := twin.Definitions[len(twin.Definitions)-1]
	for _, attr := range def.Attributes {
		if err := tc.client.SRem(attrKey(attr), twin.ID).Err(); err != nil {
			return errors.Wrap(ErrRedisTwinRemove, err)
		}
	}
	return nil
}

func twinKey(twinID string) string {
	return fmt.Sprintf("%s:%s", prefix, twinID)
}

func attrKey(attr twins.Attribute) string {
	return fmt.Sprintf("%s:%s-%s", prefix, attr.Channel, attr.Subtopic)
}

func attrVal(attr twins.Attribute) string {
	return fmt.Sprintf("%s-%s", attr.Channel, attr.Subtopic)
}

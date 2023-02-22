// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

const groupPrefix = "group"

var _ policies.Cache = (*pcache)(nil)

type pcache struct {
	client *redis.Client
}

// NewCache returns redis policy cache implementation.
func NewCache(client *redis.Client) policies.Cache {
	return pcache{client: client}
}

func (cc pcache) AddPolicy(ctx context.Context, policy policies.Policy) error {
	obj, subs := kv(policy)
	for _, sub := range subs {
		if err := cc.client.SAdd(ctx, obj, sub).Err(); err != nil {
			return errors.Wrap(errors.ErrCreateEntity, err)
		}
	}
	return nil
}

func (cc pcache) Evaluate(ctx context.Context, policy policies.Policy) bool {
	obj, subs := kv(policy)
	return cc.client.SIsMember(ctx, obj, subs[0]).Val()
}

func (cc pcache) DeletePolicy(ctx context.Context, policy policies.Policy) error {
	obj, _ := kv(policy)
	if err := cc.client.Del(ctx, obj).Err(); err != nil {
		return errors.Wrap(errors.ErrRemoveEntity, err)
	}
	return nil
}

// Generates key-value pair
func kv(policy policies.Policy) (string, []string) {
	var subs []string
	for _, a := range policy.Actions {
		subs = append(subs, fmt.Sprintf("%s:%s", policy.Subject, a))
	}
	return fmt.Sprintf("%s:%s", groupPrefix, policy.Object), subs
}

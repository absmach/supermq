// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

// Client represents Auth cache.
type Client interface {
	Authorize(ctx context.Context, chanID, thingID, action string) error
	Identify(ctx context.Context, thingKey string) (string, error)
}

const (
	chanPrefix = "channel"
	keyPrefix  = "thing_key"
)

type client struct {
	redisClient  *redis.Client
	thingsClient policies.ThingsServiceClient
}

// New returns redis channel cache implementation.
func New(redisClient *redis.Client, thingsClient policies.ThingsServiceClient) Client {
	return client{
		redisClient:  redisClient,
		thingsClient: thingsClient,
	}
}

func (c client) Identify(ctx context.Context, thingKey string) (string, error) {
	tkey := keyPrefix + ":" + thingKey
	thingID, err := c.redisClient.Get(ctx, tkey).Result()
	if err != nil {
		t := &policies.Key{
			Value: string(thingKey),
		}

		thid, err := c.thingsClient.Identify(context.TODO(), t)
		if err != nil {
			return "", err
		}
		return thid.GetValue(), nil
	}
	return thingID, nil
}

func (c client) Authorize(ctx context.Context, chanID, thingID, action string) error {
	if c.redisClient.SIsMember(ctx, chanPrefix+":"+chanID, thingID).Val() {
		return nil
	}

	ar := &policies.AuthorizeReq{
		Sub:        thingID,
		Obj:        chanID,
		Act:        action,
		EntityType: policies.ThingEntityType,
	}
	res, err := c.thingsClient.Authorize(ctx, ar)
	if err != nil {
		return err
	}
	if !res.GetAuthorized() {
		return errors.ErrAuthorization
	}

	return err
}

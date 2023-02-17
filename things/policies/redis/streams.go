// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/things/policies"
)

const (
	streamID  = "mainflux.things"
	streamLen = 1000
)

var _ policies.Service = (*eventStore)(nil)

type eventStore struct {
	svc    policies.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(svc policies.Service, client *redis.Client) policies.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) AddPolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.AddPolicy(ctx, token, policy); err != nil {
		return err
	}

	event := connectThingEvent{
		chanID:  policy.Object,
		thingID: policy.Subject,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return nil
}

func (es eventStore) DeletePolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.DeletePolicy(ctx, token, policy); err != nil {
		return err
	}

	event := disconnectThingEvent{
		chanID:  policy.Object,
		thingID: policy.Subject,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return nil
}

func (es eventStore) CanAccessByKey(ctx context.Context, chanID string, key string) (string, error) {
	return es.svc.CanAccessByKey(ctx, chanID, key)
}

func (es eventStore) CanAccessByID(ctx context.Context, chanID string, thingID string) error {
	return es.svc.CanAccessByID(ctx, chanID, thingID)
}

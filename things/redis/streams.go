// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/things"
)

const (
	streamID  = "mainflux.things"
	streamLen = 1000
)

var _ things.Service = (*eventStore)(nil)

type eventStore struct {
	svc    things.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(svc things.Service, client *redis.Client) things.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) CreateThings(ctx context.Context, token string, ths ...things.Thing) ([]things.Thing, error) {
	sths, err := es.svc.CreateThings(ctx, token, ths...)
	if err != nil {
		return sths, err
	}

	for _, thing := range sths {
		event := createThingEvent{
			id:       thing.ID,
			owner:    thing.Owner,
			name:     thing.Name,
			metadata: thing.Metadata,
		}
		record := &redis.XAddArgs{
			Stream:       streamID,
			MaxLenApprox: streamLen,
			Values:       event.Encode(),
		}
		es.client.XAdd(record).Err()
	}

	return sths, nil
}

func (es eventStore) UpdateThing(ctx context.Context, token string, thing things.Thing) error {
	if err := es.svc.UpdateThing(ctx, token, thing); err != nil {
		return err
	}

	event := updateThingEvent{
		id:       thing.ID,
		name:     thing.Name,
		metadata: thing.Metadata,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(record).Err()

	return nil
}

// UpdateKey doesn't send event because key shouldn't be sent over stream.
// Maybe we can start publishing this event at some point, without key value
// in order to notify adapters to disconnect connected things after key update.
func (es eventStore) UpdateKey(ctx context.Context, token, id, key string) error {
	return es.svc.UpdateKey(ctx, token, id, key)
}

func (es eventStore) ViewThing(ctx context.Context, token, id string) (things.Thing, error) {
	return es.svc.ViewThing(ctx, token, id)
}

func (es eventStore) ListThings(ctx context.Context, token string, offset, limit uint64, name string, metadata things.Metadata) (things.ThingsPage, error) {
	return es.svc.ListThings(ctx, token, offset, limit, name, metadata)
}

func (es eventStore) ListThingsByChannel(ctx context.Context, token, id string, offset, limit uint64) (things.ThingsPage, error) {
	return es.svc.ListThingsByChannel(ctx, token, id, offset, limit)
}

func (es eventStore) RemoveThing(ctx context.Context, token, id string) error {
	if err := es.svc.RemoveThing(ctx, token, id); err != nil {
		return err
	}

	event := removeThingEvent{
		id: id,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(record).Err()

	return nil
}

func (es eventStore) CreateChannels(ctx context.Context, token string, channels ...things.Channel) ([]things.Channel, error) {
	schs, err := es.svc.CreateChannels(ctx, token, channels...)
	if err != nil {
		return schs, err
	}

	for _, channel := range schs {
		event := createChannelEvent{
			id:       channel.ID,
			owner:    channel.Owner,
			name:     channel.Name,
			metadata: channel.Metadata,
		}
		record := &redis.XAddArgs{
			Stream:       streamID,
			MaxLenApprox: streamLen,
			Values:       event.Encode(),
		}
		es.client.XAdd(record).Err()
	}

	return schs, nil
}

func (es eventStore) UpdateChannel(ctx context.Context, token string, channel things.Channel) error {
	if err := es.svc.UpdateChannel(ctx, token, channel); err != nil {
		return err
	}

	event := updateChannelEvent{
		id:       channel.ID,
		name:     channel.Name,
		metadata: channel.Metadata,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(record).Err()

	return nil
}

func (es eventStore) ViewChannel(ctx context.Context, token, id string) (things.Channel, error) {
	return es.svc.ViewChannel(ctx, token, id)
}

func (es eventStore) ListChannels(ctx context.Context, token string, offset, limit uint64, name string, metadata things.Metadata) (things.ChannelsPage, error) {
	return es.svc.ListChannels(ctx, token, offset, limit, name, metadata)
}

func (es eventStore) ListChannelsByThing(ctx context.Context, token, id string, offset, limit uint64) (things.ChannelsPage, error) {
	return es.svc.ListChannelsByThing(ctx, token, id, offset, limit)
}

func (es eventStore) RemoveChannel(ctx context.Context, token, id string) error {
	if err := es.svc.RemoveChannel(ctx, token, id); err != nil {
		return err
	}

	event := removeChannelEvent{
		id: id,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(record).Err()

	return nil
}

func (es eventStore) Identify(ctx context.Context, key string) (string, error) {
	return es.svc.Identify(ctx, key)
}

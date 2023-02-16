// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/clients/clients"
)

const (
	streamID  = "mainflux.things"
	streamLen = 1000
)

var _ clients.Service = (*eventStore)(nil)

type eventStore struct {
	svc    clients.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(svc clients.Service, client *redis.Client) clients.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) CreateThing(ctx context.Context, token string, thing clients.Client) (clients.Client, error) {
	sths, err := es.svc.CreateThing(ctx, token, thing)
	if err != nil {
		return sths, err
	}

	event := createClientEvent{
		id:       sths.ID,
		owner:    sths.Owner,
		name:     sths.Name,
		metadata: sths.Metadata,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return sths, nil
}

func (es eventStore) UpdateClient(ctx context.Context, token string, thing clients.Client) (clients.Client, error) {
	cli, err := es.svc.UpdateClient(ctx, token, thing)
	if err != nil {
		return clients.Client{}, err
	}

	event := updateClientEvent{
		id:       cli.ID,
		name:     cli.Name,
		metadata: cli.Metadata,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return cli, nil
}

func (es eventStore) UpdateClientOwner(ctx context.Context, token string, thing clients.Client) (clients.Client, error) {
	cli, err := es.svc.UpdateClientOwner(ctx, token, thing)
	if err != nil {
		return clients.Client{}, err
	}

	event := updateClientEvent{
		owner: cli.Owner,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return cli, nil
}

func (es eventStore) UpdateClientTags(ctx context.Context, token string, thing clients.Client) (clients.Client, error) {
	cli, err := es.svc.UpdateClientTags(ctx, token, thing)
	if err != nil {
		return clients.Client{}, err
	}

	event := updateClientEvent{
		tags: cli.Tags,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return cli, nil
}

// UpdateClientSecret doesn't send event because key shouldn't be sent over stream.
// Maybe we can start publishing this event at some point, without key value
// in order to notify adapters to disconnect connected things after key update.
func (es eventStore) UpdateClientSecret(ctx context.Context, token, id, key string) (clients.Client, error) {
	return es.svc.UpdateClientSecret(ctx, token, id, key)
}

func (es eventStore) ShareThing(ctx context.Context, token, thingID string, actions, userIDs []string) error {
	return es.svc.ShareThing(ctx, token, thingID, actions, userIDs)
}

func (es eventStore) ViewClient(ctx context.Context, token, id string) (clients.Client, error) {
	return es.svc.ViewClient(ctx, token, id)
}

func (es eventStore) ListClients(ctx context.Context, token string, pm clients.Page) (clients.ClientsPage, error) {
	return es.svc.ListClients(ctx, token, pm)
}

func (es eventStore) ListThingsByChannel(ctx context.Context, token, chID string, pm clients.Page) (clients.MembersPage, error) {
	return es.svc.ListThingsByChannel(ctx, token, chID, pm)
}

func (es eventStore) EnableClient(ctx context.Context, token, id string) (clients.Client, error) {
	cli, err := es.svc.EnableClient(ctx, token, id)
	if err != nil {
		return clients.Client{}, err
	}

	event := removeClientEvent{
		id: id,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return cli, nil
}

func (es eventStore) DisableClient(ctx context.Context, token, id string) (clients.Client, error) {
	cli, err := es.svc.DisableClient(ctx, token, id)
	if err != nil {
		return clients.Client{}, err
	}

	event := removeClientEvent{
		id: id,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(ctx, record).Err()

	return cli, nil
}

func (es eventStore) Identify(ctx context.Context, key string) (string, error) {
	return es.svc.Identify(ctx, key)
}

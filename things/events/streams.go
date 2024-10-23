// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/absmach/magistrala/pkg/authn"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/events"
	"github.com/absmach/magistrala/pkg/events/store"
	rmEvents "github.com/absmach/magistrala/pkg/roles/rolemanager/events"
	"github.com/absmach/magistrala/things"
)

const streamID = "magistrala.things"

var _ things.Service = (*eventStore)(nil)

type eventStore struct {
	events.Publisher
	svc things.Service
	rmEvents.RoleManagerEventStore
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc things.Service, url string) (things.Service, error) {
	publisher, err := store.NewPublisher(ctx, url, streamID)
	if err != nil {
		return nil, err
	}
	res := rmEvents.NewRoleManagerEventStore("domains", svc, publisher)

	return &eventStore{
		svc:                   svc,
		Publisher:             publisher,
		RoleManagerEventStore: res,
	}, nil
}

func (es *eventStore) CreateThings(ctx context.Context, session authn.Session, thing ...mgclients.Client) ([]mgclients.Client, error) {
	sths, err := es.svc.CreateThings(ctx, session, thing...)
	if err != nil {
		return sths, err
	}

	for _, th := range sths {
		event := createClientEvent{
			th,
		}
		if err := es.Publish(ctx, event); err != nil {
			return sths, err
		}
	}

	return sths, nil
}

func (es *eventStore) UpdateClient(ctx context.Context, session authn.Session, thing mgclients.Client) (mgclients.Client, error) {
	cli, err := es.svc.UpdateClient(ctx, session, thing)
	if err != nil {
		return cli, err
	}

	return es.updateThing(ctx, "", cli)
}

func (es *eventStore) UpdateClientTags(ctx context.Context, session authn.Session, thing mgclients.Client) (mgclients.Client, error) {
	cli, err := es.svc.UpdateClientTags(ctx, session, thing)
	if err != nil {
		return cli, err
	}

	return es.updateThing(ctx, "tags", cli)
}

func (es *eventStore) UpdateClientSecret(ctx context.Context, session authn.Session, id, key string) (mgclients.Client, error) {
	cli, err := es.svc.UpdateClientSecret(ctx, session, id, key)
	if err != nil {
		return cli, err
	}

	return es.updateThing(ctx, "secret", cli)
}

func (es *eventStore) updateThing(ctx context.Context, operation string, thing mgclients.Client) (mgclients.Client, error) {
	event := updateClientEvent{
		thing, operation,
	}

	if err := es.Publish(ctx, event); err != nil {
		return thing, err
	}

	return thing, nil
}

func (es *eventStore) ViewClient(ctx context.Context, session authn.Session, id string) (mgclients.Client, error) {
	cli, err := es.svc.ViewClient(ctx, session, id)
	if err != nil {
		return cli, err
	}

	event := viewClientEvent{
		cli,
	}
	if err := es.Publish(ctx, event); err != nil {
		return cli, err
	}

	return cli, nil
}

func (es *eventStore) ListClients(ctx context.Context, session authn.Session, reqUserID string, pm mgclients.Page) (mgclients.ClientsPage, error) {
	cp, err := es.svc.ListClients(ctx, session, reqUserID, pm)
	if err != nil {
		return cp, err
	}
	event := listClientEvent{
		reqUserID,
		pm,
	}
	if err := es.Publish(ctx, event); err != nil {
		return cp, err
	}

	return cp, nil
}

func (es *eventStore) EnableClient(ctx context.Context, session authn.Session, id string) (mgclients.Client, error) {
	cli, err := es.svc.EnableClient(ctx, session, id)
	if err != nil {
		return cli, err
	}

	return es.changeThingStatus(ctx, cli)
}

func (es *eventStore) DisableClient(ctx context.Context, session authn.Session, id string) (mgclients.Client, error) {
	cli, err := es.svc.DisableClient(ctx, session, id)
	if err != nil {
		return cli, err
	}

	return es.changeThingStatus(ctx, cli)
}

func (es *eventStore) changeThingStatus(ctx context.Context, cli mgclients.Client) (mgclients.Client, error) {
	event := changeStatusClientEvent{
		id:        cli.ID,
		updatedAt: cli.UpdatedAt,
		updatedBy: cli.UpdatedBy,
		status:    cli.Status.String(),
	}
	if err := es.Publish(ctx, event); err != nil {
		return cli, err
	}

	return cli, nil
}

func (es *eventStore) Identify(ctx context.Context, key string) (string, error) {
	thingID, err := es.svc.Identify(ctx, key)
	if err != nil {
		return thingID, err
	}
	event := identifyClientEvent{
		thingID: thingID,
	}

	if err := es.Publish(ctx, event); err != nil {
		return thingID, err
	}
	return thingID, nil
}

func (es *eventStore) Authorize(ctx context.Context, req things.AuthzReq) (string, error) {
	thingID, err := es.svc.Authorize(ctx, req)
	if err != nil {
		return thingID, err
	}

	event := authorizeClientEvent{
		thingID:    thingID,
		channelID:  req.ChannelID,
		permission: req.Permission,
	}

	if err := es.Publish(ctx, event); err != nil {
		return thingID, err
	}

	return thingID, nil
}

func (es *eventStore) DeleteClient(ctx context.Context, session authn.Session, id string) error {
	if err := es.svc.DeleteClient(ctx, session, id); err != nil {
		return err
	}

	event := removeClientEvent{id}

	if err := es.Publish(ctx, event); err != nil {
		return err
	}

	return nil
}

func (es *eventStore) RetrieveById(ctx context.Context, id string) (mgclients.Client, error) {
	return es.svc.RetrieveById(ctx, id)
}
func (es *eventStore) RetrieveByIds(ctx context.Context, ids []string) (mgclients.ClientsPage, error) {
	return es.svc.RetrieveByIds(ctx, ids)
}

func (es *eventStore) AddConnections(ctx context.Context, conns []things.Connection) (err error) {
	return es.svc.AddConnections(ctx, conns)
}
func (es *eventStore) RemoveConnections(ctx context.Context, conns []things.Connection) (err error) {
	return es.svc.RemoveConnections(ctx, conns)
}
func (es *eventStore) RemoveChannelConnections(ctx context.Context, channelID string) error {
	return es.svc.RemoveChannelConnections(ctx, channelID)
}

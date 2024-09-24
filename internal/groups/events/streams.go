// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	entityRolesEvents "github.com/absmach/magistrala/pkg/entityroles/events"
	"github.com/absmach/magistrala/pkg/events"
	"github.com/absmach/magistrala/pkg/events/store"
	"github.com/absmach/magistrala/pkg/groups"
)

var _ groups.Service = (*eventStore)(nil)

type eventStore struct {
	events.Publisher
	svc groups.Service
	entityRolesEvents.RolesSvcEventStoreMiddleware
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc groups.Service, url, streamID string) (groups.Service, error) {
	publisher, err := store.NewPublisher(ctx, url, streamID)
	if err != nil {
		return nil, err
	}
	rsesm := entityRolesEvents.NewRolesSvcEventStoreMiddleware("groups", svc, publisher)

	return &eventStore{
		svc:                          svc,
		Publisher:                    publisher,
		RolesSvcEventStoreMiddleware: rsesm,
	}, nil
}

func (es eventStore) CreateGroup(ctx context.Context, token, kind string, group groups.Group) (groups.Group, error) {
	group, err := es.svc.CreateGroup(ctx, token, kind, group)
	if err != nil {
		return group, err
	}

	event := createGroupEvent{
		group,
	}

	if err := es.Publish(ctx, event); err != nil {
		return group, err
	}

	return group, nil
}

func (es eventStore) UpdateGroup(ctx context.Context, token string, group groups.Group) (groups.Group, error) {
	group, err := es.svc.UpdateGroup(ctx, token, group)
	if err != nil {
		return group, err
	}

	event := updateGroupEvent{
		group,
	}

	if err := es.Publish(ctx, event); err != nil {
		return group, err
	}

	return group, nil
}

func (es eventStore) ViewGroup(ctx context.Context, token, id string) (groups.Group, error) {
	group, err := es.svc.ViewGroup(ctx, token, id)
	if err != nil {
		return group, err
	}
	event := viewGroupEvent{
		group,
	}

	if err := es.Publish(ctx, event); err != nil {
		return group, err
	}

	return group, nil
}

func (es eventStore) ListGroups(ctx context.Context, token string, pm groups.Page) (groups.Page, error) {
	gp, err := es.svc.ListGroups(ctx, token, pm)
	if err != nil {
		return gp, err
	}
	event := listGroupEvent{
		pm,
	}

	if err := es.Publish(ctx, event); err != nil {
		return gp, err
	}

	return gp, nil
}

func (es eventStore) EnableGroup(ctx context.Context, token, id string) (groups.Group, error) {
	group, err := es.svc.EnableGroup(ctx, token, id)
	if err != nil {
		return group, err
	}

	return es.changeStatus(ctx, group)
}

func (es eventStore) DisableGroup(ctx context.Context, token, id string) (groups.Group, error) {
	group, err := es.svc.DisableGroup(ctx, token, id)
	if err != nil {
		return group, err
	}

	return es.changeStatus(ctx, group)
}

func (es eventStore) changeStatus(ctx context.Context, group groups.Group) (groups.Group, error) {
	event := changeStatusGroupEvent{
		id:        group.ID,
		updatedAt: group.UpdatedAt,
		updatedBy: group.UpdatedBy,
		status:    group.Status.String(),
	}

	if err := es.Publish(ctx, event); err != nil {
		return group, err
	}

	return group, nil
}

func (es eventStore) DeleteGroup(ctx context.Context, token, id string) error {
	if err := es.svc.DeleteGroup(ctx, token, id); err != nil {
		return err
	}
	if err := es.Publish(ctx, deleteGroupEvent{id}); err != nil {
		return err
	}
	return nil
}

func (es eventStore) ListParentGroups(ctx context.Context, token, id string, gm groups.Page) (groups.Page, error) {
	g, err := es.svc.ListParentGroups(ctx, token, id, gm)
	if err != nil {
		return g, err
	}
	if err := es.Publish(ctx, listParentGroupsEvent{id, gm}); err != nil {
		return g, err
	}
	return g, nil
}

func (es eventStore) AddParentGroup(ctx context.Context, token, id, parentID string) error {
	if err := es.svc.AddParentGroup(ctx, token, id, parentID); err != nil {
		return err
	}
	if err := es.Publish(ctx, addParentGroupEvent{id, parentID}); err != nil {
		return err
	}
	return nil
}

func (es eventStore) RemoveParentGroup(ctx context.Context, token, id string) error {
	if err := es.svc.RemoveParentGroup(ctx, token, id); err != nil {
		return err
	}
	if err := es.Publish(ctx, removeParentGroupEvent{id}); err != nil {
		return err
	}
	return nil
}

func (es eventStore) ViewParentGroup(ctx context.Context, token, id string) (groups.Group, error) {
	g, err := es.svc.ViewParentGroup(ctx, token, id)
	if err != nil {
		return g, err
	}
	if err := es.Publish(ctx, viewParentGroupEvent{id}); err != nil {
		return g, err
	}
	return g, nil
}

func (es eventStore) AddChildrenGroups(ctx context.Context, token, id string, childrenGroupIDs []string) error {
	if err := es.svc.AddChildrenGroups(ctx, token, id, childrenGroupIDs); err != nil {
		return err
	}
	if err := es.Publish(ctx, addChildrenGroupsEvent{id, childrenGroupIDs}); err != nil {
		return err
	}
	return nil
}

func (es eventStore) RemoveChildrenGroups(ctx context.Context, token, id string, childrenGroupIDs []string) error {
	if err := es.svc.RemoveChildrenGroups(ctx, token, id, childrenGroupIDs); err != nil {
		return err
	}
	if err := es.Publish(ctx, removeChildrenGroupsEvent{id, childrenGroupIDs}); err != nil {
		return err
	}

	return nil
}
func (es eventStore) RemoveAllChildrenGroups(ctx context.Context, token, id string) error {
	if err := es.svc.RemoveAllChildrenGroups(ctx, token, id); err != nil {
		return err
	}
	if err := es.Publish(ctx, removeAllChildrenGroupsEvent{id}); err != nil {
		return err
	}
	return nil
}

func (es eventStore) ListChildrenGroups(ctx context.Context, token, id string, gm groups.Page) (groups.Page, error) {
	g, err := es.svc.ListChildrenGroups(ctx, token, id, gm)
	if err != nil {
		return g, err
	}
	if err := es.Publish(ctx, listChildrenGroupsEvent{id, gm}); err != nil {
		return g, err
	}
	return g, nil
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/roles"
)

var _ roles.RoleManager = (*RoleManagerEventStore)(nil)

type RoleManagerEventStore struct {
	events.Publisher
	svc             roles.RoleManager
	operationPrefix string
	svcName         string
}

// NewEventStoreMiddleware returns wrapper around auth service that sends
// events to event store.
func NewRoleManagerEventStore(svcName, operationPrefix string, svc roles.RoleManager, publisher events.Publisher) RoleManagerEventStore {
	return RoleManagerEventStore{
		svcName:   svcName,
		svc:       svc,
		Publisher: publisher,
	}
}

func (res *RoleManagerEventStore) AddRole(ctx context.Context, session authn.Session, entityID, roleName string, optionalActions []string, optionalMembers []string) (roles.Role, error) {
	ro, err := res.svc.AddRole(ctx, session, entityID, roleName, optionalActions, optionalMembers)
	if err != nil {
		return ro, err
	}

	e := addRoleEvent{
		operationPrefix: res.operationPrefix,
		Role:            ro,
	}
	if err := res.Publish(ctx, e); err != nil {
		return ro, err
	}
	return ro, nil
}

func (res *RoleManagerEventStore) RemoveRole(ctx context.Context, session authn.Session, entityID, roleName string) error {
	if err := res.svc.RemoveRole(ctx, session, entityID, roleName); err != nil {
		return err
	}
	e := removeRoleEvent{
		operationPrefix: res.operationPrefix,
		roleName:        roleName,
		entityID:        entityID,
	}
	if err := res.Publish(ctx, e); err != nil {
		return err
	}
	return nil
}

func (res *RoleManagerEventStore) UpdateRoleName(ctx context.Context, session authn.Session, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	ro, err := res.svc.UpdateRoleName(ctx, session, entityID, oldRoleName, newRoleName)
	if err != nil {
		return ro, err
	}

	e := updateRoleEvent{
		operationPrefix: res.operationPrefix,
		Role:            ro,
	}
	if err := res.Publish(ctx, e); err != nil {
		return ro, err
	}
	return ro, nil
}

func (res *RoleManagerEventStore) RetrieveRole(ctx context.Context, session authn.Session, entityID, roleName string) (roles.Role, error) {
	ro, err := res.svc.RetrieveRole(ctx, session, entityID, roleName)
	if err != nil {
		return ro, err
	}
	e := retrieveRoleEvent{
		operationPrefix: res.operationPrefix,
		Role:            ro,
	}
	if err := res.Publish(ctx, e); err != nil {
		return ro, err
	}
	return ro, nil
}

func (res *RoleManagerEventStore) RetrieveAllRoles(ctx context.Context, session authn.Session, entityID string, limit, offset uint64) (roles.RolePage, error) {
	rp, err := res.svc.RetrieveAllRoles(ctx, session, entityID, limit, offset)
	if err != nil {
		return rp, err
	}

	e := retrieveAllRolesEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		limit:           limit,
		offset:          offset,
	}
	if err := res.Publish(ctx, e); err != nil {
		return rp, err
	}
	return rp, nil
}

func (res *RoleManagerEventStore) ListAvailableActions(ctx context.Context, session authn.Session) ([]string, error) {
	actions, err := res.svc.ListAvailableActions(ctx, session)
	if err != nil {
		return actions, err
	}
	e := listAvailableActionsEvent{
		operationPrefix: res.operationPrefix,
	}
	if err := res.Publish(ctx, e); err != nil {
		return actions, err
	}
	return actions, nil
}

func (res *RoleManagerEventStore) RoleAddActions(ctx context.Context, session authn.Session, entityID, roleName string, actions []string) ([]string, error) {
	actions, err := res.svc.RoleAddActions(ctx, session, entityID, roleName, actions)
	if err != nil {
		return actions, err
	}
	e := roleAddActionsEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
		actions:         actions,
	}
	if err := res.Publish(ctx, e); err != nil {
		return actions, err
	}
	return actions, nil
}

func (res *RoleManagerEventStore) RoleListActions(ctx context.Context, session authn.Session, entityID, roleName string) ([]string, error) {
	actions, err := res.svc.RoleListActions(ctx, session, entityID, roleName)
	if err != nil {
		return actions, err
	}

	e := roleListActionsEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
	}
	if err := res.Publish(ctx, e); err != nil {
		return actions, err
	}
	return actions, nil
}

func (res *RoleManagerEventStore) RoleCheckActionsExists(ctx context.Context, session authn.Session, entityID, roleName string, actions []string) (bool, error) {
	isAllExists, err := res.svc.RoleCheckActionsExists(ctx, session, entityID, roleName, actions)
	if err != nil {
		return isAllExists, err
	}

	e := roleCheckActionsExistsEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
		actions:         actions,
		isAllExists:     isAllExists,
	}
	if err := res.Publish(ctx, e); err != nil {
		return isAllExists, err
	}
	return isAllExists, nil
}

func (res *RoleManagerEventStore) RoleRemoveActions(ctx context.Context, session authn.Session, entityID, roleName string, actions []string) (err error) {
	if err := res.svc.RoleRemoveActions(ctx, session, entityID, roleName, actions); err != nil {
		return err
	}

	e := roleRemoveActionsEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
		actions:         actions,
	}
	if err := res.Publish(ctx, e); err != nil {
		return err
	}
	return nil

}

func (res *RoleManagerEventStore) RoleRemoveAllActions(ctx context.Context, session authn.Session, entityID, roleName string) error {
	if err := res.svc.RoleRemoveAllActions(ctx, session, entityID, roleName); err != nil {
		return err
	}

	e := roleRemoveAllActionsEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
	}
	if err := res.Publish(ctx, e); err != nil {
		return err
	}
	return nil
}

func (res *RoleManagerEventStore) RoleAddMembers(ctx context.Context, session authn.Session, entityID, roleName string, members []string) ([]string, error) {
	mems, err := res.svc.RoleAddMembers(ctx, session, entityID, roleName, members)
	if err != nil {
		return mems, err
	}

	e := roleAddMembersEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
		members:         members,
	}
	if err := res.Publish(ctx, e); err != nil {
		return mems, err
	}
	return mems, nil
}

func (res *RoleManagerEventStore) RoleListMembers(ctx context.Context, session authn.Session, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	mp, err := res.svc.RoleListMembers(ctx, session, entityID, roleName, limit, offset)
	if err != nil {
		return mp, err
	}

	e := roleListMembersEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
		limit:           limit,
		offset:          offset,
	}
	if err := res.Publish(ctx, e); err != nil {
		return mp, err
	}
	return mp, nil
}

func (res *RoleManagerEventStore) RoleCheckMembersExists(ctx context.Context, session authn.Session, entityID, roleName string, members []string) (bool, error) {
	isAllExists, err := res.svc.RoleCheckMembersExists(ctx, session, entityID, roleName, members)
	if err != nil {
		return isAllExists, err
	}

	e := roleCheckMembersExistsEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
		members:         members,
	}
	if err := res.Publish(ctx, e); err != nil {
		return isAllExists, err
	}
	return isAllExists, nil
}

func (res *RoleManagerEventStore) RoleRemoveMembers(ctx context.Context, session authn.Session, entityID, roleName string, members []string) (err error) {
	if err := res.svc.RoleRemoveMembers(ctx, session, entityID, roleName, members); err != nil {
		return err
	}

	e := roleRemoveMembersEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
		members:         members,
	}
	if err := res.Publish(ctx, e); err != nil {
		return err
	}
	return nil
}

func (res *RoleManagerEventStore) RoleRemoveAllMembers(ctx context.Context, session authn.Session, entityID, roleName string) (err error) {
	if err := res.svc.RoleRemoveAllMembers(ctx, session, entityID, roleName); err != nil {
		return err
	}

	e := roleRemoveAllMembersEvent{
		operationPrefix: res.operationPrefix,
		entityID:        entityID,
		roleName:        roleName,
	}
	if err := res.Publish(ctx, e); err != nil {
		return err
	}
	return nil
}

func (res *RoleManagerEventStore) RemoveMemberFromAllRoles(ctx context.Context, session authn.Session, memberID string) (err error) {
	if err := res.svc.RemoveMemberFromAllRoles(ctx, session, memberID); err != nil {
		return err
	}

	e := removeMemberFromAllRolesEvent{
		operationPrefix: res.operationPrefix,
		memberID:        memberID,
	}
	if err := res.Publish(ctx, e); err != nil {
		return err
	}
	return nil
}

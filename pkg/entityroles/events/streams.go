// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/absmach/magistrala/pkg/events"
	"github.com/absmach/magistrala/pkg/roles"
)

var _ roles.Roles = (*RolesSvcEventStoreMiddleware)(nil)

type RolesSvcEventStoreMiddleware struct {
	events.Publisher
	svc     roles.Roles
	svcName string
}

// NewEventStoreMiddleware returns wrapper around auth service that sends
// events to event store.
func NewRolesSvcEventStoreMiddleware(svcName string, svc roles.Roles, publisher events.Publisher) RolesSvcEventStoreMiddleware {
	return RolesSvcEventStoreMiddleware{
		svcName:   svcName,
		svc:       svc,
		Publisher: publisher,
	}
}

func (res *RolesSvcEventStoreMiddleware) AddRole(ctx context.Context, token, entityID, roleName string, optionalActions []string, optionalMembers []string) (roles.Role, error) {
	return res.svc.AddRole(ctx, token, entityID, roleName, optionalActions, optionalMembers)
}
func (res *RolesSvcEventStoreMiddleware) RemoveRole(ctx context.Context, token, entityID, roleName string) error {
	return res.svc.RemoveRole(ctx, token, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) UpdateRoleName(ctx context.Context, token, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return res.svc.UpdateRoleName(ctx, token, entityID, oldRoleName, newRoleName)
}
func (res *RolesSvcEventStoreMiddleware) RetrieveRole(ctx context.Context, token, entityID, roleName string) (roles.Role, error) {
	return res.svc.RetrieveRole(ctx, token, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RetrieveAllRoles(ctx context.Context, token, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return res.svc.RetrieveAllRoles(ctx, token, entityID, limit, offset)
}
func (res *RolesSvcEventStoreMiddleware) RoleAddActions(ctx context.Context, token, entityID, roleName string, actions []string) (ops []string, err error) {
	return res.svc.RoleAddActions(ctx, token, entityID, roleName, actions)
}
func (res *RolesSvcEventStoreMiddleware) RoleListActions(ctx context.Context, token, entityID, roleName string) ([]string, error) {
	return res.svc.RoleListActions(ctx, token, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RoleCheckActionsExists(ctx context.Context, token, entityID, roleName string, actions []string) (bool, error) {
	return res.svc.RoleCheckActionsExists(ctx, token, entityID, roleName, actions)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveActions(ctx context.Context, token, entityID, roleName string, actions []string) (err error) {
	return res.svc.RoleRemoveActions(ctx, token, entityID, roleName, actions)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveAllActions(ctx context.Context, token, entityID, roleName string) error {
	return res.svc.RoleRemoveAllActions(ctx, token, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RoleAddMembers(ctx context.Context, token, entityID, roleName string, members []string) ([]string, error) {
	return res.svc.RoleAddMembers(ctx, token, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RoleListMembers(ctx context.Context, token, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return res.svc.RoleListMembers(ctx, token, entityID, roleName, limit, offset)
}
func (res *RolesSvcEventStoreMiddleware) RoleCheckMembersExists(ctx context.Context, token, entityID, roleName string, members []string) (bool, error) {
	return res.svc.RoleCheckMembersExists(ctx, token, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveMembers(ctx context.Context, token, entityID, roleName string, members []string) (err error) {
	return res.svc.RoleRemoveMembers(ctx, token, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveAllMembers(ctx context.Context, token, entityID, roleName string) (err error) {
	return res.svc.RoleRemoveAllMembers(ctx, token, entityID, roleName)
}

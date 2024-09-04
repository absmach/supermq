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

func (res *RolesSvcEventStoreMiddleware) AddNewEntityRoles(ctx context.Context, entityID string, newEntityDefaultRoles map[string][]string, optionalPolicies []roles.OptionalPolicy) ([]roles.Role, error) {
	return res.svc.AddNewEntityRoles(ctx, entityID, newEntityDefaultRoles, optionalPolicies)
}
func (res *RolesSvcEventStoreMiddleware) AddRole(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	return res.svc.AddRole(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (res *RolesSvcEventStoreMiddleware) RemoveRole(ctx context.Context, entityID, roleName string) error {
	return res.svc.RemoveRole(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) UpdateRoleName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return res.svc.UpdateRoleName(ctx, entityID, oldRoleName, newRoleName)
}
func (res *RolesSvcEventStoreMiddleware) RetrieveRole(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return res.svc.RetrieveRole(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RetrieveAllRoles(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return res.svc.RetrieveAllRoles(ctx, entityID, limit, offset)
}
func (res *RolesSvcEventStoreMiddleware) RoleAddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return res.svc.RoleAddOperation(ctx, entityID, roleName, operations)
}
func (res *RolesSvcEventStoreMiddleware) RoleListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return res.svc.RoleListOperations(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RoleCheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return res.svc.RoleCheckOperationsExists(ctx, entityID, roleName, operations)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return res.svc.RoleRemoveOperations(ctx, entityID, roleName, operations)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return res.svc.RoleRemoveAllOperations(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RoleAddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return res.svc.RoleAddMembers(ctx, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RoleListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return res.svc.RoleListMembers(ctx, entityID, roleName, limit, offset)
}
func (res *RolesSvcEventStoreMiddleware) RoleCheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return res.svc.RoleCheckMembersExists(ctx, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return res.svc.RoleRemoveMembers(ctx, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return res.svc.RoleRemoveAllMembers(ctx, entityID, roleName)
}

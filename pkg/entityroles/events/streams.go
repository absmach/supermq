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
func (res *RolesSvcEventStoreMiddleware) Add(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	return res.svc.Add(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (res *RolesSvcEventStoreMiddleware) Remove(ctx context.Context, entityID, roleName string) error {
	return res.svc.Remove(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) UpdateName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return res.svc.UpdateName(ctx, entityID, oldRoleName, newRoleName)
}
func (res *RolesSvcEventStoreMiddleware) Retrieve(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return res.svc.Retrieve(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RetrieveAll(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return res.svc.RetrieveAll(ctx, entityID, limit, offset)
}
func (res *RolesSvcEventStoreMiddleware) AddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return res.svc.AddOperation(ctx, entityID, roleName, operations)
}
func (res *RolesSvcEventStoreMiddleware) ListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return res.svc.ListOperations(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) CheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return res.svc.CheckOperationsExists(ctx, entityID, roleName, operations)
}
func (res *RolesSvcEventStoreMiddleware) RemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return res.svc.RemoveOperations(ctx, entityID, roleName, operations)
}
func (res *RolesSvcEventStoreMiddleware) RemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return res.svc.RemoveAllOperations(ctx, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) AddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return res.svc.AddMembers(ctx, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) ListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return res.svc.ListMembers(ctx, entityID, roleName, limit, offset)
}
func (res *RolesSvcEventStoreMiddleware) CheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return res.svc.CheckMembersExists(ctx, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return res.svc.RemoveMembers(ctx, entityID, roleName, members)
}
func (res *RolesSvcEventStoreMiddleware) RemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return res.svc.RemoveAllMembers(ctx, entityID, roleName)
}

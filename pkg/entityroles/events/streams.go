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

func (res *RolesSvcEventStoreMiddleware) AddRole(ctx context.Context, token, entityID, roleName string, optionalCapabilities []string, optionalMembers []string) (roles.Role, error) {
	return res.svc.AddRole(ctx, token, entityID, roleName, optionalCapabilities, optionalMembers)
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
func (res *RolesSvcEventStoreMiddleware) RoleAddCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (ops []string, err error) {
	return res.svc.RoleAddCapabilities(ctx, token, entityID, roleName, capabilities)
}
func (res *RolesSvcEventStoreMiddleware) RoleListCapabilities(ctx context.Context, token, entityID, roleName string) ([]string, error) {
	return res.svc.RoleListCapabilities(ctx, token, entityID, roleName)
}
func (res *RolesSvcEventStoreMiddleware) RoleCheckCapabilitiesExists(ctx context.Context, token, entityID, roleName string, capabilities []string) (bool, error) {
	return res.svc.RoleCheckCapabilitiesExists(ctx, token, entityID, roleName, capabilities)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (err error) {
	return res.svc.RoleRemoveCapabilities(ctx, token, entityID, roleName, capabilities)
}
func (res *RolesSvcEventStoreMiddleware) RoleRemoveAllCapabilities(ctx context.Context, token, entityID, roleName string) error {
	return res.svc.RoleRemoveAllCapabilities(ctx, token, entityID, roleName)
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

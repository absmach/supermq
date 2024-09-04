// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"
	"log/slog"

	"github.com/absmach/magistrala/pkg/roles"
)

var _ roles.Roles = (*RolesSvcLoggingMiddleware)(nil)

type RolesSvcLoggingMiddleware struct {
	svcName string
	svc     roles.Roles
	logger  *slog.Logger
}

func NewRolesSvcLoggingMiddleware(svcName string, svc roles.Roles, logger *slog.Logger) RolesSvcLoggingMiddleware {
	return RolesSvcLoggingMiddleware{
		svcName: svcName,
		svc:     svc,
		logger:  logger,
	}
}

func (lm *RolesSvcLoggingMiddleware) AddNewEntityRoles(ctx context.Context, entityID string, newEntityDefaultRoles map[string][]string, optionalPolicies []roles.OptionalPolicy) ([]roles.Role, error) {
	return lm.svc.AddNewEntityRoles(ctx, entityID, newEntityDefaultRoles, optionalPolicies)
}
func (lm *RolesSvcLoggingMiddleware) Add(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	return lm.svc.Add(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (lm *RolesSvcLoggingMiddleware) Remove(ctx context.Context, entityID, roleName string) error {
	return lm.svc.Remove(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) UpdateName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return lm.svc.UpdateName(ctx, entityID, oldRoleName, newRoleName)
}
func (lm *RolesSvcLoggingMiddleware) Retrieve(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return lm.svc.Retrieve(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) RetrieveAll(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return lm.svc.RetrieveAll(ctx, entityID, limit, offset)
}
func (lm *RolesSvcLoggingMiddleware) AddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return lm.svc.AddOperation(ctx, entityID, roleName, operations)
}
func (lm *RolesSvcLoggingMiddleware) ListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return lm.svc.ListOperations(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) CheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return lm.svc.CheckOperationsExists(ctx, entityID, roleName, operations)
}
func (lm *RolesSvcLoggingMiddleware) RemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return lm.svc.RemoveOperations(ctx, entityID, roleName, operations)
}
func (lm *RolesSvcLoggingMiddleware) RemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return lm.svc.RemoveAllOperations(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) AddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return lm.svc.AddMembers(ctx, entityID, roleName, members)
}
func (lm *RolesSvcLoggingMiddleware) ListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return lm.svc.ListMembers(ctx, entityID, roleName, limit, offset)
}
func (lm *RolesSvcLoggingMiddleware) CheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return lm.svc.CheckMembersExists(ctx, entityID, roleName, members)
}
func (lm *RolesSvcLoggingMiddleware) RemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return lm.svc.RemoveMembers(ctx, entityID, roleName, members)
}
func (lm *RolesSvcLoggingMiddleware) RemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return lm.svc.RemoveAllMembers(ctx, entityID, roleName)
}

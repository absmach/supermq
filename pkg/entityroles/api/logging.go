// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
func (lm *RolesSvcLoggingMiddleware) AddRole(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (ro roles.Role, err error) {
	prefix := fmt.Sprintf("Add %s roles", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_roles_retrieve_all",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
				slog.Any("optional_operations", optionalOperations),
				slog.Any("optional_members", optionalMembers),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())

	return lm.svc.AddRole(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (lm *RolesSvcLoggingMiddleware) RemoveRole(ctx context.Context, entityID, roleName string) error {
	return lm.svc.RemoveRole(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) UpdateRoleName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return lm.svc.UpdateRoleName(ctx, entityID, oldRoleName, newRoleName)
}
func (lm *RolesSvcLoggingMiddleware) RetrieveRole(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return lm.svc.RetrieveRole(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) RetrieveAllRoles(ctx context.Context, entityID string, limit, offset uint64) (rp roles.RolePage, err error) {
	prefix := fmt.Sprintf("List %s roles", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_roles_retrieve_all",
				slog.String("entity_id", entityID),
				slog.Uint64("limit", limit),
				slog.Uint64("offset", offset),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())

	return lm.svc.RetrieveAllRoles(ctx, entityID, limit, offset)
}
func (lm *RolesSvcLoggingMiddleware) RoleAddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return lm.svc.RoleAddOperation(ctx, entityID, roleName, operations)
}
func (lm *RolesSvcLoggingMiddleware) RoleListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return lm.svc.RoleListOperations(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) RoleCheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return lm.svc.RoleCheckOperationsExists(ctx, entityID, roleName, operations)
}
func (lm *RolesSvcLoggingMiddleware) RoleRemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return lm.svc.RoleRemoveOperations(ctx, entityID, roleName, operations)
}
func (lm *RolesSvcLoggingMiddleware) RoleRemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return lm.svc.RoleRemoveAllOperations(ctx, entityID, roleName)
}
func (lm *RolesSvcLoggingMiddleware) RoleAddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return lm.svc.RoleAddMembers(ctx, entityID, roleName, members)
}
func (lm *RolesSvcLoggingMiddleware) RoleListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return lm.svc.RoleListMembers(ctx, entityID, roleName, limit, offset)
}
func (lm *RolesSvcLoggingMiddleware) RoleCheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return lm.svc.RoleCheckMembersExists(ctx, entityID, roleName, members)
}
func (lm *RolesSvcLoggingMiddleware) RoleRemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return lm.svc.RoleRemoveMembers(ctx, entityID, roleName, members)
}
func (lm *RolesSvcLoggingMiddleware) RoleRemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return lm.svc.RoleRemoveAllMembers(ctx, entityID, roleName)
}

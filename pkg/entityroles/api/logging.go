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

func (lm *RolesSvcLoggingMiddleware) AddRole(ctx context.Context, token, entityID, roleName string, optionalCapabilities []string, optionalMembers []string) (ro roles.Role, err error) {
	prefix := fmt.Sprintf("Add %s roles", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_add_role",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
				slog.Any("optional_capabilities", optionalCapabilities),
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
	return lm.svc.AddRole(ctx, token, entityID, roleName, optionalCapabilities, optionalMembers)
}

func (lm *RolesSvcLoggingMiddleware) RemoveRole(ctx context.Context, token, entityID, roleName string) (err error) {
	prefix := fmt.Sprintf("Delete %s role", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_delete_role",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RemoveRole(ctx, token, entityID, roleName)
}

func (lm *RolesSvcLoggingMiddleware) UpdateRoleName(ctx context.Context, token, entityID, oldRoleName, newRoleName string) (ro roles.Role, err error) {
	prefix := fmt.Sprintf("Update %s role name", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_update_role_name",
				slog.String("entity_id", entityID),
				slog.String("old_role_name", oldRoleName),
				slog.String("new_role_name", newRoleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.UpdateRoleName(ctx, token, entityID, oldRoleName, newRoleName)
}

func (lm *RolesSvcLoggingMiddleware) RetrieveRole(ctx context.Context, token, entityID, roleName string) (ro roles.Role, err error) {
	prefix := fmt.Sprintf("Retrieve %s role", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_update_role_name",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RetrieveRole(ctx, token, entityID, roleName)
}

func (lm *RolesSvcLoggingMiddleware) RetrieveAllRoles(ctx context.Context, token, entityID string, limit, offset uint64) (rp roles.RolePage, err error) {
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
	return lm.svc.RetrieveAllRoles(ctx, token, entityID, limit, offset)
}

func (lm *RolesSvcLoggingMiddleware) RoleAddCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (caps []string, err error) {
	prefix := fmt.Sprintf("%s role add capabilities", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_role_add_capabilities",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
				slog.Any("capabilities", capabilities),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RoleAddCapabilities(ctx, token, entityID, roleName, capabilities)
}

func (lm *RolesSvcLoggingMiddleware) RoleListCapabilities(ctx context.Context, token, entityID, roleName string) (roOps []string, err error) {
	prefix := fmt.Sprintf("%s role list capabilities", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_list_role_capabilities",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RoleListCapabilities(ctx, token, entityID, roleName)
}

func (lm *RolesSvcLoggingMiddleware) RoleCheckCapabilitiesExists(ctx context.Context, token, entityID, roleName string, capabilities []string) (bool, error) {
	return lm.svc.RoleCheckCapabilitiesExists(ctx, token, entityID, roleName, capabilities)
}

func (lm *RolesSvcLoggingMiddleware) RoleRemoveCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (err error) {
	prefix := fmt.Sprintf("%s role remove capabilities", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_role_remove_capabilities",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
				slog.Any("capabilities", capabilities),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RoleRemoveCapabilities(ctx, token, entityID, roleName, capabilities)
}

func (lm *RolesSvcLoggingMiddleware) RoleRemoveAllCapabilities(ctx context.Context, token, entityID, roleName string) (err error) {
	prefix := fmt.Sprintf("%s role remove all capabilities", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_role_remove_all_capabilities",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RoleRemoveAllCapabilities(ctx, token, entityID, roleName)
}

func (lm *RolesSvcLoggingMiddleware) RoleAddMembers(ctx context.Context, token, entityID, roleName string, members []string) (mems []string, err error) {
	prefix := fmt.Sprintf("%s role add members", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_role_add_members",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
				slog.Any("members", members),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RoleAddMembers(ctx, token, entityID, roleName, members)
}

func (lm *RolesSvcLoggingMiddleware) RoleListMembers(ctx context.Context, token, entityID, roleName string, limit, offset uint64) (mp roles.MembersPage, err error) {
	prefix := fmt.Sprintf("%s role list members", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_role_add_members",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
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
	return lm.svc.RoleListMembers(ctx, token, entityID, roleName, limit, offset)
}

func (lm *RolesSvcLoggingMiddleware) RoleCheckMembersExists(ctx context.Context, token, entityID, roleName string, members []string) (bool, error) {
	return lm.svc.RoleCheckMembersExists(ctx, token, entityID, roleName, members)
}

func (lm *RolesSvcLoggingMiddleware) RoleRemoveMembers(ctx context.Context, token, entityID, roleName string, members []string) (err error) {
	prefix := fmt.Sprintf("%s role remove members", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_role_remove_members",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
				slog.Any("members", members),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RoleRemoveMembers(ctx, token, entityID, roleName, members)
}

func (lm *RolesSvcLoggingMiddleware) RoleRemoveAllMembers(ctx context.Context, token, entityID, roleName string) (err error) {
	prefix := fmt.Sprintf("%s role remove all members", lm.svcName)
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group(lm.svcName+"_role_remove_all_members",
				slog.String("entity_id", entityID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(prefix+" failed", args...)
			return
		}
		lm.logger.Info(prefix+" completed successfully", args...)
	}(time.Now())
	return lm.svc.RoleRemoveAllMembers(ctx, token, entityID, roleName)
}

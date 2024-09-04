// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"

	"github.com/absmach/magistrala/pkg/roles"
	"github.com/go-kit/kit/metrics"
)

var _ roles.Roles = (*RolesSvcMetricsMiddleware)(nil)

type RolesSvcMetricsMiddleware struct {
	svcName string
	svc     roles.Roles
	counter metrics.Counter
	latency metrics.Histogram
}

func NewRolesSvcMetricsMiddleware(svcName string, svc roles.Roles, counter metrics.Counter, latency metrics.Histogram) RolesSvcMetricsMiddleware {
	return RolesSvcMetricsMiddleware{
		svcName: svcName,
		svc:     svc,
		counter: counter,
		latency: latency,
	}
}

func (rmm *RolesSvcMetricsMiddleware) AddNewEntityRoles(ctx context.Context, entityID string, newEntityDefaultRoles map[string][]string, optionalPolicies []roles.OptionalPolicy) ([]roles.Role, error) {
	return rmm.svc.AddNewEntityRoles(ctx, entityID, newEntityDefaultRoles, optionalPolicies)
}
func (rmm *RolesSvcMetricsMiddleware) AddRole(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	return rmm.svc.AddRole(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (rmm *RolesSvcMetricsMiddleware) RemoveRole(ctx context.Context, entityID, roleName string) error {
	return rmm.svc.RemoveRole(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) UpdateRoleName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return rmm.svc.UpdateRoleName(ctx, entityID, oldRoleName, newRoleName)
}
func (rmm *RolesSvcMetricsMiddleware) RetrieveRole(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return rmm.svc.RetrieveRole(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) RetrieveAllRoles(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return rmm.svc.RetrieveAllRoles(ctx, entityID, limit, offset)
}
func (rmm *RolesSvcMetricsMiddleware) RoleAddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return rmm.svc.RoleAddOperation(ctx, entityID, roleName, operations)
}
func (rmm *RolesSvcMetricsMiddleware) RoleListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return rmm.svc.RoleListOperations(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) RoleCheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return rmm.svc.RoleCheckOperationsExists(ctx, entityID, roleName, operations)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return rmm.svc.RoleRemoveOperations(ctx, entityID, roleName, operations)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return rmm.svc.RoleRemoveAllOperations(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) RoleAddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return rmm.svc.RoleAddMembers(ctx, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RoleListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return rmm.svc.RoleListMembers(ctx, entityID, roleName, limit, offset)
}
func (rmm *RolesSvcMetricsMiddleware) RoleCheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return rmm.svc.RoleCheckMembersExists(ctx, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return rmm.svc.RoleRemoveMembers(ctx, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return rmm.svc.RoleRemoveAllMembers(ctx, entityID, roleName)
}

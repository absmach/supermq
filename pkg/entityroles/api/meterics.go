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
func (rmm *RolesSvcMetricsMiddleware) Add(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	return rmm.svc.Add(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (rmm *RolesSvcMetricsMiddleware) Remove(ctx context.Context, entityID, roleName string) error {
	return rmm.svc.Remove(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) UpdateName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return rmm.svc.UpdateName(ctx, entityID, oldRoleName, newRoleName)
}
func (rmm *RolesSvcMetricsMiddleware) Retrieve(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return rmm.svc.Retrieve(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) RetrieveAll(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return rmm.svc.RetrieveAll(ctx, entityID, limit, offset)
}
func (rmm *RolesSvcMetricsMiddleware) AddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return rmm.svc.AddOperation(ctx, entityID, roleName, operations)
}
func (rmm *RolesSvcMetricsMiddleware) ListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return rmm.svc.ListOperations(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) CheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return rmm.svc.CheckOperationsExists(ctx, entityID, roleName, operations)
}
func (rmm *RolesSvcMetricsMiddleware) RemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return rmm.svc.RemoveOperations(ctx, entityID, roleName, operations)
}
func (rmm *RolesSvcMetricsMiddleware) RemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return rmm.svc.RemoveAllOperations(ctx, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) AddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return rmm.svc.AddMembers(ctx, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) ListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return rmm.svc.ListMembers(ctx, entityID, roleName, limit, offset)
}
func (rmm *RolesSvcMetricsMiddleware) CheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return rmm.svc.CheckMembersExists(ctx, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return rmm.svc.RemoveMembers(ctx, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return rmm.svc.RemoveAllMembers(ctx, entityID, roleName)
}

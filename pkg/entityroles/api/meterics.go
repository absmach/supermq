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

func (rmm *RolesSvcMetricsMiddleware) AddRole(ctx context.Context, token, entityID, roleName string, optionalCapabilities []string, optionalMembers []string) (roles.Role, error) {
	return rmm.svc.AddRole(ctx, token, entityID, roleName, optionalCapabilities, optionalMembers)
}
func (rmm *RolesSvcMetricsMiddleware) RemoveRole(ctx context.Context, token, entityID, roleName string) error {
	return rmm.svc.RemoveRole(ctx, token, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) UpdateRoleName(ctx context.Context, token, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return rmm.svc.UpdateRoleName(ctx, token, entityID, oldRoleName, newRoleName)
}
func (rmm *RolesSvcMetricsMiddleware) RetrieveRole(ctx context.Context, token, entityID, roleName string) (roles.Role, error) {
	return rmm.svc.RetrieveRole(ctx, token, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) RetrieveAllRoles(ctx context.Context, token, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return rmm.svc.RetrieveAllRoles(ctx, token, entityID, limit, offset)
}
func (rmm *RolesSvcMetricsMiddleware) RoleAddCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (caps []string, err error) {
	return rmm.svc.RoleAddCapabilities(ctx, token, entityID, roleName, capabilities)
}
func (rmm *RolesSvcMetricsMiddleware) RoleListCapabilities(ctx context.Context, token, entityID, roleName string) ([]string, error) {
	return rmm.svc.RoleListCapabilities(ctx, token, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) RoleCheckCapabilitiesExists(ctx context.Context, token, entityID, roleName string, capabilities []string) (bool, error) {
	return rmm.svc.RoleCheckCapabilitiesExists(ctx, token, entityID, roleName, capabilities)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (err error) {
	return rmm.svc.RoleRemoveCapabilities(ctx, token, entityID, roleName, capabilities)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveAllCapabilities(ctx context.Context, token, entityID, roleName string) error {
	return rmm.svc.RoleRemoveAllCapabilities(ctx, token, entityID, roleName)
}
func (rmm *RolesSvcMetricsMiddleware) RoleAddMembers(ctx context.Context, token, entityID, roleName string, members []string) ([]string, error) {
	return rmm.svc.RoleAddMembers(ctx, token, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RoleListMembers(ctx context.Context, token, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return rmm.svc.RoleListMembers(ctx, token, entityID, roleName, limit, offset)
}
func (rmm *RolesSvcMetricsMiddleware) RoleCheckMembersExists(ctx context.Context, token, entityID, roleName string, members []string) (bool, error) {
	return rmm.svc.RoleCheckMembersExists(ctx, token, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveMembers(ctx context.Context, token, entityID, roleName string, members []string) (err error) {
	return rmm.svc.RoleRemoveMembers(ctx, token, entityID, roleName, members)
}
func (rmm *RolesSvcMetricsMiddleware) RoleRemoveAllMembers(ctx context.Context, token, entityID, roleName string) (err error) {
	return rmm.svc.RoleRemoveAllMembers(ctx, token, entityID, roleName)
}

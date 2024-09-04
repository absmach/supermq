package tracing

import (
	"context"

	"github.com/absmach/magistrala/pkg/roles"
	"go.opentelemetry.io/otel/trace"
)

var _ roles.Roles = (*RolesSvcTracingMiddleware)(nil)

type RolesSvcTracingMiddleware struct {
	svcName string
	roles   roles.Roles
	tracer  trace.Tracer
}

func NewRolesSvcTracingMiddleware(svcName string, svc roles.Roles, tracer trace.Tracer) RolesSvcTracingMiddleware {
	return RolesSvcTracingMiddleware{svcName, svc, tracer}
}

func (rtm *RolesSvcTracingMiddleware) AddNewEntityRoles(ctx context.Context, entityID string, newEntityDefaultRoles map[string][]string, optionalPolicies []roles.OptionalPolicy) ([]roles.Role, error) {
	return rtm.roles.AddNewEntityRoles(ctx, entityID, newEntityDefaultRoles, optionalPolicies)
}
func (rtm *RolesSvcTracingMiddleware) Add(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	return rtm.roles.Add(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (rtm *RolesSvcTracingMiddleware) Remove(ctx context.Context, entityID, roleName string) error {
	return rtm.roles.Remove(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) UpdateName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return rtm.roles.UpdateName(ctx, entityID, oldRoleName, newRoleName)
}
func (rtm *RolesSvcTracingMiddleware) Retrieve(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return rtm.roles.Retrieve(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) RetrieveAll(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return rtm.roles.RetrieveAll(ctx, entityID, limit, offset)
}
func (rtm *RolesSvcTracingMiddleware) AddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return rtm.roles.AddOperation(ctx, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) ListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return rtm.roles.ListOperations(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) CheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return rtm.roles.CheckOperationsExists(ctx, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return rtm.roles.RemoveOperations(ctx, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return rtm.roles.RemoveAllOperations(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) AddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return rtm.roles.AddMembers(ctx, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) ListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return rtm.roles.ListMembers(ctx, entityID, roleName, limit, offset)
}
func (rtm *RolesSvcTracingMiddleware) CheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return rtm.roles.CheckMembersExists(ctx, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return rtm.roles.RemoveMembers(ctx, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return rtm.roles.RemoveAllMembers(ctx, entityID, roleName)
}

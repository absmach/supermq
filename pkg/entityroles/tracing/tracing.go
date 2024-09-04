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
func (rtm *RolesSvcTracingMiddleware) AddRole(ctx context.Context, entityID, roleName string, optionalOperations []roles.Operation, optionalMembers []string) (roles.Role, error) {
	return rtm.roles.AddRole(ctx, entityID, roleName, optionalOperations, optionalMembers)
}
func (rtm *RolesSvcTracingMiddleware) RemoveRole(ctx context.Context, entityID, roleName string) error {
	return rtm.roles.RemoveRole(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) UpdateRoleName(ctx context.Context, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return rtm.roles.UpdateRoleName(ctx, entityID, oldRoleName, newRoleName)
}
func (rtm *RolesSvcTracingMiddleware) RetrieveRole(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	return rtm.roles.RetrieveRole(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) RetrieveAllRoles(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return rtm.roles.RetrieveAllRoles(ctx, entityID, limit, offset)
}
func (rtm *RolesSvcTracingMiddleware) RoleAddOperation(ctx context.Context, entityID, roleName string, operations []roles.Operation) (ops []roles.Operation, err error) {
	return rtm.roles.RoleAddOperation(ctx, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RoleListOperations(ctx context.Context, entityID, roleName string) ([]roles.Operation, error) {
	return rtm.roles.RoleListOperations(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) RoleCheckOperationsExists(ctx context.Context, entityID, roleName string, operations []roles.Operation) (bool, error) {
	return rtm.roles.RoleCheckOperationsExists(ctx, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveOperations(ctx context.Context, entityID, roleName string, operations []roles.Operation) (err error) {
	return rtm.roles.RoleRemoveOperations(ctx, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveAllOperations(ctx context.Context, entityID, roleName string) error {
	return rtm.roles.RoleRemoveAllOperations(ctx, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) RoleAddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error) {
	return rtm.roles.RoleAddMembers(ctx, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RoleListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return rtm.roles.RoleListMembers(ctx, entityID, roleName, limit, offset)
}
func (rtm *RolesSvcTracingMiddleware) RoleCheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error) {
	return rtm.roles.RoleCheckMembersExists(ctx, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error) {
	return rtm.roles.RoleRemoveMembers(ctx, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveAllMembers(ctx context.Context, entityID, roleName string) (err error) {
	return rtm.roles.RoleRemoveAllMembers(ctx, entityID, roleName)
}

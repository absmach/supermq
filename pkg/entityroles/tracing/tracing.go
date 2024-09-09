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

func (rtm *RolesSvcTracingMiddleware) AddRole(ctx context.Context, token, entityID, roleName string, optionalOperations []string, optionalMembers []string) (roles.Role, error) {
	return rtm.roles.AddRole(ctx, token, entityID, roleName, optionalOperations, optionalMembers)
}
func (rtm *RolesSvcTracingMiddleware) RemoveRole(ctx context.Context, token, entityID, roleName string) error {
	return rtm.roles.RemoveRole(ctx, token, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) UpdateRoleName(ctx context.Context, token, entityID, oldRoleName, newRoleName string) (roles.Role, error) {
	return rtm.roles.UpdateRoleName(ctx, token, entityID, oldRoleName, newRoleName)
}
func (rtm *RolesSvcTracingMiddleware) RetrieveRole(ctx context.Context, token, entityID, roleName string) (roles.Role, error) {
	return rtm.roles.RetrieveRole(ctx, token, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) RetrieveAllRoles(ctx context.Context, token, entityID string, limit, offset uint64) (roles.RolePage, error) {
	return rtm.roles.RetrieveAllRoles(ctx, token, entityID, limit, offset)
}
func (rtm *RolesSvcTracingMiddleware) RoleAddOperations(ctx context.Context, token, entityID, roleName string, operations []string) (ops []string, err error) {
	return rtm.roles.RoleAddOperations(ctx, token, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RoleListOperations(ctx context.Context, token, entityID, roleName string) ([]string, error) {
	return rtm.roles.RoleListOperations(ctx, token, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) RoleCheckOperationsExists(ctx context.Context, token, entityID, roleName string, operations []string) (bool, error) {
	return rtm.roles.RoleCheckOperationsExists(ctx, token, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveOperations(ctx context.Context, token, entityID, roleName string, operations []string) (err error) {
	return rtm.roles.RoleRemoveOperations(ctx, token, entityID, roleName, operations)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveAllOperations(ctx context.Context, token, entityID, roleName string) error {
	return rtm.roles.RoleRemoveAllOperations(ctx, token, entityID, roleName)
}
func (rtm *RolesSvcTracingMiddleware) RoleAddMembers(ctx context.Context, token, entityID, roleName string, members []string) ([]string, error) {
	return rtm.roles.RoleAddMembers(ctx, token, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RoleListMembers(ctx context.Context, token, entityID, roleName string, limit, offset uint64) (roles.MembersPage, error) {
	return rtm.roles.RoleListMembers(ctx, token, entityID, roleName, limit, offset)
}
func (rtm *RolesSvcTracingMiddleware) RoleCheckMembersExists(ctx context.Context, token, entityID, roleName string, members []string) (bool, error) {
	return rtm.roles.RoleCheckMembersExists(ctx, token, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveMembers(ctx context.Context, token, entityID, roleName string, members []string) (err error) {
	return rtm.roles.RoleRemoveMembers(ctx, token, entityID, roleName, members)
}
func (rtm *RolesSvcTracingMiddleware) RoleRemoveAllMembers(ctx context.Context, token, entityID, roleName string) (err error) {
	return rtm.roles.RoleRemoveAllMembers(ctx, token, entityID, roleName)
}

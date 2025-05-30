// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"maps"
	"time"

	"github.com/absmach/supermq/pkg/authn"
	smqauthz "github.com/absmach/supermq/pkg/authz"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	"github.com/absmach/supermq/pkg/svcutil"
)

var _ roles.RoleManager = (*RoleManagerAuthorizationMiddleware)(nil)

type RoleManagerAuthorizationMiddleware struct {
	entityType string
	svc        roles.RoleManager
	authz      smqauthz.Authorization
	callout    callout.Callout
	opp        svcutil.OperationPerm
}

// AuthorizationMiddleware adds authorization to the clients service.
func NewRoleManagerAuthorizationMiddleware(entityType string, svc roles.RoleManager, authz smqauthz.Authorization, opPerm map[svcutil.Operation]svcutil.Permission, callout callout.Callout) (RoleManagerAuthorizationMiddleware, error) {
	opp := roles.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(opPerm); err != nil {
		return RoleManagerAuthorizationMiddleware{}, err
	}
	if err := opp.Validate(); err != nil {
		return RoleManagerAuthorizationMiddleware{}, err
	}

	ram := RoleManagerAuthorizationMiddleware{
		entityType: entityType,
		svc:        svc,
		authz:      authz,
		opp:        opp,
		callout:    callout,
	}
	if err := ram.validate(); err != nil {
		return RoleManagerAuthorizationMiddleware{}, err
	}
	return ram, nil
}

func (ram RoleManagerAuthorizationMiddleware) validate() error {
	if err := ram.opp.Validate(); err != nil {
		return err
	}
	return nil
}

func (ram RoleManagerAuthorizationMiddleware) AddRole(ctx context.Context, session authn.Session, entityID, roleName string, optionalActions []string, optionalMembers []string) (roles.RoleProvision, error) {
	if err := ram.authorize(ctx, roles.OpAddRole, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return roles.RoleProvision{}, err
	}
	if err := ram.validateMembers(ctx, session, optionalMembers); err != nil {
		return roles.RoleProvision{}, err
	}
	params := map[string]any{
		"entity_id":        entityID,
		"role_name":        roleName,
		"optional_actions": optionalActions,
		"optional_members": optionalMembers,
		"count":            1,
	}
	if err := ram.callOut(ctx, session, roles.OpAddRole.String(roles.OperationNames), params); err != nil {
		return roles.RoleProvision{}, err
	}
	return ram.svc.AddRole(ctx, session, entityID, roleName, optionalActions, optionalMembers)
}

func (ram RoleManagerAuthorizationMiddleware) RemoveRole(ctx context.Context, session authn.Session, entityID, roleID string) error {
	if err := ram.authorize(ctx, roles.OpRemoveRole, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
	}
	if err := ram.callOut(ctx, session, roles.OpRemoveRole.String(roles.OperationNames), params); err != nil {
		return err
	}
	return ram.svc.RemoveRole(ctx, session, entityID, roleID)
}

func (ram RoleManagerAuthorizationMiddleware) UpdateRoleName(ctx context.Context, session authn.Session, entityID, roleID, newRoleName string) (roles.Role, error) {
	if err := ram.authorize(ctx, roles.OpUpdateRoleName, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return roles.Role{}, err
	}
	params := map[string]any{
		"entity_id":     entityID,
		"role_id":       roleID,
		"new_role_name": newRoleName,
	}
	if err := ram.callOut(ctx, session, roles.OpUpdateRoleName.String(roles.OperationNames), params); err != nil {
		return roles.Role{}, err
	}
	return ram.svc.UpdateRoleName(ctx, session, entityID, roleID, newRoleName)
}

func (ram RoleManagerAuthorizationMiddleware) RetrieveRole(ctx context.Context, session authn.Session, entityID, roleID string) (roles.Role, error) {
	if err := ram.authorize(ctx, roles.OpRetrieveRole, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return roles.Role{}, err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
	}
	if err := ram.callOut(ctx, session, roles.OpRetrieveRole.String(roles.OperationNames), params); err != nil {
		return roles.Role{}, err
	}
	return ram.svc.RetrieveRole(ctx, session, entityID, roleID)
}

func (ram RoleManagerAuthorizationMiddleware) RetrieveAllRoles(ctx context.Context, session authn.Session, entityID string, limit, offset uint64) (roles.RolePage, error) {
	if err := ram.authorize(ctx, roles.OpRetrieveAllRoles, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return roles.RolePage{}, err
	}
	params := map[string]any{
		"entity_id": entityID,
		"limit":     limit,
		"offset":    offset,
	}
	if err := ram.callOut(ctx, session, roles.OpRetrieveAllRoles.String(roles.OperationNames), params); err != nil {
		return roles.RolePage{}, err
	}
	return ram.svc.RetrieveAllRoles(ctx, session, entityID, limit, offset)
}

func (ram RoleManagerAuthorizationMiddleware) ListAvailableActions(ctx context.Context, session authn.Session) ([]string, error) {
	params := map[string]any{}
	if err := ram.callOut(ctx, session, roles.OpListAvailableActions.String(roles.OperationNames), params); err != nil {
		return []string{}, err
	}
	return ram.svc.ListAvailableActions(ctx, session)
}

func (ram RoleManagerAuthorizationMiddleware) RoleAddActions(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) (ops []string, err error) {
	if err := ram.authorize(ctx, roles.OpRoleAddActions, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return []string{}, err
	}

	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
		"actions":   actions,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleAddActions.String(roles.OperationNames), params); err != nil {
		return []string{}, err
	}

	return ram.svc.RoleAddActions(ctx, session, entityID, roleID, actions)
}

func (ram RoleManagerAuthorizationMiddleware) RoleListActions(ctx context.Context, session authn.Session, entityID, roleID string) ([]string, error) {
	if err := ram.authorize(ctx, roles.OpRoleListActions, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return []string{}, err
	}

	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleListActions.String(roles.OperationNames), params); err != nil {
		return []string{}, err
	}

	return ram.svc.RoleListActions(ctx, session, entityID, roleID)
}

func (ram RoleManagerAuthorizationMiddleware) RoleCheckActionsExists(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) (bool, error) {
	if err := ram.authorize(ctx, roles.OpRoleCheckActionsExists, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return false, err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
		"actions":   actions,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleCheckActionsExists.String(roles.OperationNames), params); err != nil {
		return false, err
	}
	return ram.svc.RoleCheckActionsExists(ctx, session, entityID, roleID, actions)
}

func (ram RoleManagerAuthorizationMiddleware) RoleRemoveActions(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) (err error) {
	if err := ram.authorize(ctx, roles.OpRoleRemoveActions, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
		"actions":   actions,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleRemoveActions.String(roles.OperationNames), params); err != nil {
		return err
	}

	return ram.svc.RoleRemoveActions(ctx, session, entityID, roleID, actions)
}

func (ram RoleManagerAuthorizationMiddleware) RoleRemoveAllActions(ctx context.Context, session authn.Session, entityID, roleID string) error {
	if err := ram.authorize(ctx, roles.OpRoleRemoveAllActions, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleRemoveAllActions.String(roles.OperationNames), params); err != nil {
		return err
	}
	return ram.svc.RoleRemoveAllActions(ctx, session, entityID, roleID)
}

func (ram RoleManagerAuthorizationMiddleware) RoleAddMembers(ctx context.Context, session authn.Session, entityID, roleID string, members []string) ([]string, error) {
	if err := ram.authorize(ctx, roles.OpRoleAddMembers, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return []string{}, err
	}

	if err := ram.validateMembers(ctx, session, members); err != nil {
		return []string{}, err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
		"members":   members,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleAddMembers.String(roles.OperationNames), params); err != nil {
		return []string{}, err
	}
	return ram.svc.RoleAddMembers(ctx, session, entityID, roleID, members)
}

func (ram RoleManagerAuthorizationMiddleware) RoleListMembers(ctx context.Context, session authn.Session, entityID, roleID string, limit, offset uint64) (roles.MembersPage, error) {
	if err := ram.authorize(ctx, roles.OpRoleListMembers, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return roles.MembersPage{}, err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
		"limit":     limit,
		"offset":    offset,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleListMembers.String(roles.OperationNames), params); err != nil {
		return roles.MembersPage{}, err
	}
	return ram.svc.RoleListMembers(ctx, session, entityID, roleID, limit, offset)
}

func (ram RoleManagerAuthorizationMiddleware) RoleCheckMembersExists(ctx context.Context, session authn.Session, entityID, roleID string, members []string) (bool, error) {
	if err := ram.authorize(ctx, roles.OpRoleCheckMembersExists, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return false, err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
		"members":   members,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleCheckMembersExists.String(roles.OperationNames), params); err != nil {
		return false, err
	}
	return ram.svc.RoleCheckMembersExists(ctx, session, entityID, roleID, members)
}

func (ram RoleManagerAuthorizationMiddleware) RoleRemoveAllMembers(ctx context.Context, session authn.Session, entityID, roleID string) (err error) {
	if err := ram.authorize(ctx, roles.OpRoleRemoveAllMembers, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleRemoveAllMembers.String(roles.OperationNames), params); err != nil {
		return err
	}
	return ram.svc.RoleRemoveAllMembers(ctx, session, entityID, roleID)
}

func (ram RoleManagerAuthorizationMiddleware) ListEntityMembers(ctx context.Context, session authn.Session, entityID string, pageQuery roles.MembersRolePageQuery) (roles.MembersRolePage, error) {
	if err := ram.authorize(ctx, roles.OpRoleListMembers, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return roles.MembersRolePage{}, err
	}
	params := map[string]any{
		"entity_id":  entityID,
		"page_query": pageQuery,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleListMembers.String(roles.OperationNames), params); err != nil {
		return roles.MembersRolePage{}, err
	}
	return ram.svc.ListEntityMembers(ctx, session, entityID, pageQuery)
}

func (ram RoleManagerAuthorizationMiddleware) RemoveEntityMembers(ctx context.Context, session authn.Session, entityID string, members []string) error {
	if err := ram.authorize(ctx, roles.OpRoleRemoveAllMembers, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return err
	}
	params := map[string]any{
		"entity_id": entityID,
		"members":   members,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleRemoveAllMembers.String(roles.OperationNames), params); err != nil {
		return err
	}
	return ram.svc.RemoveEntityMembers(ctx, session, entityID, members)
}

func (ram RoleManagerAuthorizationMiddleware) RoleRemoveMembers(ctx context.Context, session authn.Session, entityID, roleID string, members []string) (err error) {
	if err := ram.authorize(ctx, roles.OpRoleRemoveMembers, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      entityID,
		ObjectType:  ram.entityType,
	}); err != nil {
		return err
	}
	params := map[string]any{
		"entity_id": entityID,
		"role_id":   roleID,
		"members":   members,
	}
	if err := ram.callOut(ctx, session, roles.OpRoleRemoveMembers.String(roles.OperationNames), params); err != nil {
		return err
	}
	return ram.svc.RoleRemoveMembers(ctx, session, entityID, roleID, members)
}

func (ram RoleManagerAuthorizationMiddleware) authorize(ctx context.Context, op svcutil.Operation, pr smqauthz.PolicyReq) error {
	perm, err := ram.opp.GetPermission(op)
	if err != nil {
		return err
	}

	pr.Permission = perm.String()

	if err := ram.authz.Authorize(ctx, pr); err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}

	return nil
}

func (ram RoleManagerAuthorizationMiddleware) RemoveMemberFromAllRoles(ctx context.Context, session authn.Session, memberID string) (err error) {
	return ram.svc.RemoveMemberFromAllRoles(ctx, session, memberID)
}

func (ram RoleManagerAuthorizationMiddleware) validateMembers(ctx context.Context, session authn.Session, members []string) error {
	switch ram.entityType {
	case policies.DomainType:
		for _, member := range members {
			if err := ram.authz.Authorize(ctx, smqauthz.PolicyReq{
				Permission:  policies.MembershipPermission,
				Subject:     member,
				SubjectType: policies.UserType,
				SubjectKind: policies.UsersKind,
				Object:      policies.SuperMQObject,
				ObjectType:  policies.PlatformType,
			}); err != nil {
				return errors.Wrap(errors.ErrMissingMember, err)
			}
		}
		return nil

	default:
		for _, member := range members {
			if err := ram.authz.Authorize(ctx, smqauthz.PolicyReq{
				Permission:  policies.MembershipPermission,
				Subject:     policies.EncodeDomainUserID(session.DomainID, member),
				SubjectType: policies.UserType,
				SubjectKind: policies.UsersKind,
				Object:      session.DomainID,
				ObjectType:  policies.DomainType,
			}); err != nil {
				return errors.Wrap(errors.ErrMissingDomainMember, err)
			}
		}
		return nil
	}
}

func (ram RoleManagerAuthorizationMiddleware) callOut(ctx context.Context, session authn.Session, op string, params map[string]interface{}) error {
	pl := map[string]any{
		"entity_type":  ram.entityType,
		"subject_type": policies.UserType,
		"subject_id":   session.UserID,
		"domain":       session.DomainID,
		"time":         time.Now().UTC(),
	}

	maps.Copy(params, pl)

	if err := ram.callout.Callout(ctx, op, params); err != nil {
		return err
	}

	return nil
}

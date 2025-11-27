// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
)

var _ roles.RoleManager = (*RoleManagerCalloutMiddleware)(nil)

type RoleManagerCalloutMiddleware struct {
	entityType string
	svc        roles.RoleManager
	callout    callout.Callout
}

func NewCallout(entityType string, svc roles.RoleManager, callout callout.Callout) (RoleManagerCalloutMiddleware, error) {
	return RoleManagerCalloutMiddleware{
		svc:        svc,
		callout:    callout,
		entityType: entityType,
	}, nil
}

func (rcm *RoleManagerCalloutMiddleware) AddRole(ctx context.Context, session authn.Session, entityID, roleName string, optionalActions []string, optionalMembers []string) (roles.RoleProvision, error) {
	params := map[string]any{
		"role_name":        roleName,
		"optional_actions": optionalActions,
		"optional_members": optionalMembers,
		"count":            1,
	}
	if err := rcm.callOut(ctx, session, roles.OpAddRole.String(roles.OperationNames), entityID, params); err != nil {
		return roles.RoleProvision{}, err
	}
	return rcm.svc.AddRole(ctx, session, entityID, roleName, optionalActions, optionalMembers)
}

func (rcm *RoleManagerCalloutMiddleware) RemoveRole(ctx context.Context, session authn.Session, entityID, roleID string) error {
	params := map[string]any{
		"role_id": roleID,
	}
	if err := rcm.callOut(ctx, session, roles.OpRemoveRole.String(roles.OperationNames), entityID, params); err != nil {
		return err
	}
	return rcm.svc.RemoveRole(ctx, session, entityID, roleID)
}

func (rcm *RoleManagerCalloutMiddleware) UpdateRoleName(ctx context.Context, session authn.Session, entityID, roleID, newRoleName string) (roles.Role, error) {
	params := map[string]any{
		"role_id":       roleID,
		"new_role_name": newRoleName,
	}
	if err := rcm.callOut(ctx, session, roles.OpUpdateRoleName.String(roles.OperationNames), entityID, params); err != nil {
		return roles.Role{}, err
	}
	return rcm.svc.UpdateRoleName(ctx, session, entityID, roleID, newRoleName)
}

func (rcm *RoleManagerCalloutMiddleware) RetrieveRole(ctx context.Context, session authn.Session, entityID, roleID string) (roles.Role, error) {
	params := map[string]any{
		"role_id": roleID,
	}
	if err := rcm.callOut(ctx, session, roles.OpRetrieveRole.String(roles.OperationNames), entityID, params); err != nil {
		return roles.Role{}, err
	}
	return rcm.svc.RetrieveRole(ctx, session, entityID, roleID)
}

func (rcm *RoleManagerCalloutMiddleware) RetrieveAllRoles(ctx context.Context, session authn.Session, entityID string, limit, offset uint64) (roles.RolePage, error) {
	params := map[string]any{
		"limit":  limit,
		"offset": offset,
	}
	if err := rcm.callOut(ctx, session, roles.OpRetrieveAllRoles.String(roles.OperationNames), entityID, params); err != nil {
		return roles.RolePage{}, err
	}
	return rcm.svc.RetrieveAllRoles(ctx, session, entityID, limit, offset)
}

func (rcm *RoleManagerCalloutMiddleware) ListAvailableActions(ctx context.Context, session authn.Session) ([]string, error) {
	if err := rcm.callOut(ctx, session, roles.OpListAvailableActions.String(roles.OperationNames), "", nil); err != nil {
		return []string{}, err
	}
	return rcm.svc.ListAvailableActions(ctx, session)
}

func (rcm *RoleManagerCalloutMiddleware) RoleAddActions(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) ([]string, error) {
	params := map[string]any{
		"role_id": roleID,
		"actions": actions,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleAddActions.String(roles.OperationNames), entityID, params); err != nil {
		return []string{}, err
	}
	return rcm.svc.RoleAddActions(ctx, session, entityID, roleID, actions)
}

func (rcm *RoleManagerCalloutMiddleware) RoleListActions(ctx context.Context, session authn.Session, entityID, roleID string) ([]string, error) {
	params := map[string]any{
		"role_id": roleID,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleListActions.String(roles.OperationNames), entityID, params); err != nil {
		return []string{}, err
	}
	return rcm.svc.RoleListActions(ctx, session, entityID, roleID)
}

func (rcm *RoleManagerCalloutMiddleware) RoleCheckActionsExists(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) (bool, error) {
	params := map[string]any{
		"role_id": roleID,
		"actions": actions,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleCheckActionsExists.String(roles.OperationNames), entityID, params); err != nil {
		return false, err
	}
	return rcm.svc.RoleCheckActionsExists(ctx, session, entityID, roleID, actions)
}

func (rcm *RoleManagerCalloutMiddleware) RoleRemoveActions(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) error {
	params := map[string]any{
		"role_id": roleID,
		"actions": actions,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleRemoveActions.String(roles.OperationNames), entityID, params); err != nil {
		return err
	}
	return rcm.svc.RoleRemoveActions(ctx, session, entityID, roleID, actions)
}

func (rcm *RoleManagerCalloutMiddleware) RoleRemoveAllActions(ctx context.Context, session authn.Session, entityID, roleID string) error {
	params := map[string]any{
		"role_id": roleID,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleRemoveAllActions.String(roles.OperationNames), entityID, params); err != nil {
		return err
	}
	return rcm.svc.RoleRemoveAllActions(ctx, session, entityID, roleID)
}

func (rcm *RoleManagerCalloutMiddleware) RoleAddMembers(ctx context.Context, session authn.Session, entityID, roleID string, members []string) ([]string, error) {
	params := map[string]any{
		"role_id": roleID,
		"members": members,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleAddMembers.String(roles.OperationNames), entityID, params); err != nil {
		return []string{}, err
	}
	return rcm.svc.RoleAddMembers(ctx, session, entityID, roleID, members)
}

func (rcm *RoleManagerCalloutMiddleware) RoleListMembers(ctx context.Context, session authn.Session, entityID, roleID string, limit, offset uint64) (roles.MembersPage, error) {
	params := map[string]any{
		"role_id": roleID,
		"limit":   limit,
		"offset":  offset,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleListMembers.String(roles.OperationNames), entityID, params); err != nil {
		return roles.MembersPage{}, err
	}
	return rcm.svc.RoleListMembers(ctx, session, entityID, roleID, limit, offset)
}

func (rcm *RoleManagerCalloutMiddleware) RoleCheckMembersExists(ctx context.Context, session authn.Session, entityID, roleID string, members []string) (bool, error) {
	params := map[string]any{
		"role_id": roleID,
		"members": members,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleCheckMembersExists.String(roles.OperationNames), entityID, params); err != nil {
		return false, err
	}
	return rcm.svc.RoleCheckMembersExists(ctx, session, entityID, roleID, members)
}

func (rcm *RoleManagerCalloutMiddleware) RoleRemoveAllMembers(ctx context.Context, session authn.Session, entityID, roleID string) error {
	params := map[string]any{
		"role_id": roleID,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleRemoveAllMembers.String(roles.OperationNames), entityID, params); err != nil {
		return err
	}
	return rcm.svc.RoleRemoveAllMembers(ctx, session, entityID, roleID)
}

func (rcm *RoleManagerCalloutMiddleware) ListEntityMembers(ctx context.Context, session authn.Session, entityID string, pageQuery roles.MembersRolePageQuery) (roles.MembersRolePage, error) {
	params := map[string]any{
		"page_query": pageQuery,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleListMembers.String(roles.OperationNames), entityID, params); err != nil {
		return roles.MembersRolePage{}, err
	}
	return rcm.svc.ListEntityMembers(ctx, session, entityID, pageQuery)
}

func (rcm *RoleManagerCalloutMiddleware) RemoveEntityMembers(ctx context.Context, session authn.Session, entityID string, members []string) error {
	params := map[string]any{
		"members": members,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleRemoveAllMembers.String(roles.OperationNames), entityID, params); err != nil {
		return err
	}
	return rcm.svc.RemoveEntityMembers(ctx, session, entityID, members)
}

func (rcm *RoleManagerCalloutMiddleware) RoleRemoveMembers(ctx context.Context, session authn.Session, entityID, roleID string, members []string) error {
	params := map[string]any{
		"role_id": roleID,
		"members": members,
	}
	if err := rcm.callOut(ctx, session, roles.OpRoleRemoveMembers.String(roles.OperationNames), entityID, params); err != nil {
		return err
	}
	return rcm.svc.RoleRemoveMembers(ctx, session, entityID, roleID, members)
}

func (rcm *RoleManagerCalloutMiddleware) RemoveMemberFromAllRoles(ctx context.Context, session authn.Session, memberID string) error {
	return rcm.svc.RemoveMemberFromAllRoles(ctx, session, memberID)
}

func (rcm *RoleManagerCalloutMiddleware) callOut(ctx context.Context, session authn.Session, op, entityID string, pld map[string]any) error {
	req := callout.Request{
		BaseRequest: callout.BaseRequest{
			Operation:  op,
			EntityType: rcm.entityType,
			EntityID:   entityID,
			CallerID:   session.UserID,
			CallerType: policies.UserType,
			DomainID:   session.DomainID,
			Time:       time.Now().UTC(),
		},
		Payload: pld,
	}

	if err := rcm.callout.Callout(ctx, req); err != nil {
		return err
	}

	return nil
}

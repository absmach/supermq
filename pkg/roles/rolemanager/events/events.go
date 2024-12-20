// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/roles"
)

const (
	addRole                  = "role.add"
	removeRole               = "role.remove"
	updateRole               = "role.update"
	viewRole                 = "role.view"
	viewAllRole              = "role.view_all"
	listAvailableActions     = "role.list_available_actions"
	addRoleActions           = "role.actions.add"
	listRoleActions          = "role.actions.ist"
	checkRoleActions         = "role.actions.check"
	removeRoleActions        = "role.actions.remove"
	removeAllRoleActions     = "role.actions.remove_all"
	addRoleMembers           = "role.members.add"
	listRoleMembers          = "role.members.list"
	checkRoleMembers         = "role.members.check"
	removeRoleMembers        = "role.members.remove"
	removeRoleAllMembers     = "role.members.remove_all"
	removeMemberFromAllRoles = "role.members.remove_from_all_roles"
)

var (
	_ events.Event = (*addRoleEvent)(nil)
	_ events.Event = (*removeRoleEvent)(nil)
	_ events.Event = (*updateRoleEvent)(nil)
	_ events.Event = (*retrieveRoleEvent)(nil)
	_ events.Event = (*retrieveAllRolesEvent)(nil)
	_ events.Event = (*listAvailableActionsEvent)(nil)
	_ events.Event = (*roleAddActionsEvent)(nil)
	_ events.Event = (*roleListActionsEvent)(nil)
	_ events.Event = (*roleCheckActionsExistsEvent)(nil)
	_ events.Event = (*roleRemoveActionsEvent)(nil)
	_ events.Event = (*roleRemoveAllActionsEvent)(nil)
	_ events.Event = (*roleAddMembersEvent)(nil)
	_ events.Event = (*roleListMembersEvent)(nil)
	_ events.Event = (*roleCheckMembersExistsEvent)(nil)
	_ events.Event = (*roleRemoveMembersEvent)(nil)
	_ events.Event = (*roleRemoveAllMembersEvent)(nil)
	_ events.Event = (*removeMemberFromAllRolesEvent)(nil)
)

type addRoleEvent struct {
	operationPrefix string
	roles.Role
}

func (are addRoleEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation":  are.operationPrefix + addRole,
		"id":         are.ID,
		"name":       are.Name,
		"entity_id":  are.EntityID,
		"created_by": are.CreatedBy,
		"created_at": are.CreatedAt,
		"updated_by": are.UpdatedBy,
		"updated_at": are.UpdatedAt,
	}
	return val, nil
}

type removeRoleEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
}

func (rre removeRoleEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rre.operationPrefix + removeRole,
		"entity_id": rre.entityID,
		"role_id":   rre.roleID,
	}
	return val, nil
}

type updateRoleEvent struct {
	operationPrefix string
	roles.Role
}

func (ure updateRoleEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation":  ure.operationPrefix + updateRole,
		"id":         ure.ID,
		"name":       ure.Name,
		"entity_id":  ure.EntityID,
		"created_by": ure.CreatedBy,
		"created_at": ure.CreatedAt,
		"updated_by": ure.UpdatedBy,
		"updated_at": ure.UpdatedAt,
	}
	return val, nil
}

type retrieveRoleEvent struct {
	operationPrefix string
	roles.Role
}

func (rre retrieveRoleEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation":  rre.operationPrefix + viewRole,
		"id":         rre.ID,
		"name":       rre.Name,
		"entity_id":  rre.EntityID,
		"created_by": rre.CreatedBy,
		"created_at": rre.CreatedAt,
		"updated_by": rre.UpdatedBy,
		"updated_at": rre.UpdatedAt,
	}
	return val, nil
}

type retrieveAllRolesEvent struct {
	operationPrefix string
	entityID        string
	limit           uint64
	offset          uint64
}

func (rare retrieveAllRolesEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rare.operationPrefix + viewAllRole,
		"entity_id": rare.entityID,
		"limit":     rare.limit,
		"offset":    rare.offset,
	}
	return val, nil
}

type listAvailableActionsEvent struct {
	operationPrefix string
}

func (laae listAvailableActionsEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": laae.operationPrefix + listAvailableActions,
	}
	return val, nil
}

type roleAddActionsEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
	actions         []string
}

func (raae roleAddActionsEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": raae.operationPrefix + addRoleActions,
		"entity_id": raae.entityID,
		"role_id":   raae.roleID,
		"actions":   raae.actions,
	}
	return val, nil
}

type roleListActionsEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
}

func (rlae roleListActionsEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rlae.operationPrefix + listRoleActions,
		"entity_id": rlae.entityID,
		"role_id":   rlae.roleID,
	}
	return val, nil
}

type roleCheckActionsExistsEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
	actions         []string
	isAllExists     bool
}

func (rcaee roleCheckActionsExistsEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation":     rcaee.operationPrefix + checkRoleActions,
		"entity_id":     rcaee.entityID,
		"role_id":       rcaee.roleID,
		"actions":       rcaee.actions,
		"is_all_exists": rcaee.isAllExists,
	}
	return val, nil
}

type roleRemoveActionsEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
	actions         []string
}

func (rrae roleRemoveActionsEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rrae.operationPrefix + removeRoleActions,
		"entity_id": rrae.entityID,
		"role_id":   rrae.roleID,
		"actions":   rrae.actions,
	}
	return val, nil
}

type roleRemoveAllActionsEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
}

func (rraae roleRemoveAllActionsEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rraae.operationPrefix + removeAllRoleActions,
		"entity_id": rraae.entityID,
		"role_id":   rraae.roleID,
	}
	return val, nil
}

type roleAddMembersEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
	members         []string
}

func (rame roleAddMembersEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rame.operationPrefix + addRoleMembers,
		"entity_id": rame.entityID,
		"role_id":   rame.roleID,
		"members":   rame.members,
	}
	return val, nil
}

type roleListMembersEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
	limit           uint64
	offset          uint64
}

func (rlme roleListMembersEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rlme.operationPrefix + listRoleMembers,
		"entity_id": rlme.entityID,
		"role_id":   rlme.roleID,
		"limit":     rlme.limit,
		"offset":    rlme.offset,
	}
	return val, nil
}

type roleCheckMembersExistsEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
	members         []string
}

func (rcmee roleCheckMembersExistsEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rcmee.operationPrefix + checkRoleMembers,
		"entity_id": rcmee.entityID,
		"role_id":   rcmee.roleID,
		"members":   rcmee.members,
	}
	return val, nil
}

type roleRemoveMembersEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
	members         []string
}

func (rrme roleRemoveMembersEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rrme.operationPrefix + removeRoleMembers,
		"entity_id": rrme.entityID,
		"role_id":   rrme.roleID,
		"members":   rrme.members,
	}
	return val, nil
}

type roleRemoveAllMembersEvent struct {
	operationPrefix string
	entityID        string
	roleID          string
}

func (rrame roleRemoveAllMembersEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rrame.operationPrefix + removeRoleAllMembers,
		"entity_id": rrame.entityID,
		"role_id":   rrame.roleID,
	}
	return val, nil
}

type removeMemberFromAllRolesEvent struct {
	operationPrefix string
	memberID        string
}

func (rmare removeMemberFromAllRolesEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation": rmare.operationPrefix + removeMemberFromAllRoles,
		"member_id": rmare.memberID,
	}
	return val, nil
}

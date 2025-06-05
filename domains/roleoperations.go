// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package domains

import "github.com/absmach/supermq/pkg/roles"

const (
	OpUpdateDomain Operation = iota
	OpRetrieveDomain
	OpEnableDomain
	OpDisableDomain
	OpSendInvitation
	OpAcceptInvitation
	OpCreateDomain
	OpFreezeDomain
	OpListDomains
	OpViewInvitation
	OpListInvitations
	OpRejectInvitation
	OpDeleteInvitation
)

var expectedOperations = []Operation{
	OpRetrieveDomain,
	OpUpdateDomain,
	OpEnableDomain,
	OpDisableDomain,
}

const (
	OpUpdateDomainStr     = "OpRetrieveDomain"
	OpRetrieveDomainStr   = "OpUpdateDomain"
	OpEnableDomainStr     = "OpEnableDomain"
	OpDisableDomainStr    = "OpDisableDomain"
	OpSendInvitationStr   = "OpSendInvitation"
	OpAcceptInvitationStr = "OpAcceptInvitation"
	OpCreateDomainStr     = "OpCreateDomain"
	OpFreezeDomainStr     = "OpFreezeDomain"
	OpListDomainsStr      = "OpListDomains"
	OpViewInvitationStr   = "OpViewInvitation"
	OpListInvitationsStr  = "OpListInvitations"
	OpRejectInvitationStr = "OpRejectInvitation"
	OpDeleteInvitationStr = "OpDeleteInvitation"
)

func NewOperationPerm() OperationPerm {
	return newOperationPerm(expectedOperations)
}

// Below codes should moved out of service, may be can be kept in `cmd/<svc>/main.go`

const (
	updatePermission          = "update_permission"
	enablePermission          = "enable_permission"
	disablePermission         = "disable_permission"
	readPermission            = "read_permission"
	deletePermission          = "delete_permission"
	manageRolePermission      = "manage_role_permission"
	addRoleUsersPermission    = "add_role_users_permission"
	removeRoleUsersPermission = "remove_role_users_permission"
	viewRoleUsersPermission   = "view_role_users_permission"
)

const (
	ClientCreatePermission  = "client_create_permission"
	ChannelCreatePermission = "channel_create_permission"
	GroupCreatePermission   = "group_create_permission"
)

func NewOperationPermissionMap() map[Operation]Permission {
	opPerm := map[Operation]Permission{
		OpRetrieveDomain: readPermission,
		OpUpdateDomain:   updatePermission,
		OpEnableDomain:   enablePermission,
		OpDisableDomain:  disablePermission,
	}
	return opPerm
}

func NewRolesOperationPermissionMap() map[roles.Operation]roles.Permission {
	opPerm := map[roles.Operation]roles.Permission{
		roles.OpAddRole:                manageRolePermission,
		roles.OpRemoveRole:             manageRolePermission,
		roles.OpUpdateRoleName:         manageRolePermission,
		roles.OpRetrieveRole:           manageRolePermission,
		roles.OpRetrieveAllRoles:       manageRolePermission,
		roles.OpRoleAddActions:         manageRolePermission,
		roles.OpRoleListActions:        manageRolePermission,
		roles.OpRoleCheckActionsExists: manageRolePermission,
		roles.OpRoleRemoveActions:      manageRolePermission,
		roles.OpRoleRemoveAllActions:   manageRolePermission,
		roles.OpRoleAddMembers:         addRoleUsersPermission,
		roles.OpRoleListMembers:        viewRoleUsersPermission,
		roles.OpRoleCheckMembersExists: viewRoleUsersPermission,
		roles.OpRoleRemoveMembers:      removeRoleUsersPermission,
		roles.OpRoleRemoveAllMembers:   manageRolePermission,
	}
	return opPerm
}

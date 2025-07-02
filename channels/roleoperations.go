// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package channels

import (
	"github.com/absmach/supermq/pkg/roles"
	"github.com/absmach/supermq/pkg/svcutil"
)

// Internal Operations.
const (
	OpViewChannel Operation = iota
	OpUpdateChannel
	OpUpdateChannelTags
	OpEnableChannel
	OpDisableChannel
	OpDeleteChannel
	OpSetParentGroup
	OpRemoveParentGroup
	OpConnectClient
	OpDisconnectClient
	OpCreateChannel
	OpListChannels
	OpListUserChannels
)

const (
	OpViewChannelStr       = "OpViewChannel"
	OpUpdateChannelStr     = "OpUpdateChannel"
	OpUpdateChannelTagsStr = "OpUpdateChannelTags"
	OpEnableChannelStr     = "OpEnableChannel"
	OpDisableChannelStr    = "OpDisableChannel"
	OpDeleteChannelStr     = "OpDeleteChannel"
	OpSetParentGroupStr    = "OpSetParentGroup"
	OpRemoveParentGroupStr = "OpRemoveParentGroup"
	OpConnectClientStr     = "OpConnectClient"
	OpDisconnectClientStr  = "OpDisconnectClient"
	OpCreateChannelStr     = "OpCreateChannel"
	OpListChannelsStr      = "OpListChannels"
	OpListUserChannelsStr  = "OpListUserChannels"
)

var expectedOperations = []Operation{
	OpViewChannel,
	OpUpdateChannel,
	OpUpdateChannelTags,
	OpEnableChannel,
	OpDisableChannel,
	OpDeleteChannel,
	OpSetParentGroup,
	OpRemoveParentGroup,
	OpConnectClient,
	OpDisconnectClient,
}

func NewOperationPerm() OperationPerm {
	return newOperationPerm(expectedOperations)
}

// External Operations.
const (
	DomainOpCreateChannel ExternalOperation = iota
	DomainOpListChannel
	GroupOpSetChildChannel
	GroupsOpRemoveChildChannel
	ClientsOpConnectChannel
	ClientsOpDisconnectChannel
)

const (
	DomainOpCreateChannelStr      = "DomainOpCreateChannel"
	DomainOpListChannelStr        = "DomainOpListChannel"
	GroupOpSetChildChannelStr     = "GroupOpSetChildChannel"
	GroupsOpRemoveChildChannelStr = "GroupsOpRemoveChildChannel"
	ClientsOpConnectChannelStr    = "ClientsOpConnectChannel"
	ClientsOpDisconnectChannelStr = "ClientsOpDisconnectChannel"
)

var expectedExternalOperations = []ExternalOperation{
	DomainOpCreateChannel,
	DomainOpListChannel,
	GroupOpSetChildChannel,
	GroupsOpRemoveChildChannel,
	ClientsOpConnectChannel,
	ClientsOpDisconnectChannel,
}

func NewExternalOperationPerm() ExternalOperationPerm {
	return newExternalOperationPerm(expectedExternalOperations)
}

// Below codes should moved out of service, may be can be kept in `cmd/<svc>/main.go`

const (
	updatePermission          = "update_permission"
	readPermission            = "read_permission"
	deletePermission          = "delete_permission"
	setParentGroupPermission  = "set_parent_group_permission"
	connectToClientPermission = "connect_to_client_permission"

	manageRolePermission      = "manage_role_permission"
	addRoleUsersPermission    = "add_role_users_permission"
	removeRoleUsersPermission = "remove_role_users_permission"
	viewRoleUsersPermission   = "view_role_users_permission"
)

func NewOperationPermissionMap() map[Operation]Permission {
	opPerm := map[Operation]Permission{
		OpViewChannel:       readPermission,
		OpUpdateChannel:     updatePermission,
		OpUpdateChannelTags: updatePermission,
		OpEnableChannel:     updatePermission,
		OpDisableChannel:    updatePermission,
		OpDeleteChannel:     deletePermission,
		OpSetParentGroup:    setParentGroupPermission,
		OpRemoveParentGroup: setParentGroupPermission,
		OpConnectClient:     connectToClientPermission,
		OpDisconnectClient:  connectToClientPermission,
	}
	return opPerm
}

func NewRolesOperationPermissionMap() map[svcutil.Operation]svcutil.Permission {
	opPerm := map[svcutil.Operation]svcutil.Permission{
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

const (
	// External Permission
	// Domains.
	domainCreateChannelPermission = "channel_create_permission"
	domainListChanelPermission    = "channel_read_permission"
	// Groups.
	groupSetChildChannelPermission    = "channel_create_permission"
	groupRemoveChildChannelPermission = "channel_create_permission"
	// Client.
	clientsConnectChannelPermission    = "connect_to_channel_permission"
	clientsDisconnectChannelPermission = "connect_to_channel_permission"
)

func NewExternalOperationPermissionMap() map[ExternalOperation]Permission {
	extOpPerm := map[ExternalOperation]Permission{
		DomainOpCreateChannel:      domainCreateChannelPermission,
		DomainOpListChannel:        domainListChanelPermission,
		GroupOpSetChildChannel:     groupSetChildChannelPermission,
		GroupsOpRemoveChildChannel: groupRemoveChildChannelPermission,
		ClientsOpConnectChannel:    clientsConnectChannelPermission,
		ClientsOpDisconnectChannel: clientsDisconnectChannelPermission,
	}
	return extOpPerm
}

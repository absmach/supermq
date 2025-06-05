// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package clients

import "github.com/absmach/supermq/pkg/roles"

// Internal Operations.
const (
	OpViewClient Operation = iota
	OpUpdateClient
	OpUpdateClientTags
	OpUpdateClientSecret
	OpEnableClient
	OpDisableClient
	OpDeleteClient
	OpSetParentGroup
	OpRemoveParentGroup
	OpConnectToChannel
	OpDisconnectFromChannel
	OpCreateClient
	OpListClients
	OpListUserClients
)

const (
	OpViewClientStr            = "OpViewClient"
	OpUpdateClientStr          = "OpUpdateClient"
	OpUpdateClientTagsStr      = "OpUpdateClientTags"
	OpUpdateClientSecretStr    = "OpUpdateClientSecret"
	OpEnableClientStr          = "OpEnableClient"
	OpDisableClientStr         = "OpDisableClient"
	OpDeleteClientStr          = "OpDeleteClient"
	OpSetParentGroupStr        = "OpSetParentGroup"
	OpRemoveParentGroupStr     = "OpRemoveParentGroup"
	OpConnectToChannelStr      = "OpConnectToChannel"
	OpDisconnectFromChannelStr = "OpDisconnectFromChannel"
	OpCreateClientStr          = "OpCreateClient"
	OpListClientsStr           = "OpListClients"
	OpListUserClientsStr       = "OpListUserClients"
)

var expectedOperations = []Operation{
	OpViewClient,
	OpUpdateClient,
	OpUpdateClientTags,
	OpUpdateClientSecret,
	OpEnableClient,
	OpDisableClient,
	OpDeleteClient,
	OpSetParentGroup,
	OpRemoveParentGroup,
	OpConnectToChannel,
	OpDisconnectFromChannel,
}

func NewOperationPerm() OperationPerm {
	return newOperationPerm(expectedOperations)
}

// External Operations.
const (
	DomainOpCreateClient ExternalOperation = iota
	DomainOpListClients
	GroupOpSetChildClient
	GroupsOpRemoveChildClient
	ChannelsOpConnectChannel
	ChannelsOpDisconnectChannel
)

const (
	DomainOpCreateClientStr        = "DomainOpCreateClient"
	DomainOpListClientsStr         = "DomainOpListClients"
	GroupOpSetChildClientStr       = "GroupOpSetChildClient"
	GroupsOpRemoveChildClientStr   = "GroupsOpRemoveChildClient"
	ChannelsOpConnectChannelStr    = "ChannelsOpConnectChannel"
	ChannelsOpDisconnectChannelStr = "ChannelsOpDisconnectChannel"
)

var expectedExternalOperations = []ExternalOperation{
	DomainOpCreateClient,
	DomainOpListClients,
	GroupOpSetChildClient,
	GroupsOpRemoveChildClient,
	ChannelsOpConnectChannel,
	ChannelsOpDisconnectChannel,
}

func NewExternalOperationPerm() ExternalOperationPerm {
	return newExternalOperationPerm(expectedExternalOperations)
}

// Below codes should moved out of service, may be can be kept in `cmd/<svc>/main.go`

const (
	updatePermission           = "update_permission"
	readPermission             = "read_permission"
	deletePermission           = "delete_permission"
	setParentGroupPermission   = "set_parent_group_permission"
	connectToChannelPermission = "connect_to_channel_permission"

	manageRolePermission      = "manage_role_permission"
	addRoleUsersPermission    = "add_role_users_permission"
	removeRoleUsersPermission = "remove_role_users_permission"
	viewRoleUsersPermission   = "view_role_users_permission"
)

func NewOperationPermissionMap() map[Operation]Permission {
	opPerm := map[Operation]Permission{
		OpViewClient:            readPermission,
		OpUpdateClient:          updatePermission,
		OpUpdateClientTags:      updatePermission,
		OpUpdateClientSecret:    updatePermission,
		OpEnableClient:          updatePermission,
		OpDisableClient:         updatePermission,
		OpDeleteClient:          deletePermission,
		OpSetParentGroup:        setParentGroupPermission,
		OpRemoveParentGroup:     setParentGroupPermission,
		OpConnectToChannel:      connectToChannelPermission,
		OpDisconnectFromChannel: connectToChannelPermission,
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

const (
	// External Permission for domains.
	domainCreateClientPermission = "client_create_permission"
	domainListClientsPermission  = "client_read_permission"
	// External Permission for groups.
	groupSetChildClientPermission    = "client_create_permission"
	groupRemoveChildClientPermission = "client_create_permission"
	// External Permission for channels.
	channelsConnectClientPermission    = "connect_to_client_permission"
	channelsDisconnectClientPermission = "connect_to_client_permission"
)

func NewExternalOperationPermissionMap() map[ExternalOperation]Permission {
	extOpPerm := map[ExternalOperation]Permission{
		DomainOpCreateClient:        domainCreateClientPermission,
		DomainOpListClients:         domainListClientsPermission,
		GroupOpSetChildClient:       groupSetChildClientPermission,
		GroupsOpRemoveChildClient:   groupRemoveChildClientPermission,
		ChannelsOpConnectChannel:    channelsConnectClientPermission,
		ChannelsOpDisconnectChannel: channelsDisconnectClientPermission,
	}
	return extOpPerm
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package clients

// Internal Operations
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

var OperationNames = []string{
	"OpViewClient",
	"OpUpdateClient",
	"OpUpdateClientTags",
	"OpUpdateClientSecret",
	"OpEnableClient",
	"OpDisableClient",
	"OpDeleteClient",
	"OpSetParentGroup",
	"OpRemoveParentGroup",
	"OpConnectToChannel",
	"OpDisconnectFromChannel",
	"OpCreateClient",
	"OpListClients",
	"OpListUserClients",
}

func NewOperationPerm() OperationPerm {
	return newOperationPerm(expectedOperations, OperationNames)
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

var expectedExternalOperations = []ExternalOperation{
	DomainOpCreateClient,
	DomainOpListClients,
	GroupOpSetChildClient,
	GroupsOpRemoveChildClient,
	ChannelsOpConnectChannel,
	ChannelsOpDisconnectChannel,
}

var externalOperationNames = []string{
	"DomainOpCreateClient",
	"DomainOpListClients",
	"GroupOpSetChildClient",
	"GroupsOpRemoveChildClient",
	"ChannelsOpConnectChannel",
	"ChannelsOpDisconnectChannel",
}

func NewExternalOperationPerm() ExternalOperationPerm {
	return newExternalOperationPerm(expectedExternalOperations, externalOperationNames)
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

const (
	OpAddRole Operation = iota
	OpRemoveRole
	OpUpdateRoleName
	OpRetrieveRole
	OpRetrieveAllRoles
	OpRoleAddActions
	OpRoleListActions
	OpRoleCheckActionsExists
	OpRoleRemoveActions
	OpRoleRemoveAllActions
	OpRoleAddMembers
	OpRoleListMembers
	OpRoleCheckMembersExists
	OpRoleRemoveMembers
	OpRoleRemoveAllMembers
)

var rolesExpectedOperations = []Operation{
	OpAddRole,
	OpRemoveRole,
	OpUpdateRoleName,
	OpRetrieveRole,
	OpRetrieveAllRoles,
	OpRoleAddActions,
	OpRoleListActions,
	OpRoleCheckActionsExists,
	OpRoleRemoveActions,
	OpRoleRemoveAllActions,
	OpRoleAddMembers,
	OpRoleListMembers,
	OpRoleCheckMembersExists,
	OpRoleRemoveMembers,
	OpRoleRemoveAllMembers,
}

var operationNames = []string{
	"OpAddRole",
	"OpRemoveRole",
	"OpUpdateRoleName",
	"OpRetrieveRole",
	"OpRetrieveAllRoles",
	"OpRoleAddActions",
	"OpRoleListActions",
	"OpRoleCheckActionsExists",
	"OpRoleRemoveActions",
	"OpRoleRemoveAllActions",
	"OpRoleAddMembers",
	"OpRoleListMembers",
	"OpRoleCheckMembersExists",
	"OpRoleRemoveMembers",
	"OpRoleRemoveAllMembers",
}

func NewRolesOperationPermissionMap() map[Operation]Permission {
	opPerm := map[Operation]Permission{
		OpAddRole:                manageRolePermission,
		OpRemoveRole:             manageRolePermission,
		OpUpdateRoleName:         manageRolePermission,
		OpRetrieveRole:           manageRolePermission,
		OpRetrieveAllRoles:       manageRolePermission,
		OpRoleAddActions:         manageRolePermission,
		OpRoleListActions:        manageRolePermission,
		OpRoleCheckActionsExists: manageRolePermission,
		OpRoleRemoveActions:      manageRolePermission,
		OpRoleRemoveAllActions:   manageRolePermission,
		OpRoleAddMembers:         addRoleUsersPermission,
		OpRoleListMembers:        viewRoleUsersPermission,
		OpRoleCheckMembersExists: viewRoleUsersPermission,
		OpRoleRemoveMembers:      removeRoleUsersPermission,
		OpRoleRemoveAllMembers:   manageRolePermission,
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

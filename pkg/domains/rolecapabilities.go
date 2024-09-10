package domains

import "github.com/absmach/magistrala/pkg/roles"

const (
	Update                 roles.Capability = "update"
	Read                   roles.Capability = "read"
	Delete                 roles.Capability = "delete"
	Membership             roles.Capability = "membership"
	ManageRole             roles.Capability = "manage_role"
	AddRoleUsers           roles.Capability = "add_role_users"
	RemoveRoleUsers        roles.Capability = "remove_role_users"
	ViewRoleUsers          roles.Capability = "view_role_users"
	ThingCreate            roles.Capability = "thing_create"
	ThingList              roles.Capability = "thing_list"
	ChannelCreate          roles.Capability = "channel_create"
	ChannelList            roles.Capability = "channel_list"
	GroupCreate            roles.Capability = "group_create"
	GroupList              roles.Capability = "group_list"
	ThingUpdate            roles.Capability = "thing_update"
	ThingRead              roles.Capability = "thing_read"
	ThingDelete            roles.Capability = "thing_delete"
	ThingSetParentGroup    roles.Capability = "thing_set_parent_group"
	ThingConnectToChannel  roles.Capability = "thing_connect_to_channel"
	ThingManageRole        roles.Capability = "thing_manage_role"
	ThingAddRoleUsers      roles.Capability = "thing_add_role_users"
	ThingRemoveRoleUsers   roles.Capability = "thing_remove_role_users"
	ThingViewRoleUsers     roles.Capability = "thing_view_role_users"
	ChannelUpdate          roles.Capability = "channel_update"
	ChannelRead            roles.Capability = "channel_read"
	ChannelDelete          roles.Capability = "channel_delete"
	ChannelSetParentGroup  roles.Capability = "channel_set_parent_group"
	ChannelConnectToThing  roles.Capability = "channel_connect_to_thing"
	ChannelPublish         roles.Capability = "channel_publish"
	ChannelSubscribe       roles.Capability = "channel_subscribe"
	ChannelManageRole      roles.Capability = "channel_manage_role"
	ChannelAddRoleUsers    roles.Capability = "channel_add_role_users"
	ChannelRemoveRoleUsers roles.Capability = "channel_remove_role_users"
	ChannelViewRoleUsers   roles.Capability = "channel_view_role_users"
	GroupUpdate            roles.Capability = "group_update"
	GroupRead              roles.Capability = "group_read"
	GroupDelete            roles.Capability = "group_delete"
	GroupSetChild          roles.Capability = "group_set_child"
	GroupSetParent         roles.Capability = "group_set_parent"
	GroupManageRole        roles.Capability = "group_manage_role"
	GroupAddRoleUsers      roles.Capability = "group_add_role_users"
	GroupRemoveRoleUsers   roles.Capability = "group_remove_role_users"
	GroupViewRoleUsers     roles.Capability = "group_view_role_users"
)

const (
	BuiltInRoleAdmin      = "admin"
	BuiltInRoleMembership = "membership"
)

func AvailableCapabilities() []roles.Capability {
	return []roles.Capability{
		Update,
		Read,
		Delete,
		Membership,
		ManageRole,
		AddRoleUsers,
		RemoveRoleUsers,
		ViewRoleUsers,
		ThingCreate,
		ThingUpdate,
		ThingRead,
		ThingDelete,
		ThingSetParentGroup,
		ThingConnectToChannel,
		ThingManageRole,
		ThingAddRoleUsers,
		ThingRemoveRoleUsers,
		ThingViewRoleUsers,
		ChannelCreate,
		ChannelUpdate,
		ChannelRead,
		ChannelDelete,
		ChannelSetParentGroup,
		ChannelConnectToThing,
		ChannelPublish,
		ChannelSubscribe,
		ChannelManageRole,
		ChannelAddRoleUsers,
		ChannelRemoveRoleUsers,
		ChannelViewRoleUsers,
		GroupCreate,
		GroupUpdate,
		GroupRead,
		GroupDelete,
		GroupSetChild,
		GroupSetParent,
		GroupManageRole,
		GroupAddRoleUsers,
		GroupRemoveRoleUsers,
		GroupViewRoleUsers,
	}
}

func membershipRoleCapabilities() []roles.Capability {
	return []roles.Capability{
		Membership,
	}
}

func BuiltInRoles() map[roles.BuiltInRoleName][]roles.Capability {
	return map[roles.BuiltInRoleName][]roles.Capability{
		"admin":      AvailableCapabilities(),
		"membership": membershipRoleCapabilities(),
	}
}

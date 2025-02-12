// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	channels "github.com/absmach/supermq/channels"
	authn "github.com/absmach/supermq/pkg/authn"

	connections "github.com/absmach/supermq/pkg/connections"

	context "context"

	mock "github.com/stretchr/testify/mock"

	roles "github.com/absmach/supermq/pkg/roles"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// AddRole provides a mock function with given fields: ctx, session, entityID, roleName, optionalActions, optionalMembers
func (_m *Service) AddRole(ctx context.Context, session authn.Session, entityID string, roleName string, optionalActions []string, optionalMembers []string) (roles.RoleProvision, error) {
	ret := _m.Called(ctx, session, entityID, roleName, optionalActions, optionalMembers)

	if len(ret) == 0 {
		panic("no return value specified for AddRole")
	}

	var r0 roles.RoleProvision
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string, []string) (roles.RoleProvision, error)); ok {
		return rf(ctx, session, entityID, roleName, optionalActions, optionalMembers)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string, []string) roles.RoleProvision); ok {
		r0 = rf(ctx, session, entityID, roleName, optionalActions, optionalMembers)
	} else {
		r0 = ret.Get(0).(roles.RoleProvision)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleName, optionalActions, optionalMembers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Connect provides a mock function with given fields: ctx, session, chIDs, clIDs, connType
func (_m *Service) Connect(ctx context.Context, session authn.Session, chIDs []string, clIDs []string, connType []connections.ConnType) error {
	ret := _m.Called(ctx, session, chIDs, clIDs, connType)

	if len(ret) == 0 {
		panic("no return value specified for Connect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, []string, []string, []connections.ConnType) error); ok {
		r0 = rf(ctx, session, chIDs, clIDs, connType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateChannels provides a mock function with given fields: ctx, session, _a2
func (_m *Service) CreateChannels(ctx context.Context, session authn.Session, _a2 ...channels.Channel) ([]channels.Channel, []roles.RoleProvision, error) {
	_va := make([]interface{}, len(_a2))
	for _i := range _a2 {
		_va[_i] = _a2[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, session)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for CreateChannels")
	}

	var r0 []channels.Channel
	var r1 []roles.RoleProvision
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, ...channels.Channel) ([]channels.Channel, []roles.RoleProvision, error)); ok {
		return rf(ctx, session, _a2...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, ...channels.Channel) []channels.Channel); ok {
		r0 = rf(ctx, session, _a2...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]channels.Channel)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, ...channels.Channel) []roles.RoleProvision); ok {
		r1 = rf(ctx, session, _a2...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]roles.RoleProvision)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, authn.Session, ...channels.Channel) error); ok {
		r2 = rf(ctx, session, _a2...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DisableChannel provides a mock function with given fields: ctx, session, id
func (_m *Service) DisableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for DisableChannel")
	}

	var r0 channels.Channel
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (channels.Channel, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) channels.Channel); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(channels.Channel)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Disconnect provides a mock function with given fields: ctx, session, chIDs, clIDs, connType
func (_m *Service) Disconnect(ctx context.Context, session authn.Session, chIDs []string, clIDs []string, connType []connections.ConnType) error {
	ret := _m.Called(ctx, session, chIDs, clIDs, connType)

	if len(ret) == 0 {
		panic("no return value specified for Disconnect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, []string, []string, []connections.ConnType) error); ok {
		r0 = rf(ctx, session, chIDs, clIDs, connType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnableChannel provides a mock function with given fields: ctx, session, id
func (_m *Service) EnableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for EnableChannel")
	}

	var r0 channels.Channel
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (channels.Channel, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) channels.Channel); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(channels.Channel)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAvailableActions provides a mock function with given fields: ctx, session
func (_m *Service) ListAvailableActions(ctx context.Context, session authn.Session) ([]string, error) {
	ret := _m.Called(ctx, session)

	if len(ret) == 0 {
		panic("no return value specified for ListAvailableActions")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) ([]string, error)); ok {
		return rf(ctx, session)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) []string); ok {
		r0 = rf(ctx, session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session) error); ok {
		r1 = rf(ctx, session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListChannels provides a mock function with given fields: ctx, session, pm
func (_m *Service) ListChannels(ctx context.Context, session authn.Session, pm channels.PageMetadata) (channels.Page, error) {
	ret := _m.Called(ctx, session, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListChannels")
	}

	var r0 channels.Page
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, channels.PageMetadata) (channels.Page, error)); ok {
		return rf(ctx, session, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, channels.PageMetadata) channels.Page); ok {
		r0 = rf(ctx, session, pm)
	} else {
		r0 = ret.Get(0).(channels.Page)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, channels.PageMetadata) error); ok {
		r1 = rf(ctx, session, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListEntityMembers provides a mock function with given fields: ctx, session, entityID, pq
func (_m *Service) ListEntityMembers(ctx context.Context, session authn.Session, entityID string, pq roles.MembersRolePageQuery) (roles.MembersRolePage, error) {
	ret := _m.Called(ctx, session, entityID, pq)

	if len(ret) == 0 {
		panic("no return value specified for ListEntityMembers")
	}

	var r0 roles.MembersRolePage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, roles.MembersRolePageQuery) (roles.MembersRolePage, error)); ok {
		return rf(ctx, session, entityID, pq)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, roles.MembersRolePageQuery) roles.MembersRolePage); ok {
		r0 = rf(ctx, session, entityID, pq)
	} else {
		r0 = ret.Get(0).(roles.MembersRolePage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, roles.MembersRolePageQuery) error); ok {
		r1 = rf(ctx, session, entityID, pq)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListUserChannels provides a mock function with given fields: ctx, session, userID, pm
func (_m *Service) ListUserChannels(ctx context.Context, session authn.Session, userID string, pm channels.PageMetadata) (channels.Page, error) {
	ret := _m.Called(ctx, session, userID, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListUserChannels")
	}

	var r0 channels.Page
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, channels.PageMetadata) (channels.Page, error)); ok {
		return rf(ctx, session, userID, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, channels.PageMetadata) channels.Page); ok {
		r0 = rf(ctx, session, userID, pm)
	} else {
		r0 = ret.Get(0).(channels.Page)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, channels.PageMetadata) error); ok {
		r1 = rf(ctx, session, userID, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveChannel provides a mock function with given fields: ctx, session, id
func (_m *Service) RemoveChannel(ctx context.Context, session authn.Session, id string) error {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for RemoveChannel")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveMemberFromAllRoles provides a mock function with given fields: ctx, session, memberID
func (_m *Service) RemoveMemberFromAllRoles(ctx context.Context, session authn.Session, memberID string) error {
	ret := _m.Called(ctx, session, memberID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveMemberFromAllRoles")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, memberID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveMemberFromEntity provides a mock function with given fields: ctx, session, entityID, memberID
func (_m *Service) RemoveMemberFromEntity(ctx context.Context, session authn.Session, entityID string, memberID string) error {
	ret := _m.Called(ctx, session, entityID, memberID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveMemberFromEntity")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, memberID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveParentGroup provides a mock function with given fields: ctx, session, id
func (_m *Service) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for RemoveParentGroup")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveRole provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RemoveRole(ctx context.Context, session authn.Session, entityID string, roleID string) error {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveRole")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RetrieveAllRoles provides a mock function with given fields: ctx, session, entityID, limit, offset
func (_m *Service) RetrieveAllRoles(ctx context.Context, session authn.Session, entityID string, limit uint64, offset uint64) (roles.RolePage, error) {
	ret := _m.Called(ctx, session, entityID, limit, offset)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveAllRoles")
	}

	var r0 roles.RolePage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, uint64, uint64) (roles.RolePage, error)); ok {
		return rf(ctx, session, entityID, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, uint64, uint64) roles.RolePage); ok {
		r0 = rf(ctx, session, entityID, limit, offset)
	} else {
		r0 = ret.Get(0).(roles.RolePage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, uint64, uint64) error); ok {
		r1 = rf(ctx, session, entityID, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveRole provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RetrieveRole(ctx context.Context, session authn.Session, entityID string, roleID string) (roles.Role, error) {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveRole")
	}

	var r0 roles.Role
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) (roles.Role, error)); ok {
		return rf(ctx, session, entityID, roleID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) roles.Role); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Get(0).(roles.Role)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, entityID, roleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleAddActions provides a mock function with given fields: ctx, session, entityID, roleID, actions
func (_m *Service) RoleAddActions(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) ([]string, error) {
	ret := _m.Called(ctx, session, entityID, roleID, actions)

	if len(ret) == 0 {
		panic("no return value specified for RoleAddActions")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) ([]string, error)); ok {
		return rf(ctx, session, entityID, roleID, actions)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) []string); ok {
		r0 = rf(ctx, session, entityID, roleID, actions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleAddMembers provides a mock function with given fields: ctx, session, entityID, roleID, members
func (_m *Service) RoleAddMembers(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) ([]string, error) {
	ret := _m.Called(ctx, session, entityID, roleID, members)

	if len(ret) == 0 {
		panic("no return value specified for RoleAddMembers")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) ([]string, error)); ok {
		return rf(ctx, session, entityID, roleID, members)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) []string); ok {
		r0 = rf(ctx, session, entityID, roleID, members)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, members)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleCheckActionsExists provides a mock function with given fields: ctx, session, entityID, roleID, actions
func (_m *Service) RoleCheckActionsExists(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) (bool, error) {
	ret := _m.Called(ctx, session, entityID, roleID, actions)

	if len(ret) == 0 {
		panic("no return value specified for RoleCheckActionsExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) (bool, error)); ok {
		return rf(ctx, session, entityID, roleID, actions)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) bool); ok {
		r0 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleCheckMembersExists provides a mock function with given fields: ctx, session, entityID, roleID, members
func (_m *Service) RoleCheckMembersExists(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) (bool, error) {
	ret := _m.Called(ctx, session, entityID, roleID, members)

	if len(ret) == 0 {
		panic("no return value specified for RoleCheckMembersExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) (bool, error)); ok {
		return rf(ctx, session, entityID, roleID, members)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) bool); ok {
		r0 = rf(ctx, session, entityID, roleID, members)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, members)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleListActions provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RoleListActions(ctx context.Context, session authn.Session, entityID string, roleID string) ([]string, error) {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RoleListActions")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) ([]string, error)); ok {
		return rf(ctx, session, entityID, roleID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) []string); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, entityID, roleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleListMembers provides a mock function with given fields: ctx, session, entityID, roleID, limit, offset
func (_m *Service) RoleListMembers(ctx context.Context, session authn.Session, entityID string, roleID string, limit uint64, offset uint64) (roles.MembersPage, error) {
	ret := _m.Called(ctx, session, entityID, roleID, limit, offset)

	if len(ret) == 0 {
		panic("no return value specified for RoleListMembers")
	}

	var r0 roles.MembersPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, uint64, uint64) (roles.MembersPage, error)); ok {
		return rf(ctx, session, entityID, roleID, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, uint64, uint64) roles.MembersPage); ok {
		r0 = rf(ctx, session, entityID, roleID, limit, offset)
	} else {
		r0 = ret.Get(0).(roles.MembersPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, uint64, uint64) error); ok {
		r1 = rf(ctx, session, entityID, roleID, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleRemoveActions provides a mock function with given fields: ctx, session, entityID, roleID, actions
func (_m *Service) RoleRemoveActions(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) error {
	ret := _m.Called(ctx, session, entityID, roleID, actions)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveActions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r0 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RoleRemoveAllActions provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RoleRemoveAllActions(ctx context.Context, session authn.Session, entityID string, roleID string) error {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveAllActions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RoleRemoveAllMembers provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RoleRemoveAllMembers(ctx context.Context, session authn.Session, entityID string, roleID string) error {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveAllMembers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RoleRemoveMembers provides a mock function with given fields: ctx, session, entityID, roleID, members
func (_m *Service) RoleRemoveMembers(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) error {
	ret := _m.Called(ctx, session, entityID, roleID, members)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveMembers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r0 = rf(ctx, session, entityID, roleID, members)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetParentGroup provides a mock function with given fields: ctx, session, parentGroupID, id
func (_m *Service) SetParentGroup(ctx context.Context, session authn.Session, parentGroupID string, id string) error {
	ret := _m.Called(ctx, session, parentGroupID, id)

	if len(ret) == 0 {
		panic("no return value specified for SetParentGroup")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, parentGroupID, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateChannel provides a mock function with given fields: ctx, session, channel
func (_m *Service) UpdateChannel(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	ret := _m.Called(ctx, session, channel)

	if len(ret) == 0 {
		panic("no return value specified for UpdateChannel")
	}

	var r0 channels.Channel
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, channels.Channel) (channels.Channel, error)); ok {
		return rf(ctx, session, channel)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, channels.Channel) channels.Channel); ok {
		r0 = rf(ctx, session, channel)
	} else {
		r0 = ret.Get(0).(channels.Channel)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, channels.Channel) error); ok {
		r1 = rf(ctx, session, channel)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateChannelTags provides a mock function with given fields: ctx, session, channel
func (_m *Service) UpdateChannelTags(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	ret := _m.Called(ctx, session, channel)

	if len(ret) == 0 {
		panic("no return value specified for UpdateChannelTags")
	}

	var r0 channels.Channel
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, channels.Channel) (channels.Channel, error)); ok {
		return rf(ctx, session, channel)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, channels.Channel) channels.Channel); ok {
		r0 = rf(ctx, session, channel)
	} else {
		r0 = ret.Get(0).(channels.Channel)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, channels.Channel) error); ok {
		r1 = rf(ctx, session, channel)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateRoleName provides a mock function with given fields: ctx, session, entityID, roleID, newRoleName
func (_m *Service) UpdateRoleName(ctx context.Context, session authn.Session, entityID string, roleID string, newRoleName string) (roles.Role, error) {
	ret := _m.Called(ctx, session, entityID, roleID, newRoleName)

	if len(ret) == 0 {
		panic("no return value specified for UpdateRoleName")
	}

	var r0 roles.Role
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, string) (roles.Role, error)); ok {
		return rf(ctx, session, entityID, roleID, newRoleName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, string) roles.Role); ok {
		r0 = rf(ctx, session, entityID, roleID, newRoleName)
	} else {
		r0 = ret.Get(0).(roles.Role)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, newRoleName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewChannel provides a mock function with given fields: ctx, session, id
func (_m *Service) ViewChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for ViewChannel")
	}

	var r0 channels.Channel
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (channels.Channel, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) channels.Channel); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(channels.Channel)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

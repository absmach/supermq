// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	authn "github.com/absmach/supermq/pkg/authn"

	mock "github.com/stretchr/testify/mock"

	roles "github.com/absmach/supermq/pkg/roles"
)

// RoleManager is an autogenerated mock type for the RoleManager type
type RoleManager struct {
	mock.Mock
}

// AddRole provides a mock function with given fields: ctx, session, entityID, roleName, optionalActions, optionalMembers
func (_m *RoleManager) AddRole(ctx context.Context, session authn.Session, entityID string, roleName string, optionalActions []string, optionalMembers []string) (roles.RoleProvision, error) {
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

// ListAvailableActions provides a mock function with given fields: ctx, session
func (_m *RoleManager) ListAvailableActions(ctx context.Context, session authn.Session) ([]string, error) {
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

// ListEntityMembers provides a mock function with given fields: ctx, session, entityID, pq
func (_m *RoleManager) ListEntityMembers(ctx context.Context, session authn.Session, entityID string, pq roles.MembersRolePageQuery) (roles.MembersRolePage, error) {
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

// RemoveMemberFromAllRoles provides a mock function with given fields: ctx, session, memberID
func (_m *RoleManager) RemoveMemberFromAllRoles(ctx context.Context, session authn.Session, memberID string) error {
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
func (_m *RoleManager) RemoveMemberFromEntity(ctx context.Context, session authn.Session, entityID string, memberID string) error {
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

// RemoveRole provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *RoleManager) RemoveRole(ctx context.Context, session authn.Session, entityID string, roleID string) error {
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
func (_m *RoleManager) RetrieveAllRoles(ctx context.Context, session authn.Session, entityID string, limit uint64, offset uint64) (roles.RolePage, error) {
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
func (_m *RoleManager) RetrieveRole(ctx context.Context, session authn.Session, entityID string, roleID string) (roles.Role, error) {
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
func (_m *RoleManager) RoleAddActions(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) ([]string, error) {
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
func (_m *RoleManager) RoleAddMembers(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) ([]string, error) {
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
func (_m *RoleManager) RoleCheckActionsExists(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) (bool, error) {
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
func (_m *RoleManager) RoleCheckMembersExists(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) (bool, error) {
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
func (_m *RoleManager) RoleListActions(ctx context.Context, session authn.Session, entityID string, roleID string) ([]string, error) {
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
func (_m *RoleManager) RoleListMembers(ctx context.Context, session authn.Session, entityID string, roleID string, limit uint64, offset uint64) (roles.MembersPage, error) {
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
func (_m *RoleManager) RoleRemoveActions(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) error {
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
func (_m *RoleManager) RoleRemoveAllActions(ctx context.Context, session authn.Session, entityID string, roleID string) error {
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
func (_m *RoleManager) RoleRemoveAllMembers(ctx context.Context, session authn.Session, entityID string, roleID string) error {
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
func (_m *RoleManager) RoleRemoveMembers(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) error {
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

// UpdateRoleName provides a mock function with given fields: ctx, session, entityID, roleID, newRoleName
func (_m *RoleManager) UpdateRoleName(ctx context.Context, session authn.Session, entityID string, roleID string, newRoleName string) (roles.Role, error) {
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

// NewRoleManager creates a new instance of RoleManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRoleManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *RoleManager {
	mock := &RoleManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

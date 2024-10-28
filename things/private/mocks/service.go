// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	clients "github.com/absmach/magistrala/pkg/clients"

	mock "github.com/stretchr/testify/mock"

	things "github.com/absmach/magistrala/things"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// AddConnections provides a mock function with given fields: ctx, conns
func (_m *Service) AddConnections(ctx context.Context, conns []things.Connection) error {
	ret := _m.Called(ctx, conns)

	if len(ret) == 0 {
		panic("no return value specified for AddConnections")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []things.Connection) error); ok {
		r0 = rf(ctx, conns)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Authenticate provides a mock function with given fields: ctx, key
func (_m *Service) Authenticate(ctx context.Context, key string) (string, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for Authenticate")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, key)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveChannelConnections provides a mock function with given fields: ctx, channelID
func (_m *Service) RemoveChannelConnections(ctx context.Context, channelID string) error {
	ret := _m.Called(ctx, channelID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveChannelConnections")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, channelID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveConnections provides a mock function with given fields: ctx, conns
func (_m *Service) RemoveConnections(ctx context.Context, conns []things.Connection) error {
	ret := _m.Called(ctx, conns)

	if len(ret) == 0 {
		panic("no return value specified for RemoveConnections")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []things.Connection) error); ok {
		r0 = rf(ctx, conns)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RetrieveById provides a mock function with given fields: ctx, id
func (_m *Service) RetrieveById(ctx context.Context, id string) (clients.Client, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveById")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (clients.Client, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) clients.Client); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveByIds provides a mock function with given fields: ctx, ids
func (_m *Service) RetrieveByIds(ctx context.Context, ids []string) (clients.ClientsPage, error) {
	ret := _m.Called(ctx, ids)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveByIds")
	}

	var r0 clients.ClientsPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) (clients.ClientsPage, error)); ok {
		return rf(ctx, ids)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) clients.ClientsPage); ok {
		r0 = rf(ctx, ids)
	} else {
		r0 = ret.Get(0).(clients.ClientsPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnsetParentGroupFromThings provides a mock function with given fields: ctx, parentGroupID
func (_m *Service) UnsetParentGroupFromThings(ctx context.Context, parentGroupID string) error {
	ret := _m.Called(ctx, parentGroupID)

	if len(ret) == 0 {
		panic("no return value specified for UnsetParentGroupFromThings")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, parentGroupID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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

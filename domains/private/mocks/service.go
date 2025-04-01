// Copyright (c) Abstract Machines

// SPDX-License-Identifier: Apache-2.0

// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	domains "github.com/absmach/supermq/domains"
	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

type Service_Expecter struct {
	mock *mock.Mock
}

func (_m *Service) EXPECT() *Service_Expecter {
	return &Service_Expecter{mock: &_m.Mock}
}

// DeleteUserFromDomains provides a mock function with given fields: ctx, id
func (_m *Service) DeleteUserFromDomains(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUserFromDomains")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Service_DeleteUserFromDomains_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteUserFromDomains'
type Service_DeleteUserFromDomains_Call struct {
	*mock.Call
}

// DeleteUserFromDomains is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *Service_Expecter) DeleteUserFromDomains(ctx interface{}, id interface{}) *Service_DeleteUserFromDomains_Call {
	return &Service_DeleteUserFromDomains_Call{Call: _e.mock.On("DeleteUserFromDomains", ctx, id)}
}

func (_c *Service_DeleteUserFromDomains_Call) Run(run func(ctx context.Context, id string)) *Service_DeleteUserFromDomains_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Service_DeleteUserFromDomains_Call) Return(_a0 error) *Service_DeleteUserFromDomains_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Service_DeleteUserFromDomains_Call) RunAndReturn(run func(context.Context, string) error) *Service_DeleteUserFromDomains_Call {
	_c.Call.Return(run)
	return _c
}

// RetrieveEntity provides a mock function with given fields: ctx, id
func (_m *Service) RetrieveEntity(ctx context.Context, id string) (domains.Domain, error) {
	ret := _m.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveEntity")
	}

	var r0 domains.Domain
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (domains.Domain, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) domains.Domain); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(domains.Domain)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Service_RetrieveEntity_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetrieveEntity'
type Service_RetrieveEntity_Call struct {
	*mock.Call
}

// RetrieveEntity is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *Service_Expecter) RetrieveEntity(ctx interface{}, id interface{}) *Service_RetrieveEntity_Call {
	return &Service_RetrieveEntity_Call{Call: _e.mock.On("RetrieveEntity", ctx, id)}
}

func (_c *Service_RetrieveEntity_Call) Run(run func(ctx context.Context, id string)) *Service_RetrieveEntity_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Service_RetrieveEntity_Call) Return(_a0 domains.Domain, _a1 error) *Service_RetrieveEntity_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Service_RetrieveEntity_Call) RunAndReturn(run func(context.Context, string) (domains.Domain, error)) *Service_RetrieveEntity_Call {
	_c.Call.Return(run)
	return _c
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

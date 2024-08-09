// Copyright (c) Abstract Machines

// SPDX-License-Identifier: Apache-2.0

// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	grpc "google.golang.org/grpc"

	magistrala "github.com/absmach/magistrala"

	mock "github.com/stretchr/testify/mock"
)

// ThingsAuthClient is an autogenerated mock type for the AuthzServiceClient type
type ThingsAuthClient struct {
	mock.Mock
}

type ThingsAuthClient_Expecter struct {
	mock *mock.Mock
}

func (_m *ThingsAuthClient) EXPECT() *ThingsAuthClient_Expecter {
	return &ThingsAuthClient_Expecter{mock: &_m.Mock}
}

// Authorize provides a mock function with given fields: ctx, in, opts
func (_m *ThingsAuthClient) Authorize(ctx context.Context, in *magistrala.AuthorizeReq, opts ...grpc.CallOption) (*magistrala.AuthorizeRes, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Authorize")
	}

	var r0 *magistrala.AuthorizeRes
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *magistrala.AuthorizeReq, ...grpc.CallOption) (*magistrala.AuthorizeRes, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *magistrala.AuthorizeReq, ...grpc.CallOption) *magistrala.AuthorizeRes); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*magistrala.AuthorizeRes)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *magistrala.AuthorizeReq, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ThingsAuthClient_Authorize_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Authorize'
type ThingsAuthClient_Authorize_Call struct {
	*mock.Call
}

// Authorize is a helper method to define mock.On call
//   - ctx context.Context
//   - in *magistrala.AuthorizeReq
//   - opts ...grpc.CallOption
func (_e *ThingsAuthClient_Expecter) Authorize(ctx interface{}, in interface{}, opts ...interface{}) *ThingsAuthClient_Authorize_Call {
	return &ThingsAuthClient_Authorize_Call{Call: _e.mock.On("Authorize",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *ThingsAuthClient_Authorize_Call) Run(run func(ctx context.Context, in *magistrala.AuthorizeReq, opts ...grpc.CallOption)) *ThingsAuthClient_Authorize_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*magistrala.AuthorizeReq), variadicArgs...)
	})
	return _c
}

func (_c *ThingsAuthClient_Authorize_Call) Return(_a0 *magistrala.AuthorizeRes, _a1 error) *ThingsAuthClient_Authorize_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ThingsAuthClient_Authorize_Call) RunAndReturn(run func(context.Context, *magistrala.AuthorizeReq, ...grpc.CallOption) (*magistrala.AuthorizeRes, error)) *ThingsAuthClient_Authorize_Call {
	_c.Call.Return(run)
	return _c
}

// VerifyConnections provides a mock function with given fields: ctx, in, opts
func (_m *ThingsAuthClient) VerifyConnections(ctx context.Context, in *magistrala.VerifyConnectionsReq, opts ...grpc.CallOption) (*magistrala.VerifyConnectionsRes, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for VerifyConnections")
	}

	var r0 *magistrala.VerifyConnectionsRes
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *magistrala.VerifyConnectionsReq, ...grpc.CallOption) (*magistrala.VerifyConnectionsRes, error)); ok {
		return rf(ctx, in, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *magistrala.VerifyConnectionsReq, ...grpc.CallOption) *magistrala.VerifyConnectionsRes); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*magistrala.VerifyConnectionsRes)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *magistrala.VerifyConnectionsReq, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ThingsAuthClient_VerifyConnections_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'VerifyConnections'
type ThingsAuthClient_VerifyConnections_Call struct {
	*mock.Call
}

// VerifyConnections is a helper method to define mock.On call
//   - ctx context.Context
//   - in *magistrala.VerifyConnectionsReq
//   - opts ...grpc.CallOption
func (_e *ThingsAuthClient_Expecter) VerifyConnections(ctx interface{}, in interface{}, opts ...interface{}) *ThingsAuthClient_VerifyConnections_Call {
	return &ThingsAuthClient_VerifyConnections_Call{Call: _e.mock.On("VerifyConnections",
		append([]interface{}{ctx, in}, opts...)...)}
}

func (_c *ThingsAuthClient_VerifyConnections_Call) Run(run func(ctx context.Context, in *magistrala.VerifyConnectionsReq, opts ...grpc.CallOption)) *ThingsAuthClient_VerifyConnections_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]grpc.CallOption, len(args)-2)
		for i, a := range args[2:] {
			if a != nil {
				variadicArgs[i] = a.(grpc.CallOption)
			}
		}
		run(args[0].(context.Context), args[1].(*magistrala.VerifyConnectionsReq), variadicArgs...)
	})
	return _c
}

func (_c *ThingsAuthClient_VerifyConnections_Call) Return(_a0 *magistrala.VerifyConnectionsRes, _a1 error) *ThingsAuthClient_VerifyConnections_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *ThingsAuthClient_VerifyConnections_Call) RunAndReturn(run func(context.Context, *magistrala.VerifyConnectionsReq, ...grpc.CallOption) (*magistrala.VerifyConnectionsRes, error)) *ThingsAuthClient_VerifyConnections_Call {
	_c.Call.Return(run)
	return _c
}

// NewThingsAuthClient creates a new instance of ThingsAuthClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewThingsAuthClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *ThingsAuthClient {
	mock := &ThingsAuthClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

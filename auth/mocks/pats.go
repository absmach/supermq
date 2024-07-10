// Code generated by mockery v2.42.3. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	auth "github.com/absmach/magistrala/auth"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// PATS is an autogenerated mock type for the PATS type
type PATS struct {
	mock.Mock
}

// AddScopeEntry provides a mock function with given fields: ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs
func (_m *PATS) AddScopeEntry(ctx context.Context, token string, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	_va := make([]interface{}, len(entityIDs))
	for _i := range entityIDs {
		_va[_i] = entityIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for AddScopeEntry")
	}

	var r0 auth.Scope
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, auth.PlatformEntityType, string, auth.DomainEntityType, auth.OperationType, ...string) (auth.Scope, error)); ok {
		return rf(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, auth.PlatformEntityType, string, auth.DomainEntityType, auth.OperationType, ...string) auth.Scope); ok {
		r0 = rf(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	} else {
		r0 = ret.Get(0).(auth.Scope)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, auth.PlatformEntityType, string, auth.DomainEntityType, auth.OperationType, ...string) error); ok {
		r1 = rf(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// AuthorizePAT provides a mock function with given fields: ctx, paToken
func (_m *PATS) AuthorizePAT(ctx context.Context, paToken string) (auth.PAT, error) {
	ret := _m.Called(ctx, paToken)

	if len(ret) == 0 {
		panic("no return value specified for AuthorizePAT")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (auth.PAT, error)); ok {
		return rf(ctx, paToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) auth.PAT); ok {
		r0 = rf(ctx, paToken)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, paToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ClearAllScopeEntry provides a mock function with given fields: ctx, token, patID
func (_m *PATS) ClearAllScopeEntry(ctx context.Context, token string, patID string) error {
	ret := _m.Called(ctx, token, patID)

	if len(ret) == 0 {
		panic("no return value specified for ClearAllScopeEntry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, token, patID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: ctx, token, name, description, duration, scope
func (_m *PATS) Create(ctx context.Context, token string, name string, description string, duration time.Duration, scope auth.Scope) (auth.PAT, error) {
	ret := _m.Called(ctx, token, name, description, duration, scope)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, time.Duration, auth.Scope) (auth.PAT, error)); ok {
		return rf(ctx, token, name, description, duration, scope)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, time.Duration, auth.Scope) auth.PAT); ok {
		r0 = rf(ctx, token, name, description, duration, scope)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, time.Duration, auth.Scope) error); ok {
		r1 = rf(ctx, token, name, description, duration, scope)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, token, patID
func (_m *PATS) Delete(ctx context.Context, token string, patID string) error {
	ret := _m.Called(ctx, token, patID)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, token, patID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IdentifyPAT provides a mock function with given fields: ctx, paToken
func (_m *PATS) IdentifyPAT(ctx context.Context, paToken string) (auth.PAT, error) {
	ret := _m.Called(ctx, paToken)

	if len(ret) == 0 {
		panic("no return value specified for IdentifyPAT")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (auth.PAT, error)); ok {
		return rf(ctx, paToken)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) auth.PAT); ok {
		r0 = rf(ctx, paToken)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, paToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, token, pm
func (_m *PATS) List(ctx context.Context, token string, pm auth.PATSPageMeta) (auth.PATSPage, error) {
	ret := _m.Called(ctx, token, pm)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 auth.PATSPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, auth.PATSPageMeta) (auth.PATSPage, error)); ok {
		return rf(ctx, token, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, auth.PATSPageMeta) auth.PATSPage); ok {
		r0 = rf(ctx, token, pm)
	} else {
		r0 = ret.Get(0).(auth.PATSPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, auth.PATSPageMeta) error); ok {
		r1 = rf(ctx, token, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveScopeEntry provides a mock function with given fields: ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs
func (_m *PATS) RemoveScopeEntry(ctx context.Context, token string, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	_va := make([]interface{}, len(entityIDs))
	for _i := range entityIDs {
		_va[_i] = entityIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for RemoveScopeEntry")
	}

	var r0 auth.Scope
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, auth.PlatformEntityType, string, auth.DomainEntityType, auth.OperationType, ...string) (auth.Scope, error)); ok {
		return rf(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, auth.PlatformEntityType, string, auth.DomainEntityType, auth.OperationType, ...string) auth.Scope); ok {
		r0 = rf(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	} else {
		r0 = ret.Get(0).(auth.Scope)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, auth.PlatformEntityType, string, auth.DomainEntityType, auth.OperationType, ...string) error); ok {
		r1 = rf(ctx, token, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResetSecret provides a mock function with given fields: ctx, token, patID, duration
func (_m *PATS) ResetSecret(ctx context.Context, token string, patID string, duration time.Duration) (auth.PAT, error) {
	ret := _m.Called(ctx, token, patID, duration)

	if len(ret) == 0 {
		panic("no return value specified for ResetSecret")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) (auth.PAT, error)); ok {
		return rf(ctx, token, patID, duration)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, time.Duration) auth.PAT); ok {
		r0 = rf(ctx, token, patID, duration)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, time.Duration) error); ok {
		r1 = rf(ctx, token, patID, duration)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Retrieve provides a mock function with given fields: ctx, token, patID
func (_m *PATS) Retrieve(ctx context.Context, token string, patID string) (auth.PAT, error) {
	ret := _m.Called(ctx, token, patID)

	if len(ret) == 0 {
		panic("no return value specified for Retrieve")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (auth.PAT, error)); ok {
		return rf(ctx, token, patID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) auth.PAT); ok {
		r0 = rf(ctx, token, patID)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, token, patID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RevokeSecret provides a mock function with given fields: ctx, token, patID
func (_m *PATS) RevokeSecret(ctx context.Context, token string, patID string) error {
	ret := _m.Called(ctx, token, patID)

	if len(ret) == 0 {
		panic("no return value specified for RevokeSecret")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, token, patID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TestCheckScopeEntry provides a mock function with given fields: ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs
func (_m *PATS) TestCheckScopeEntry(ctx context.Context, paToken string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) error {
	_va := make([]interface{}, len(entityIDs))
	for _i := range entityIDs {
		_va[_i] = entityIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for TestCheckScopeEntry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, auth.PlatformEntityType, string, auth.DomainEntityType, auth.OperationType, ...string) error); ok {
		r0 = rf(ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateDescription provides a mock function with given fields: ctx, token, patID, description
func (_m *PATS) UpdateDescription(ctx context.Context, token string, patID string, description string) (auth.PAT, error) {
	ret := _m.Called(ctx, token, patID, description)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDescription")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (auth.PAT, error)); ok {
		return rf(ctx, token, patID, description)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) auth.PAT); ok {
		r0 = rf(ctx, token, patID, description)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, token, patID, description)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateName provides a mock function with given fields: ctx, token, patID, name
func (_m *PATS) UpdateName(ctx context.Context, token string, patID string, name string) (auth.PAT, error) {
	ret := _m.Called(ctx, token, patID, name)

	if len(ret) == 0 {
		panic("no return value specified for UpdateName")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (auth.PAT, error)); ok {
		return rf(ctx, token, patID, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) auth.PAT); ok {
		r0 = rf(ctx, token, patID, name)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, token, patID, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewPATS creates a new instance of PATS. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPATS(t interface {
	mock.TestingT
	Cleanup(func())
}) *PATS {
	mock := &PATS{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

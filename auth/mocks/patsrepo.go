// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	auth "github.com/absmach/supermq/auth"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// PATSRepository is an autogenerated mock type for the PATSRepository type
type PATSRepository struct {
	mock.Mock
}

// AddScopeEntry provides a mock function with given fields: ctx, userID, scope
func (_m *PATSRepository) AddScopeEntry(ctx context.Context, userID string, scope []auth.Scope) error {
	ret := _m.Called(ctx, userID, scope)

	if len(ret) == 0 {
		panic("no return value specified for AddScopeEntry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []auth.Scope) error); ok {
		r0 = rf(ctx, userID, scope)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CheckScopeEntry provides a mock function with given fields: ctx, userID, patID, entityType, optionalDomainID, operation, entityID
func (_m *PATSRepository) CheckScopeEntry(ctx context.Context, userID string, patID string, entityType auth.EntityType, optionalDomainID string, operation auth.Operation, entityID string) error {
	ret := _m.Called(ctx, userID, patID, entityType, optionalDomainID, operation, entityID)

	if len(ret) == 0 {
		panic("no return value specified for CheckScopeEntry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, auth.EntityType, string, auth.Operation, string) error); ok {
		r0 = rf(ctx, userID, patID, entityType, optionalDomainID, operation, entityID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Reactivate provides a mock function with given fields: ctx, userID, patID
func (_m *PATSRepository) Reactivate(ctx context.Context, userID string, patID string) error {
	ret := _m.Called(ctx, userID, patID)

	if len(ret) == 0 {
		panic("no return value specified for Reactivate")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, userID, patID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Remove provides a mock function with given fields: ctx, userID, patID
func (_m *PATSRepository) Remove(ctx context.Context, userID string, patID string) error {
	ret := _m.Called(ctx, userID, patID)

	if len(ret) == 0 {
		panic("no return value specified for Remove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, userID, patID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveAllScopeEntry provides a mock function with given fields: ctx, patID
func (_m *PATSRepository) RemoveAllScopeEntry(ctx context.Context, patID string) error {
	ret := _m.Called(ctx, patID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveAllScopeEntry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, patID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveScopeEntry provides a mock function with given fields: ctx, userID, scope
func (_m *PATSRepository) RemoveScopeEntry(ctx context.Context, userID string, scope []auth.Scope) error {
	ret := _m.Called(ctx, userID, scope)

	if len(ret) == 0 {
		panic("no return value specified for RemoveScopeEntry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []auth.Scope) error); ok {
		r0 = rf(ctx, userID, scope)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Retrieve provides a mock function with given fields: ctx, userID, patID
func (_m *PATSRepository) Retrieve(ctx context.Context, userID string, patID string) (auth.PAT, error) {
	ret := _m.Called(ctx, userID, patID)

	if len(ret) == 0 {
		panic("no return value specified for Retrieve")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (auth.PAT, error)); ok {
		return rf(ctx, userID, patID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) auth.PAT); ok {
		r0 = rf(ctx, userID, patID)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, userID, patID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveAll provides a mock function with given fields: ctx, userID, pm
func (_m *PATSRepository) RetrieveAll(ctx context.Context, userID string, pm auth.PATSPageMeta) (auth.PATSPage, error) {
	ret := _m.Called(ctx, userID, pm)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveAll")
	}

	var r0 auth.PATSPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, auth.PATSPageMeta) (auth.PATSPage, error)); ok {
		return rf(ctx, userID, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, auth.PATSPageMeta) auth.PATSPage); ok {
		r0 = rf(ctx, userID, pm)
	} else {
		r0 = ret.Get(0).(auth.PATSPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, auth.PATSPageMeta) error); ok {
		r1 = rf(ctx, userID, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveScope provides a mock function with given fields: ctx, pm
func (_m *PATSRepository) RetrieveScope(ctx context.Context, pm auth.ScopesPageMeta) (auth.ScopesPage, error) {
	ret := _m.Called(ctx, pm)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveScope")
	}

	var r0 auth.ScopesPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, auth.ScopesPageMeta) (auth.ScopesPage, error)); ok {
		return rf(ctx, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, auth.ScopesPageMeta) auth.ScopesPage); ok {
		r0 = rf(ctx, pm)
	} else {
		r0 = ret.Get(0).(auth.ScopesPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, auth.ScopesPageMeta) error); ok {
		r1 = rf(ctx, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveSecretAndRevokeStatus provides a mock function with given fields: ctx, userID, patID
func (_m *PATSRepository) RetrieveSecretAndRevokeStatus(ctx context.Context, userID string, patID string) (string, bool, bool, error) {
	ret := _m.Called(ctx, userID, patID)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveSecretAndRevokeStatus")
	}

	var r0 string
	var r1 bool
	var r2 bool
	var r3 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (string, bool, bool, error)); ok {
		return rf(ctx, userID, patID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, userID, patID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) bool); ok {
		r1 = rf(ctx, userID, patID)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, string) bool); ok {
		r2 = rf(ctx, userID, patID)
	} else {
		r2 = ret.Get(2).(bool)
	}

	if rf, ok := ret.Get(3).(func(context.Context, string, string) error); ok {
		r3 = rf(ctx, userID, patID)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// Revoke provides a mock function with given fields: ctx, userID, patID
func (_m *PATSRepository) Revoke(ctx context.Context, userID string, patID string) error {
	ret := _m.Called(ctx, userID, patID)

	if len(ret) == 0 {
		panic("no return value specified for Revoke")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, userID, patID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Save provides a mock function with given fields: ctx, pat
func (_m *PATSRepository) Save(ctx context.Context, pat auth.PAT) error {
	ret := _m.Called(ctx, pat)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, auth.PAT) error); ok {
		r0 = rf(ctx, pat)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateDescription provides a mock function with given fields: ctx, userID, patID, description
func (_m *PATSRepository) UpdateDescription(ctx context.Context, userID string, patID string, description string) (auth.PAT, error) {
	ret := _m.Called(ctx, userID, patID, description)

	if len(ret) == 0 {
		panic("no return value specified for UpdateDescription")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (auth.PAT, error)); ok {
		return rf(ctx, userID, patID, description)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) auth.PAT); ok {
		r0 = rf(ctx, userID, patID, description)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, userID, patID, description)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateName provides a mock function with given fields: ctx, userID, patID, name
func (_m *PATSRepository) UpdateName(ctx context.Context, userID string, patID string, name string) (auth.PAT, error) {
	ret := _m.Called(ctx, userID, patID, name)

	if len(ret) == 0 {
		panic("no return value specified for UpdateName")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (auth.PAT, error)); ok {
		return rf(ctx, userID, patID, name)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) auth.PAT); ok {
		r0 = rf(ctx, userID, patID, name)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, userID, patID, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateTokenHash provides a mock function with given fields: ctx, userID, patID, tokenHash, expiryAt
func (_m *PATSRepository) UpdateTokenHash(ctx context.Context, userID string, patID string, tokenHash string, expiryAt time.Time) (auth.PAT, error) {
	ret := _m.Called(ctx, userID, patID, tokenHash, expiryAt)

	if len(ret) == 0 {
		panic("no return value specified for UpdateTokenHash")
	}

	var r0 auth.PAT
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, time.Time) (auth.PAT, error)); ok {
		return rf(ctx, userID, patID, tokenHash, expiryAt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, time.Time) auth.PAT); ok {
		r0 = rf(ctx, userID, patID, tokenHash, expiryAt)
	} else {
		r0 = ret.Get(0).(auth.PAT)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, time.Time) error); ok {
		r1 = rf(ctx, userID, patID, tokenHash, expiryAt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewPATSRepository creates a new instance of PATSRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPATSRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *PATSRepository {
	mock := &PATSRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	journal "github.com/absmach/supermq/journal"
	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// AddSubscription provides a mock function with given fields: ctx, sub
func (_m *Repository) AddSubscription(ctx context.Context, sub journal.ClientSubscription) error {
	ret := _m.Called(ctx, sub)

	if len(ret) == 0 {
		panic("no return value specified for AddSubscription")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, journal.ClientSubscription) error); ok {
		r0 = rf(ctx, sub)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CountSubscriptions provides a mock function with given fields: ctx, clientID
func (_m *Repository) CountSubscriptions(ctx context.Context, clientID string) (uint64, error) {
	ret := _m.Called(ctx, clientID)

	if len(ret) == 0 {
		panic("no return value specified for CountSubscriptions")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (uint64, error)); ok {
		return rf(ctx, clientID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) uint64); ok {
		r0 = rf(ctx, clientID)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, clientID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteClientTelemetry provides a mock function with given fields: ctx, clientID, domainID
func (_m *Repository) DeleteClientTelemetry(ctx context.Context, clientID string, domainID string) error {
	ret := _m.Called(ctx, clientID, domainID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteClientTelemetry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, clientID, domainID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IncrementInboundMessages provides a mock function with given fields: ctx, clientID
func (_m *Repository) IncrementInboundMessages(ctx context.Context, clientID string) error {
	ret := _m.Called(ctx, clientID)

	if len(ret) == 0 {
		panic("no return value specified for IncrementInboundMessages")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, clientID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IncrementOutboundMessages provides a mock function with given fields: ctx, channelID, subtopic
func (_m *Repository) IncrementOutboundMessages(ctx context.Context, channelID string, subtopic string) error {
	ret := _m.Called(ctx, channelID, subtopic)

	if len(ret) == 0 {
		panic("no return value specified for IncrementOutboundMessages")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, channelID, subtopic)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveSubscription provides a mock function with given fields: ctx, subscriberID
func (_m *Repository) RemoveSubscription(ctx context.Context, subscriberID string) error {
	ret := _m.Called(ctx, subscriberID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveSubscription")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, subscriberID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RetrieveAll provides a mock function with given fields: ctx, page
func (_m *Repository) RetrieveAll(ctx context.Context, page journal.Page) (journal.JournalsPage, error) {
	ret := _m.Called(ctx, page)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveAll")
	}

	var r0 journal.JournalsPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, journal.Page) (journal.JournalsPage, error)); ok {
		return rf(ctx, page)
	}
	if rf, ok := ret.Get(0).(func(context.Context, journal.Page) journal.JournalsPage); ok {
		r0 = rf(ctx, page)
	} else {
		r0 = ret.Get(0).(journal.JournalsPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, journal.Page) error); ok {
		r1 = rf(ctx, page)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveClientTelemetry provides a mock function with given fields: ctx, clientID, domainID
func (_m *Repository) RetrieveClientTelemetry(ctx context.Context, clientID string, domainID string) (journal.ClientTelemetry, error) {
	ret := _m.Called(ctx, clientID, domainID)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveClientTelemetry")
	}

	var r0 journal.ClientTelemetry
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (journal.ClientTelemetry, error)); ok {
		return rf(ctx, clientID, domainID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) journal.ClientTelemetry); ok {
		r0 = rf(ctx, clientID, domainID)
	} else {
		r0 = ret.Get(0).(journal.ClientTelemetry)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, clientID, domainID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Save provides a mock function with given fields: ctx, _a1
func (_m *Repository) Save(ctx context.Context, _a1 journal.Journal) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, journal.Journal) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveClientTelemetry provides a mock function with given fields: ctx, ct
func (_m *Repository) SaveClientTelemetry(ctx context.Context, ct journal.ClientTelemetry) error {
	ret := _m.Called(ctx, ct)

	if len(ret) == 0 {
		panic("no return value specified for SaveClientTelemetry")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, journal.ClientTelemetry) error); ok {
		r0 = rf(ctx, ct)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

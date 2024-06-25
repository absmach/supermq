// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// EventStore is an autogenerated mock type for the EventStore type
type EventStore struct {
	mock.Mock
}

// Connect provides a mock function with given fields: ctx, clientID
func (_m *EventStore) Connect(ctx context.Context, clientID string) error {
	ret := _m.Called(ctx, clientID)

	if len(ret) == 0 {
		panic("no return value specified for Connect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, clientID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Disconnect provides a mock function with given fields: ctx, clientID
func (_m *EventStore) Disconnect(ctx context.Context, clientID string) error {
	ret := _m.Called(ctx, clientID)

	if len(ret) == 0 {
		panic("no return value specified for Disconnect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, clientID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewEventStore creates a new instance of EventStore. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEventStore(t interface {
	mock.TestingT
	Cleanup(func())
}) *EventStore {
	mock := &EventStore{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

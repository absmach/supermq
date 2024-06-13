// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	clients "github.com/absmach/magistrala/pkg/clients"

	magistrala "github.com/absmach/magistrala"

	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// DeleteClient provides a mock function with given fields: ctx, token, id
func (_m *Service) DeleteClient(ctx context.Context, token string, id string) error {
	ret := _m.Called(ctx, token, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteClient")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, token, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DisableClient provides a mock function with given fields: ctx, token, id
func (_m *Service) DisableClient(ctx context.Context, token string, id string) (clients.Client, error) {
	ret := _m.Called(ctx, token, id)

	if len(ret) == 0 {
		panic("no return value specified for DisableClient")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (clients.Client, error)); ok {
		return rf(ctx, token, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) clients.Client); ok {
		r0 = rf(ctx, token, id)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, token, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnableClient provides a mock function with given fields: ctx, token, id
func (_m *Service) EnableClient(ctx context.Context, token string, id string) (clients.Client, error) {
	ret := _m.Called(ctx, token, id)

	if len(ret) == 0 {
		panic("no return value specified for EnableClient")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (clients.Client, error)); ok {
		return rf(ctx, token, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) clients.Client); ok {
		r0 = rf(ctx, token, id)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, token, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenerateResetToken provides a mock function with given fields: ctx, email, host
func (_m *Service) GenerateResetToken(ctx context.Context, email string, host string) error {
	ret := _m.Called(ctx, email, host)

	if len(ret) == 0 {
		panic("no return value specified for GenerateResetToken")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, email, host)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Identify provides a mock function with given fields: ctx, tkn
func (_m *Service) Identify(ctx context.Context, tkn string) (string, error) {
	ret := _m.Called(ctx, tkn)

	if len(ret) == 0 {
		panic("no return value specified for Identify")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, tkn)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, tkn)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tkn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IssueToken provides a mock function with given fields: ctx, identity, secret, domainID
func (_m *Service) IssueToken(ctx context.Context, identity string, secret string, domainID string) (*magistrala.Token, error) {
	ret := _m.Called(ctx, identity, secret, domainID)

	if len(ret) == 0 {
		panic("no return value specified for IssueToken")
	}

	var r0 *magistrala.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*magistrala.Token, error)); ok {
		return rf(ctx, identity, secret, domainID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *magistrala.Token); ok {
		r0 = rf(ctx, identity, secret, domainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*magistrala.Token)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, identity, secret, domainID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListClients provides a mock function with given fields: ctx, token, pm
func (_m *Service) ListClients(ctx context.Context, token string, pm clients.Page) (clients.ClientsPage, error) {
	ret := _m.Called(ctx, token, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListClients")
	}

	var r0 clients.ClientsPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Page) (clients.ClientsPage, error)); ok {
		return rf(ctx, token, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Page) clients.ClientsPage); ok {
		r0 = rf(ctx, token, pm)
	} else {
		r0 = ret.Get(0).(clients.ClientsPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, clients.Page) error); ok {
		r1 = rf(ctx, token, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMembers provides a mock function with given fields: ctx, token, objectKind, objectID, pm
func (_m *Service) ListMembers(ctx context.Context, token string, objectKind string, objectID string, pm clients.Page) (clients.MembersPage, error) {
	ret := _m.Called(ctx, token, objectKind, objectID, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListMembers")
	}

	var r0 clients.MembersPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, clients.Page) (clients.MembersPage, error)); ok {
		return rf(ctx, token, objectKind, objectID, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, clients.Page) clients.MembersPage); ok {
		r0 = rf(ctx, token, objectKind, objectID, pm)
	} else {
		r0 = ret.Get(0).(clients.MembersPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, clients.Page) error); ok {
		r1 = rf(ctx, token, objectKind, objectID, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuthCallback provides a mock function with given fields: ctx, client
func (_m *Service) OAuthCallback(ctx context.Context, client clients.Client) (*magistrala.Token, error) {
	ret := _m.Called(ctx, client)

	if len(ret) == 0 {
		panic("no return value specified for OAuthCallback")
	}

	var r0 *magistrala.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, clients.Client) (*magistrala.Token, error)); ok {
		return rf(ctx, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, clients.Client) *magistrala.Token); ok {
		r0 = rf(ctx, client)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*magistrala.Token)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, clients.Client) error); ok {
		r1 = rf(ctx, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RefreshToken provides a mock function with given fields: ctx, accessToken, domainID
func (_m *Service) RefreshToken(ctx context.Context, accessToken string, domainID string) (*magistrala.Token, error) {
	ret := _m.Called(ctx, accessToken, domainID)

	if len(ret) == 0 {
		panic("no return value specified for RefreshToken")
	}

	var r0 *magistrala.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*magistrala.Token, error)); ok {
		return rf(ctx, accessToken, domainID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *magistrala.Token); ok {
		r0 = rf(ctx, accessToken, domainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*magistrala.Token)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, accessToken, domainID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterClient provides a mock function with given fields: ctx, token, client
func (_m *Service) RegisterClient(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	ret := _m.Called(ctx, token, client)

	if len(ret) == 0 {
		panic("no return value specified for RegisterClient")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) (clients.Client, error)); ok {
		return rf(ctx, token, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) clients.Client); ok {
		r0 = rf(ctx, token, client)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, clients.Client) error); ok {
		r1 = rf(ctx, token, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResetSecret provides a mock function with given fields: ctx, resetToken, secret
func (_m *Service) ResetSecret(ctx context.Context, resetToken string, secret string) error {
	ret := _m.Called(ctx, resetToken, secret)

	if len(ret) == 0 {
		panic("no return value specified for ResetSecret")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, resetToken, secret)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SearchClients provides a mock function with given fields: ctx, token, pm
func (_m *Service) SearchClients(ctx context.Context, token string, pm clients.Page) (clients.ClientsPage, error) {
	ret := _m.Called(ctx, token, pm)

	if len(ret) == 0 {
		panic("no return value specified for SearchClients")
	}

	var r0 clients.ClientsPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Page) (clients.ClientsPage, error)); ok {
		return rf(ctx, token, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Page) clients.ClientsPage); ok {
		r0 = rf(ctx, token, pm)
	} else {
		r0 = ret.Get(0).(clients.ClientsPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, clients.Page) error); ok {
		r1 = rf(ctx, token, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendPasswordReset provides a mock function with given fields: ctx, host, email, user, token
func (_m *Service) SendPasswordReset(ctx context.Context, host string, email string, user string, token string) error {
	ret := _m.Called(ctx, host, email, user, token)

	if len(ret) == 0 {
		panic("no return value specified for SendPasswordReset")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) error); ok {
		r0 = rf(ctx, host, email, user, token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateClient provides a mock function with given fields: ctx, token, client
func (_m *Service) UpdateClient(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	ret := _m.Called(ctx, token, client)

	if len(ret) == 0 {
		panic("no return value specified for UpdateClient")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) (clients.Client, error)); ok {
		return rf(ctx, token, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) clients.Client); ok {
		r0 = rf(ctx, token, client)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, clients.Client) error); ok {
		r1 = rf(ctx, token, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateClientIdentity provides a mock function with given fields: ctx, token, id, identity
func (_m *Service) UpdateClientIdentity(ctx context.Context, token string, id string, identity string) (clients.Client, error) {
	ret := _m.Called(ctx, token, id, identity)

	if len(ret) == 0 {
		panic("no return value specified for UpdateClientIdentity")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (clients.Client, error)); ok {
		return rf(ctx, token, id, identity)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) clients.Client); ok {
		r0 = rf(ctx, token, id, identity)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, token, id, identity)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateClientRole provides a mock function with given fields: ctx, token, client
func (_m *Service) UpdateClientRole(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	ret := _m.Called(ctx, token, client)

	if len(ret) == 0 {
		panic("no return value specified for UpdateClientRole")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) (clients.Client, error)); ok {
		return rf(ctx, token, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) clients.Client); ok {
		r0 = rf(ctx, token, client)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, clients.Client) error); ok {
		r1 = rf(ctx, token, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateClientSecret provides a mock function with given fields: ctx, token, oldSecret, newSecret
func (_m *Service) UpdateClientSecret(ctx context.Context, token string, oldSecret string, newSecret string) (clients.Client, error) {
	ret := _m.Called(ctx, token, oldSecret, newSecret)

	if len(ret) == 0 {
		panic("no return value specified for UpdateClientSecret")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (clients.Client, error)); ok {
		return rf(ctx, token, oldSecret, newSecret)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) clients.Client); ok {
		r0 = rf(ctx, token, oldSecret, newSecret)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, token, oldSecret, newSecret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateClientTags provides a mock function with given fields: ctx, token, client
func (_m *Service) UpdateClientTags(ctx context.Context, token string, client clients.Client) (clients.Client, error) {
	ret := _m.Called(ctx, token, client)

	if len(ret) == 0 {
		panic("no return value specified for UpdateClientTags")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) (clients.Client, error)); ok {
		return rf(ctx, token, client)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, clients.Client) clients.Client); ok {
		r0 = rf(ctx, token, client)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, clients.Client) error); ok {
		r1 = rf(ctx, token, client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewClient provides a mock function with given fields: ctx, token, id
func (_m *Service) ViewClient(ctx context.Context, token string, id string) (clients.Client, error) {
	ret := _m.Called(ctx, token, id)

	if len(ret) == 0 {
		panic("no return value specified for ViewClient")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (clients.Client, error)); ok {
		return rf(ctx, token, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) clients.Client); ok {
		r0 = rf(ctx, token, id)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, token, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewProfile provides a mock function with given fields: ctx, token
func (_m *Service) ViewProfile(ctx context.Context, token string) (clients.Client, error) {
	ret := _m.Called(ctx, token)

	if len(ret) == 0 {
		panic("no return value specified for ViewProfile")
	}

	var r0 clients.Client
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (clients.Client, error)); ok {
		return rf(ctx, token)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) clients.Client); ok {
		r0 = rf(ctx, token)
	} else {
		r0 = ret.Get(0).(clients.Client)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
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

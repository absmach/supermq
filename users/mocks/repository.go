// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify
// Copyright (c) Abstract Machines

// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/absmach/supermq/users"
	mock "github.com/stretchr/testify/mock"
)

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

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

type Repository_Expecter struct {
	mock *mock.Mock
}

func (_m *Repository) EXPECT() *Repository_Expecter {
	return &Repository_Expecter{mock: &_m.Mock}
}

// ChangeStatus provides a mock function for the type Repository
func (_mock *Repository) ChangeStatus(ctx context.Context, user users.User) (users.User, error) {
	ret := _mock.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for ChangeStatus")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) (users.User, error)); ok {
		return returnFunc(ctx, user)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) users.User); ok {
		r0 = returnFunc(ctx, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, users.User) error); ok {
		r1 = returnFunc(ctx, user)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_ChangeStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ChangeStatus'
type Repository_ChangeStatus_Call struct {
	*mock.Call
}

// ChangeStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - user users.User
func (_e *Repository_Expecter) ChangeStatus(ctx interface{}, user interface{}) *Repository_ChangeStatus_Call {
	return &Repository_ChangeStatus_Call{Call: _e.mock.On("ChangeStatus", ctx, user)}
}

func (_c *Repository_ChangeStatus_Call) Run(run func(ctx context.Context, user users.User)) *Repository_ChangeStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 users.User
		if args[1] != nil {
			arg1 = args[1].(users.User)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_ChangeStatus_Call) Return(user1 users.User, err error) *Repository_ChangeStatus_Call {
	_c.Call.Return(user1, err)
	return _c
}

func (_c *Repository_ChangeStatus_Call) RunAndReturn(run func(ctx context.Context, user users.User) (users.User, error)) *Repository_ChangeStatus_Call {
	_c.Call.Return(run)
	return _c
}

// CheckSuperAdmin provides a mock function for the type Repository
func (_mock *Repository) CheckSuperAdmin(ctx context.Context, adminID string) error {
	ret := _mock.Called(ctx, adminID)

	if len(ret) == 0 {
		panic("no return value specified for CheckSuperAdmin")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = returnFunc(ctx, adminID)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// Repository_CheckSuperAdmin_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CheckSuperAdmin'
type Repository_CheckSuperAdmin_Call struct {
	*mock.Call
}

// CheckSuperAdmin is a helper method to define mock.On call
//   - ctx context.Context
//   - adminID string
func (_e *Repository_Expecter) CheckSuperAdmin(ctx interface{}, adminID interface{}) *Repository_CheckSuperAdmin_Call {
	return &Repository_CheckSuperAdmin_Call{Call: _e.mock.On("CheckSuperAdmin", ctx, adminID)}
}

func (_c *Repository_CheckSuperAdmin_Call) Run(run func(ctx context.Context, adminID string)) *Repository_CheckSuperAdmin_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_CheckSuperAdmin_Call) Return(err error) *Repository_CheckSuperAdmin_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Repository_CheckSuperAdmin_Call) RunAndReturn(run func(ctx context.Context, adminID string) error) *Repository_CheckSuperAdmin_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function for the type Repository
func (_mock *Repository) Delete(ctx context.Context, id string) error {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = returnFunc(ctx, id)
	} else {
		r0 = ret.Error(0)
	}
	return r0
}

// Repository_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type Repository_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *Repository_Expecter) Delete(ctx interface{}, id interface{}) *Repository_Delete_Call {
	return &Repository_Delete_Call{Call: _e.mock.On("Delete", ctx, id)}
}

func (_c *Repository_Delete_Call) Run(run func(ctx context.Context, id string)) *Repository_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_Delete_Call) Return(err error) *Repository_Delete_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Repository_Delete_Call) RunAndReturn(run func(ctx context.Context, id string) error) *Repository_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// RetrieveAll provides a mock function for the type Repository
func (_mock *Repository) RetrieveAll(ctx context.Context, pm users.Page) (users.UsersPage, error) {
	ret := _mock.Called(ctx, pm)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveAll")
	}

	var r0 users.UsersPage
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.Page) (users.UsersPage, error)); ok {
		return returnFunc(ctx, pm)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.Page) users.UsersPage); ok {
		r0 = returnFunc(ctx, pm)
	} else {
		r0 = ret.Get(0).(users.UsersPage)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, users.Page) error); ok {
		r1 = returnFunc(ctx, pm)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_RetrieveAll_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetrieveAll'
type Repository_RetrieveAll_Call struct {
	*mock.Call
}

// RetrieveAll is a helper method to define mock.On call
//   - ctx context.Context
//   - pm users.Page
func (_e *Repository_Expecter) RetrieveAll(ctx interface{}, pm interface{}) *Repository_RetrieveAll_Call {
	return &Repository_RetrieveAll_Call{Call: _e.mock.On("RetrieveAll", ctx, pm)}
}

func (_c *Repository_RetrieveAll_Call) Run(run func(ctx context.Context, pm users.Page)) *Repository_RetrieveAll_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 users.Page
		if args[1] != nil {
			arg1 = args[1].(users.Page)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_RetrieveAll_Call) Return(usersPage users.UsersPage, err error) *Repository_RetrieveAll_Call {
	_c.Call.Return(usersPage, err)
	return _c
}

func (_c *Repository_RetrieveAll_Call) RunAndReturn(run func(ctx context.Context, pm users.Page) (users.UsersPage, error)) *Repository_RetrieveAll_Call {
	_c.Call.Return(run)
	return _c
}

// RetrieveAllByIDs provides a mock function for the type Repository
func (_mock *Repository) RetrieveAllByIDs(ctx context.Context, pm users.Page) (users.UsersPage, error) {
	ret := _mock.Called(ctx, pm)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveAllByIDs")
	}

	var r0 users.UsersPage
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.Page) (users.UsersPage, error)); ok {
		return returnFunc(ctx, pm)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.Page) users.UsersPage); ok {
		r0 = returnFunc(ctx, pm)
	} else {
		r0 = ret.Get(0).(users.UsersPage)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, users.Page) error); ok {
		r1 = returnFunc(ctx, pm)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_RetrieveAllByIDs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetrieveAllByIDs'
type Repository_RetrieveAllByIDs_Call struct {
	*mock.Call
}

// RetrieveAllByIDs is a helper method to define mock.On call
//   - ctx context.Context
//   - pm users.Page
func (_e *Repository_Expecter) RetrieveAllByIDs(ctx interface{}, pm interface{}) *Repository_RetrieveAllByIDs_Call {
	return &Repository_RetrieveAllByIDs_Call{Call: _e.mock.On("RetrieveAllByIDs", ctx, pm)}
}

func (_c *Repository_RetrieveAllByIDs_Call) Run(run func(ctx context.Context, pm users.Page)) *Repository_RetrieveAllByIDs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 users.Page
		if args[1] != nil {
			arg1 = args[1].(users.Page)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_RetrieveAllByIDs_Call) Return(usersPage users.UsersPage, err error) *Repository_RetrieveAllByIDs_Call {
	_c.Call.Return(usersPage, err)
	return _c
}

func (_c *Repository_RetrieveAllByIDs_Call) RunAndReturn(run func(ctx context.Context, pm users.Page) (users.UsersPage, error)) *Repository_RetrieveAllByIDs_Call {
	_c.Call.Return(run)
	return _c
}

// RetrieveByEmail provides a mock function for the type Repository
func (_mock *Repository) RetrieveByEmail(ctx context.Context, email string) (users.User, error) {
	ret := _mock.Called(ctx, email)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveByEmail")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (users.User, error)); ok {
		return returnFunc(ctx, email)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) users.User); ok {
		r0 = returnFunc(ctx, email)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, email)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_RetrieveByEmail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetrieveByEmail'
type Repository_RetrieveByEmail_Call struct {
	*mock.Call
}

// RetrieveByEmail is a helper method to define mock.On call
//   - ctx context.Context
//   - email string
func (_e *Repository_Expecter) RetrieveByEmail(ctx interface{}, email interface{}) *Repository_RetrieveByEmail_Call {
	return &Repository_RetrieveByEmail_Call{Call: _e.mock.On("RetrieveByEmail", ctx, email)}
}

func (_c *Repository_RetrieveByEmail_Call) Run(run func(ctx context.Context, email string)) *Repository_RetrieveByEmail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_RetrieveByEmail_Call) Return(user users.User, err error) *Repository_RetrieveByEmail_Call {
	_c.Call.Return(user, err)
	return _c
}

func (_c *Repository_RetrieveByEmail_Call) RunAndReturn(run func(ctx context.Context, email string) (users.User, error)) *Repository_RetrieveByEmail_Call {
	_c.Call.Return(run)
	return _c
}

// RetrieveByID provides a mock function for the type Repository
func (_mock *Repository) RetrieveByID(ctx context.Context, id string) (users.User, error) {
	ret := _mock.Called(ctx, id)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveByID")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (users.User, error)); ok {
		return returnFunc(ctx, id)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) users.User); ok {
		r0 = returnFunc(ctx, id)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, id)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_RetrieveByID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetrieveByID'
type Repository_RetrieveByID_Call struct {
	*mock.Call
}

// RetrieveByID is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
func (_e *Repository_Expecter) RetrieveByID(ctx interface{}, id interface{}) *Repository_RetrieveByID_Call {
	return &Repository_RetrieveByID_Call{Call: _e.mock.On("RetrieveByID", ctx, id)}
}

func (_c *Repository_RetrieveByID_Call) Run(run func(ctx context.Context, id string)) *Repository_RetrieveByID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_RetrieveByID_Call) Return(user users.User, err error) *Repository_RetrieveByID_Call {
	_c.Call.Return(user, err)
	return _c
}

func (_c *Repository_RetrieveByID_Call) RunAndReturn(run func(ctx context.Context, id string) (users.User, error)) *Repository_RetrieveByID_Call {
	_c.Call.Return(run)
	return _c
}

// RetrieveByUsername provides a mock function for the type Repository
func (_mock *Repository) RetrieveByUsername(ctx context.Context, username string) (users.User, error) {
	ret := _mock.Called(ctx, username)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveByUsername")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) (users.User, error)); ok {
		return returnFunc(ctx, username)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string) users.User); ok {
		r0 = returnFunc(ctx, username)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = returnFunc(ctx, username)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_RetrieveByUsername_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RetrieveByUsername'
type Repository_RetrieveByUsername_Call struct {
	*mock.Call
}

// RetrieveByUsername is a helper method to define mock.On call
//   - ctx context.Context
//   - username string
func (_e *Repository_Expecter) RetrieveByUsername(ctx interface{}, username interface{}) *Repository_RetrieveByUsername_Call {
	return &Repository_RetrieveByUsername_Call{Call: _e.mock.On("RetrieveByUsername", ctx, username)}
}

func (_c *Repository_RetrieveByUsername_Call) Run(run func(ctx context.Context, username string)) *Repository_RetrieveByUsername_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_RetrieveByUsername_Call) Return(user users.User, err error) *Repository_RetrieveByUsername_Call {
	_c.Call.Return(user, err)
	return _c
}

func (_c *Repository_RetrieveByUsername_Call) RunAndReturn(run func(ctx context.Context, username string) (users.User, error)) *Repository_RetrieveByUsername_Call {
	_c.Call.Return(run)
	return _c
}

// Save provides a mock function for the type Repository
func (_mock *Repository) Save(ctx context.Context, user users.User) (users.User, error) {
	ret := _mock.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for Save")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) (users.User, error)); ok {
		return returnFunc(ctx, user)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) users.User); ok {
		r0 = returnFunc(ctx, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, users.User) error); ok {
		r1 = returnFunc(ctx, user)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_Save_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Save'
type Repository_Save_Call struct {
	*mock.Call
}

// Save is a helper method to define mock.On call
//   - ctx context.Context
//   - user users.User
func (_e *Repository_Expecter) Save(ctx interface{}, user interface{}) *Repository_Save_Call {
	return &Repository_Save_Call{Call: _e.mock.On("Save", ctx, user)}
}

func (_c *Repository_Save_Call) Run(run func(ctx context.Context, user users.User)) *Repository_Save_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 users.User
		if args[1] != nil {
			arg1 = args[1].(users.User)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_Save_Call) Return(user1 users.User, err error) *Repository_Save_Call {
	_c.Call.Return(user1, err)
	return _c
}

func (_c *Repository_Save_Call) RunAndReturn(run func(ctx context.Context, user users.User) (users.User, error)) *Repository_Save_Call {
	_c.Call.Return(run)
	return _c
}

// SearchUsers provides a mock function for the type Repository
func (_mock *Repository) SearchUsers(ctx context.Context, pm users.Page) (users.UsersPage, error) {
	ret := _mock.Called(ctx, pm)

	if len(ret) == 0 {
		panic("no return value specified for SearchUsers")
	}

	var r0 users.UsersPage
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.Page) (users.UsersPage, error)); ok {
		return returnFunc(ctx, pm)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.Page) users.UsersPage); ok {
		r0 = returnFunc(ctx, pm)
	} else {
		r0 = ret.Get(0).(users.UsersPage)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, users.Page) error); ok {
		r1 = returnFunc(ctx, pm)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_SearchUsers_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SearchUsers'
type Repository_SearchUsers_Call struct {
	*mock.Call
}

// SearchUsers is a helper method to define mock.On call
//   - ctx context.Context
//   - pm users.Page
func (_e *Repository_Expecter) SearchUsers(ctx interface{}, pm interface{}) *Repository_SearchUsers_Call {
	return &Repository_SearchUsers_Call{Call: _e.mock.On("SearchUsers", ctx, pm)}
}

func (_c *Repository_SearchUsers_Call) Run(run func(ctx context.Context, pm users.Page)) *Repository_SearchUsers_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 users.Page
		if args[1] != nil {
			arg1 = args[1].(users.Page)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_SearchUsers_Call) Return(usersPage users.UsersPage, err error) *Repository_SearchUsers_Call {
	_c.Call.Return(usersPage, err)
	return _c
}

func (_c *Repository_SearchUsers_Call) RunAndReturn(run func(ctx context.Context, pm users.Page) (users.UsersPage, error)) *Repository_SearchUsers_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function for the type Repository
func (_mock *Repository) Update(ctx context.Context, id string, user users.UserReq) (users.User, error) {
	ret := _mock.Called(ctx, id, user)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, users.UserReq) (users.User, error)); ok {
		return returnFunc(ctx, id, user)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, users.UserReq) users.User); ok {
		r0 = returnFunc(ctx, id, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, users.UserReq) error); ok {
		r1 = returnFunc(ctx, id, user)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type Repository_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - id string
//   - user users.UserReq
func (_e *Repository_Expecter) Update(ctx interface{}, id interface{}, user interface{}) *Repository_Update_Call {
	return &Repository_Update_Call{Call: _e.mock.On("Update", ctx, id, user)}
}

func (_c *Repository_Update_Call) Run(run func(ctx context.Context, id string, user users.UserReq)) *Repository_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 string
		if args[1] != nil {
			arg1 = args[1].(string)
		}
		var arg2 users.UserReq
		if args[2] != nil {
			arg2 = args[2].(users.UserReq)
		}
		run(
			arg0,
			arg1,
			arg2,
		)
	})
	return _c
}

func (_c *Repository_Update_Call) Return(user1 users.User, err error) *Repository_Update_Call {
	_c.Call.Return(user1, err)
	return _c
}

func (_c *Repository_Update_Call) RunAndReturn(run func(ctx context.Context, id string, user users.UserReq) (users.User, error)) *Repository_Update_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateSecret provides a mock function for the type Repository
func (_mock *Repository) UpdateSecret(ctx context.Context, user users.User) (users.User, error) {
	ret := _mock.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateSecret")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) (users.User, error)); ok {
		return returnFunc(ctx, user)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) users.User); ok {
		r0 = returnFunc(ctx, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, users.User) error); ok {
		r1 = returnFunc(ctx, user)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_UpdateSecret_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateSecret'
type Repository_UpdateSecret_Call struct {
	*mock.Call
}

// UpdateSecret is a helper method to define mock.On call
//   - ctx context.Context
//   - user users.User
func (_e *Repository_Expecter) UpdateSecret(ctx interface{}, user interface{}) *Repository_UpdateSecret_Call {
	return &Repository_UpdateSecret_Call{Call: _e.mock.On("UpdateSecret", ctx, user)}
}

func (_c *Repository_UpdateSecret_Call) Run(run func(ctx context.Context, user users.User)) *Repository_UpdateSecret_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 users.User
		if args[1] != nil {
			arg1 = args[1].(users.User)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_UpdateSecret_Call) Return(user1 users.User, err error) *Repository_UpdateSecret_Call {
	_c.Call.Return(user1, err)
	return _c
}

func (_c *Repository_UpdateSecret_Call) RunAndReturn(run func(ctx context.Context, user users.User) (users.User, error)) *Repository_UpdateSecret_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateUsername provides a mock function for the type Repository
func (_mock *Repository) UpdateUsername(ctx context.Context, user users.User) (users.User, error) {
	ret := _mock.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUsername")
	}

	var r0 users.User
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) (users.User, error)); ok {
		return returnFunc(ctx, user)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, users.User) users.User); ok {
		r0 = returnFunc(ctx, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, users.User) error); ok {
		r1 = returnFunc(ctx, user)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// Repository_UpdateUsername_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateUsername'
type Repository_UpdateUsername_Call struct {
	*mock.Call
}

// UpdateUsername is a helper method to define mock.On call
//   - ctx context.Context
//   - user users.User
func (_e *Repository_Expecter) UpdateUsername(ctx interface{}, user interface{}) *Repository_UpdateUsername_Call {
	return &Repository_UpdateUsername_Call{Call: _e.mock.On("UpdateUsername", ctx, user)}
}

func (_c *Repository_UpdateUsername_Call) Run(run func(ctx context.Context, user users.User)) *Repository_UpdateUsername_Call {
	_c.Call.Run(func(args mock.Arguments) {
		var arg0 context.Context
		if args[0] != nil {
			arg0 = args[0].(context.Context)
		}
		var arg1 users.User
		if args[1] != nil {
			arg1 = args[1].(users.User)
		}
		run(
			arg0,
			arg1,
		)
	})
	return _c
}

func (_c *Repository_UpdateUsername_Call) Return(user1 users.User, err error) *Repository_UpdateUsername_Call {
	_c.Call.Return(user1, err)
	return _c
}

func (_c *Repository_UpdateUsername_Call) RunAndReturn(run func(ctx context.Context, user users.User) (users.User, error)) *Repository_UpdateUsername_Call {
	_c.Call.Return(run)
	return _c
}

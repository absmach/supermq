// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/absmach/magistrala/cli"
	"github.com/absmach/magistrala/internal/testsutil"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	mgsdk "github.com/absmach/magistrala/pkg/sdk/go"
	sdkmocks "github.com/absmach/magistrala/pkg/sdk/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	validToken   = "valid"
	invalidToken = "invalid"
	user         = mgsdk.User{
		Name: "testuser",
		Credentials: mgsdk.Credentials{
			Secret:   "testpassword",
			Identity: "identity@example.com",
		},
		Status: mgclients.EnabledStatus.String(),
	}
)

func TestCreateUsersCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	createCommand := "create"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		user   mgsdk.User
	}{
		{
			desc: "create user successfully with token",
			args: []string{
				createCommand,
				"john doe",
				"john.doe@example.com",
				"12345678",
				validToken,
			},
			user: user,
		},
		{
			desc: "create user successfully without token",
			args: []string{
				createCommand,
				"john doe",
				"john.doe@example.com",
				"12345678",
			},
			user: user,
		},
		{
			desc: "create user with invalid args",
			args: []string{createCommand, user.Name, user.Credentials.Identity},
		},
	}

	for _, tc := range cases {
		var usr mgsdk.User
		sdkCall := sdkMock.On("CreateUser", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &usr, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
		sdkCall.Unset()
	}
}

func TestGetUsersCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	getCommand := "get"
	all := "all"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		user   mgsdk.User
		page   mgsdk.UsersPage
	}{
		{
			desc: "get users successfully",
			args: []string{
				getCommand,
				all,
				validToken,
			},
			sdkerr: nil,
			page: mgsdk.UsersPage{
				Users: []mgsdk.User{user},
			},
		},
		{
			desc: "get user successfully with id",
			args: []string{
				getCommand,
				"id",
				validToken,
			},
			sdkerr: nil,
			user:   user,
		},
		{
			desc: "get users successfully with offset and limit",
			args: []string{
				getCommand,
				all,
				validToken,
				"--offset=2",
				"--limit=5",
			},
			sdkerr: nil,
			page: mgsdk.UsersPage{
				Users: []mgsdk.User{user},
			},
		},
		{
			desc: "get users with invalid token",
			args: []string{
				getCommand,
				all,
				invalidToken,
			},
			sdkerr: errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			error:  fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			page:   mgsdk.UsersPage{},
		},
		{
			desc: "invalid args for get users command",
			args: []string{
				getCommand,
				all,
				invalidToken,
				all,
				invalidToken,
				all,
				invalidToken,
				all,
				invalidToken,
			},
		},
	}

	for _, tc := range cases {
		var page mgsdk.UsersPage
		var usr mgsdk.User
		var errRes, out string
		var err error
		sdkCall := sdkMock.On("Users", mock.Anything, mock.Anything).Return(tc.page, tc.sdkerr)
		sdkCall1 := sdkMock.On("User", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)

		switch {
		case tc.args[1] != all:
			errRes, out, err = executeCommand(rootCmd, &usr, tc.args...)
		default:
			errRes, out, err = executeCommand(rootCmd, &page, tc.args...)
		}

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		switch {
		case tc.args[1] != all:
			assert.Equal(t, tc.user, usr, fmt.Sprintf("%v 4unexpected response, expected: %v, got: %v", tc.desc, tc.user, usr))
		case tc.args[1] == all:
			assert.Equal(t, tc.page, page, fmt.Sprintf("%v 4unexpected response, expected: %v, got: %v", tc.desc, tc.user, page))
		}
		sdkCall.Unset()
		sdkCall1.Unset()
	}
}

func TestIssueTokenCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	tokenCommand := "token"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	token := mgsdk.Token{
		AccessToken:  validToken,
		RefreshToken: validToken,
	}

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		token  mgsdk.Token
	}{
		{
			desc: "issue token successfully without domain id",
			args: []string{
				tokenCommand,
				"john.doe@example.com",
				"12345678",
			},
			sdkerr: nil,
			token:  token,
		},
		{
			desc: "issue token successfully with domain id",
			args: []string{
				tokenCommand,
				"john.doe@example.com",
				"12345678",
				testsutil.GenerateUUID(t),
			},
			sdkerr: nil,
			token:  token,
		},
		{
			desc: " failed to issue token successfully with authentication error",
			args: []string{
				tokenCommand,
				"john.doe@example.com",
				"wrong-password",
			},
			sdkerr: errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			error:  fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			token:  mgsdk.Token{},
		},
		{
			desc: "invalid args for issue token command",
			args: []string{
				tokenCommand,
				"john.doe@example.com",
			},
		},
	}

	for _, tc := range cases {
		var tkn mgsdk.Token
		sdkCall := sdkMock.On("CreateToken", mock.Anything, mock.Anything).Return(tc.token, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &tkn, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.token, tkn, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.token, tkn))
		sdkCall.Unset()
	}
}

func TestRefreshIssueTokenCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	tokenCommand := "refreshtoken"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	token := mgsdk.Token{
		AccessToken:  validToken,
		RefreshToken: validToken,
	}

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		token  mgsdk.Token
	}{
		{
			desc: "issue refresh token successfully without domain id",
			args: []string{
				tokenCommand,
				validToken,
			},
			sdkerr: nil,
			token:  token,
		},
		{
			desc: "issue refresh token successfully with domain id",
			args: []string{
				tokenCommand,
				validToken,
				testsutil.GenerateUUID(t),
			},
			sdkerr: nil,
			token:  token,
		},
		{
			desc: "invalid args for issue refresh token",
			args: []string{
				tokenCommand,
				validToken,
				testsutil.GenerateUUID(t),
				"extra-arg",
			},
		},
		{
			desc: "failed to issue token successfully",
			args: []string{
				tokenCommand,
				invalidToken,
			},
			sdkerr: errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			error:  fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			token:  mgsdk.Token{},
		},
	}

	for _, tc := range cases {
		var tkn mgsdk.Token
		sdkCall := sdkMock.On("RefreshToken", mock.Anything, mock.Anything).Return(tc.token, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &tkn, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.token, tkn, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.token, tkn))
		sdkCall.Unset()
	}
}

func TestUpdateUserCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	updateCommand := "update"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		user   mgsdk.User
	}{
		{
			desc: "update user tags successfully",
			args: []string{
				updateCommand,
				"tags",
				user.ID,
				"[\"tag1\", \"tag2\"]",
				validToken,
			},
			sdkerr: nil,
			user:   user,
		},
		{
			desc: "update user identity successfully",
			args: []string{
				updateCommand,
				"identity",
				user.ID,
				"newidentity@example.com",
				validToken,
			},
			user: user,
		},
		{
			desc: "update user successfully",
			args: []string{
				updateCommand,
				user.ID,
				"{\"name\":\"new name\", \"metadata\":{\"key\": \"value\"}}",
				validToken,
			},
			user: user,
		},
		{
			desc: "update user role successfully",
			args: []string{
				updateCommand,
				"role",
				user.ID,
				"administrator",
				validToken,
			},
			user: user,
		},
		{
			desc: "update user with invalid args",
			args: []string{
				updateCommand,
				"role",
				user.ID,
				"administrator",
				validToken,
				validToken,
			},
		},
	}

	for _, tc := range cases {
		var usr mgsdk.User
		sdkCall := sdkMock.On("UpdateUser", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		sdkCall1 := sdkMock.On("UpdateUserTags", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		sdkCall2 := sdkMock.On("UpdateUserIdentity", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		sdkCall3 := sdkMock.On("UpdateUserRole", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &usr, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))

		sdkCall.Unset()
		sdkCall1.Unset()
		sdkCall2.Unset()
		sdkCall3.Unset()
	}
}

func TestGetUserProfileCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	profileCommand := "profile"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		user   mgsdk.User
	}{
		{
			desc: "get user profile successfully",
			args: []string{
				profileCommand,
				validToken,
			},
			sdkerr: nil,
		},
		{
			desc: "get user profile with invalid args",
			args: []string{
				profileCommand,
				validToken,
				"extra-arg",
			},
		},
	}

	for _, tc := range cases {
		var usr mgsdk.User
		sdkCall := sdkMock.On("UserProfile", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &usr, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
		sdkCall.Unset()
	}
}

func TestResetPasswordRequestCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
	}{
		{
			desc: "request password reset successfully",
			args: []string{
				"resetpasswordrequest",
				"example@mail.com",
			},
			sdkerr: nil,
		},
		{
			desc: "request password reset with invalid args",
			args: []string{
				"resetpasswordrequest",
				"example@mail.com",
				"extra-arg",
			},
		},
	}

	for _, tc := range cases {
		sdkCall := sdkMock.On("ResetPasswordRequest", mock.Anything).Return(tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &mgsdk.User{}, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		sdkCall.Unset()
	}
}

func TestResetPasswordCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	validRequestToken := validToken

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
	}{
		{
			desc: "reset password successfully",
			args: []string{
				"resetpassword",
				"new-password",
				"new-password",
				validRequestToken,
			},
			sdkerr: nil,
		},
		{
			desc: "reset password with invalid args",
			args: []string{
				"resetpassword",
				"new-password",
				"new-password",
				validRequestToken,
				"extra-arg",
			},
		},
	}

	for _, tc := range cases {
		sdkCall := sdkMock.On("ResetPassword", mock.Anything, mock.Anything, mock.Anything).Return(tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &mgsdk.User{}, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		sdkCall.Unset()
	}
}

func TestUpdatePasswordCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		user   mgsdk.User
	}{
		{
			desc: "reset password successfully",
			args: []string{
				"password",
				"old-password",
				"new-password",
				validToken,
			},
			sdkerr: nil,
			user:   user,
		},
		{
			desc: "reset password with invalid args",
			args: []string{
				"password",
				"old-password",
				"new-password",
				validToken,
				validToken,
			},
			sdkerr: nil,
			user:   user,
		},
	}

	for _, tc := range cases {
		sdkCall := sdkMock.On("UpdatePassword", mock.Anything, mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &mgsdk.User{}, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		sdkCall.Unset()
	}
}

func TestEnableUserCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	enableCommand := "enable"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		user   mgsdk.User
	}{
		{
			desc: "enable user successfully",
			args: []string{
				enableCommand,
				user.ID,
				validToken,
			},
			sdkerr: nil,
			user:   user,
		},
		{
			desc: "enable user with invalid args",
			args: []string{
				enableCommand,
				user.ID,
				validToken,
				validToken,
			},
		},
	}

	for _, tc := range cases {
		var usr mgsdk.User
		sdkCall := sdkMock.On("EnableUser", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &usr, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
		sdkCall.Unset()
	}
}

func TestDisableUserCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	disableCommand := "disable"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		user   mgsdk.User
	}{
		{
			desc: "disable user successfully",
			args: []string{
				disableCommand,
				user.ID,
				validToken,
			},
			sdkerr: nil,
			user:   user,
		},
		{
			desc: "disable user with invalid args",
			args: []string{
				disableCommand,
				user.ID,
				validToken,
				validToken,
			},
		},
	}

	for _, tc := range cases {
		var usr mgsdk.User
		sdkCall := sdkMock.On("DisableUser", mock.Anything, mock.Anything).Return(tc.user, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &usr, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
		sdkCall.Unset()
	}
}

func TestListUserChannelsCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	channelsCommand := "channels"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	ch := mgsdk.Channel{
		ID:   testsutil.GenerateUUID(t),
		Name: "testchannel",
	}

	cases := []struct {
		desc    string
		args    []string
		sdkerr  errors.SDKError
		error   string
		channel mgsdk.Channel
		page    mgsdk.ChannelsPage
		output  bool
	}{
		{
			desc: "list user channels successfully",
			args: []string{
				channelsCommand,
				user.ID,
				validToken,
			},
			sdkerr: nil,
			page: mgsdk.ChannelsPage{
				Channels: []mgsdk.Channel{ch},
			},
		},
		{
			desc: "list user channels with invalid args",
			args: []string{
				channelsCommand,
				user.ID,
				validToken,
				validToken,
			},
		},
	}

	for _, tc := range cases {
		var pg mgsdk.ChannelsPage
		sdkCall := sdkMock.On("ListUserChannels", mock.Anything, mock.Anything, mock.Anything).Return(tc.page, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &pg, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.page, pg, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.page, pg))
		sdkCall.Unset()
	}
}

func TestListUserThingsCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	thingsCommand := "things"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	th := mgsdk.Thing{
		ID:   testsutil.GenerateUUID(t),
		Name: "testthing",
	}

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		thing  mgsdk.Thing
		page   mgsdk.ThingsPage
	}{
		{
			desc: "list user things successfully",
			args: []string{
				thingsCommand,
				user.ID,
				validToken,
			},
			sdkerr: nil,
			page: mgsdk.ThingsPage{
				Things: []mgsdk.Thing{th},
			},
		},
		{
			desc: "list user things with invalid args",
			args: []string{
				thingsCommand,
				user.ID,
				validToken,
				validToken,
			},
		},
	}

	for _, tc := range cases {
		var pg mgsdk.ThingsPage
		sdkCall := sdkMock.On("ListUserThings", mock.Anything, mock.Anything, mock.Anything).Return(tc.page, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &pg, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.page, pg, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.page, pg))
		sdkCall.Unset()
	}
}

func TestListUserDomainsCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	domainsCommand := "domains"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	d := mgsdk.Domain{
		ID:   testsutil.GenerateUUID(t),
		Name: "testdomain",
	}

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		page   mgsdk.DomainsPage
	}{
		{
			desc: "list user domains successfully",
			args: []string{
				domainsCommand,
				user.ID,
				validToken,
			},
			sdkerr: nil,
			page: mgsdk.DomainsPage{
				Domains: []mgsdk.Domain{d},
			},
		},
		{
			desc: "list user domains with invalid args",
			args: []string{
				domainsCommand,
				user.ID,
				validToken,
				validToken,
			},
		},
	}

	for _, tc := range cases {
		var pg mgsdk.DomainsPage
		sdkCall := sdkMock.On("ListUserDomains", mock.Anything, mock.Anything, mock.Anything).Return(tc.page, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &pg, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.page, pg, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.page, pg))
		sdkCall.Unset()
	}
}

func TestListUserGroupsCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	domainsCommand := "groups"
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	g := mgsdk.Group{
		ID:   testsutil.GenerateUUID(t),
		Name: "testgroup",
	}

	cases := []struct {
		desc   string
		args   []string
		sdkerr errors.SDKError
		error  string
		page   mgsdk.GroupsPage
	}{
		{
			desc: "list user groups successfully",
			args: []string{
				domainsCommand,
				user.ID,
				validToken,
			},
			sdkerr: nil,
			page: mgsdk.GroupsPage{
				Groups: []mgsdk.Group{g},
			},
		},
		{
			desc: "list user groups with invalid args",
			args: []string{
				domainsCommand,
				user.ID,
				validToken,
			},
		},
	}

	for _, tc := range cases {
		var pg mgsdk.GroupsPage
		sdkCall := sdkMock.On("ListUserGroups", mock.Anything, mock.Anything, mock.Anything).Return(tc.page, tc.sdkerr)
		errRes, out, err := executeCommand(rootCmd, &pg, tc.args...)

		assert.Nil(t, err, fmt.Sprintf("unexpected error when running command: %s, got error: %v, with args: %v", tc.desc, err, tc.args))
		assert.Equal(t, tc.error, errRes, fmt.Sprintf("%s unexpected error response: expected %s got error: %s", tc.desc, tc.error, errRes))
		assert.True(t, !strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
		assert.Equal(t, tc.page, pg, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.page, pg))
		sdkCall.Unset()
	}
}

func executeCommand(root *cobra.Command, v any, args ...string) (errorRes, out string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	o := os.Stdout
	e := os.Stderr

	r, w, _ := os.Pipe()
	r1, w1, _ := os.Pipe()

	os.Stdout = w
	os.Stderr = w1

	_, err = root.ExecuteC()
	if err != nil {
		return "", buf.String(), err
	}

	w.Close()
	w1.Close()
	os.Stdout = o
	os.Stderr = e

	var outBuf, errBuf bytes.Buffer
	_, err = outBuf.ReadFrom(r)
	if err != nil {
		return "", buf.String(), err
	}

	res := outBuf.Bytes()

	_, err = errBuf.ReadFrom(r1)
	if err != nil {
		return "", buf.String(), err
	}

	err = json.Unmarshal(res, v)
	if err != nil && len(res) > 0 {
		return errBuf.String(), outBuf.String(), nil
	}
	return errBuf.String(), buf.String(), nil
}

func setFlags(rootCmd *cobra.Command) *cobra.Command {
	// Root Flags
	rootCmd.PersistentFlags().BoolVarP(
		&cli.RawOutput,
		"raw",
		"r",
		cli.RawOutput,
		"Enables raw output mode for easier parsing of output",
	)

	// Client and Channels Flags
	rootCmd.PersistentFlags().Uint64VarP(
		&cli.Limit,
		"limit",
		"l",
		10,
		"Limit query parameter",
	)

	rootCmd.PersistentFlags().Uint64VarP(
		&cli.Offset,
		"offset",
		"o",
		0,
		"Offset query parameter",
	)

	rootCmd.PersistentFlags().StringVarP(
		&cli.Name,
		"name",
		"n",
		"",
		"Name query parameter",
	)

	rootCmd.PersistentFlags().StringVarP(
		&cli.Identity,
		"identity",
		"I",
		"",
		"User identity query parameter",
	)

	rootCmd.PersistentFlags().StringVarP(
		&cli.Metadata,
		"metadata",
		"m",
		"",
		"Metadata query parameter",
	)

	rootCmd.PersistentFlags().StringVarP(
		&cli.Status,
		"status",
		"S",
		"",
		"User status query parameter",
	)

	rootCmd.PersistentFlags().StringVarP(
		&cli.State,
		"state",
		"z",
		"",
		"Bootstrap state query parameter",
	)

	rootCmd.PersistentFlags().StringVarP(
		&cli.Topic,
		"topic",
		"T",
		"",
		"Subscription topic query parameter",
	)

	rootCmd.PersistentFlags().StringVarP(
		&cli.Contact,
		"contact",
		"C",
		"",
		"Subscription contact query parameter",
	)

	return rootCmd
}

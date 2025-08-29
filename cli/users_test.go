// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/absmach/supermq/cli"
	"github.com/absmach/supermq/internal/testsutil"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	mgsdk "github.com/absmach/supermq/pkg/sdk"
	sdkmocks "github.com/absmach/supermq/pkg/sdk/mocks"
	"github.com/absmach/supermq/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var user = mgsdk.User{
	ID:        testsutil.GenerateUUID(&testing.T{}),
	FirstName: "testuserfirstname",
	LastName:  "testuserfirstname",
	Credentials: mgsdk.Credentials{
		Secret:   "testpassword",
		Username: "testusername",
	},
	Status: users.EnabledStatus.String(),
}

var (
	validToken   = "valid"
	invalidToken = ""
	invalidID    = "invalidID"
	extraArg     = "extra-arg"
)

func TestCreateUsersCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	var usr mgsdk.User

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		user          mgsdk.User
		logType       outputLog
	}{
		{
			desc: "create user successfully with token",
			args: []string{
				createCmd,
				user.FirstName,
				user.LastName,
				user.Email,
				user.Credentials.Username,
				user.Credentials.Secret,
				validToken,
			},
			user:    user,
			logType: entityLog,
		},
		{
			desc: "create user successfully without token",
			args: []string{
				createCmd,
				user.FirstName,
				user.LastName,
				user.Email,
				user.Credentials.Username,
				user.Credentials.Secret,
			},
			user:    user,
			logType: entityLog,
		},
		{
			desc: "failed to create user",
			args: []string{
				createCmd,
				user.FirstName,
				user.LastName,
				user.Email,
				user.Credentials.Username,
				user.Credentials.Secret,
				validToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrCreateEntity, http.StatusUnprocessableEntity),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrCreateEntity, http.StatusUnprocessableEntity).Error()),
			logType:       errLog,
		},
		{
			desc: "create user with invalid args",
			args: []string{
				createCmd,
				user.FirstName,
				user.Credentials.Username,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return(tc.user, tc.sdkErr)
			if len(tc.args) == 6 {
				sdkUser := mgsdk.User{
					FirstName: tc.args[1],
					LastName:  tc.args[2],
					Email:     tc.args[3],
					Credentials: mgsdk.Credentials{
						Username: tc.args[4],
						Secret:   tc.args[5],
					},
					Status: users.EnabledStatus.String(),
				}
				sdkCall = sdkMock.On("CreateUser", mock.Anything, sdkUser, "").Return(tc.user, tc.sdkErr)
			} else if len(tc.args) == 7 {
				sdkUser := mgsdk.User{
					FirstName: tc.args[1],
					LastName:  tc.args[2],
					Email:     tc.args[3],
					Credentials: mgsdk.Credentials{
						Username: tc.args[4],
						Secret:   tc.args[5],
					},
					Status: users.EnabledStatus.String(),
				}
				sdkCall = sdkMock.On("CreateUser", mock.Anything, sdkUser, tc.args[6]).Return(tc.user, tc.sdkErr)
			}
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case entityLog:
				err := json.Unmarshal([]byte(out), &usr)
				assert.Nil(t, err)
				assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}

			sdkCall.Unset()
		})
	}
}

func TestGetUsersCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	var page mgsdk.UsersPage
	var usr mgsdk.User
	out := ""
	userID := testsutil.GenerateUUID(t)

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		user          mgsdk.User
		page          mgsdk.UsersPage
		logType       outputLog
	}{
		{
			desc: "get users successfully",
			args: []string{
				all,
				getCmd,
				validToken,
			},
			sdkErr: nil,
			page: mgsdk.UsersPage{
				Users: []mgsdk.User{user},
			},
			logType: entityLog,
		},
		{
			desc: "get user successfully with id",
			args: []string{
				userID,
				getCmd,
				validToken,
			},
			sdkErr:  nil,
			user:    user,
			logType: entityLog,
		},
		{
			desc: "get user with invalid id",
			args: []string{
				invalidID,
				getCmd,
				validToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrViewEntity, http.StatusBadRequest),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrViewEntity, http.StatusBadRequest).Error()),
			user:          mgsdk.User{},
			logType:       errLog,
		},
		{
			desc: "get users successfully with offset and limit",
			args: []string{
				all,
				getCmd,
				validToken,
				"--offset=2",
				"--limit=5",
			},
			sdkErr: nil,
			page: mgsdk.UsersPage{
				Users: []mgsdk.User{user},
			},
			logType: entityLog,
		},
		{
			desc: "get users with invalid token",
			args: []string{
				all,
				getCmd,
				invalidToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			page:          mgsdk.UsersPage{},
			logType:       errLog,
		},
		{
			desc: "get users with invalid args",
			args: []string{
				all,
				getCmd,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "get user with failed get operation",
			args: []string{
				userID,
				getCmd,
				validToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrViewEntity, http.StatusInternalServerError),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrViewEntity, http.StatusInternalServerError).Error()),
			user:          mgsdk.User{},
			logType:       errLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("Users", mock.Anything, mock.Anything, mock.Anything).Return(tc.page, tc.sdkErr)
			sdkCall1 := sdkMock.On("User", mock.Anything, tc.args[0], tc.args[2]).Return(tc.user, tc.sdkErr)

			out = executeCommand(t, rootCmd, tc.args...)

			if tc.logType == entityLog {
				switch {
				case tc.args[0] == all:
					err := json.Unmarshal([]byte(out), &page)
					if err != nil {
						t.Fatalf("Failed to unmarshal JSON: %v", err)
					}
				default:
					err := json.Unmarshal([]byte(out), &usr)
					if err != nil {
						t.Fatalf("Failed to unmarshal JSON: %v", err)
					}
				}
			}

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}

			if tc.logType == entityLog {
				if tc.args[0] != all {
					assert.Equal(t, tc.user, usr, fmt.Sprintf("%v unexpected response, expected: %v, got: %v", tc.desc, tc.user, usr))
				} else {
					assert.Equal(t, tc.page, page, fmt.Sprintf("%v unexpected response, expected: %v, got: %v", tc.desc, tc.page, page))
				}
			}

			sdkCall.Unset()
			sdkCall1.Unset()
		})
	}
}

func TestIssueTokenCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	var tkn mgsdk.Token
	invalidPassword := ""

	token := mgsdk.Token{
		AccessToken:  testsutil.GenerateUUID(t),
		RefreshToken: testsutil.GenerateUUID(t),
	}

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		token         mgsdk.Token
		logType       outputLog
	}{
		{
			desc: "issue token successfully",
			args: []string{
				tokCmd,
				user.Email,
				user.Credentials.Secret,
			},
			sdkErr:  nil,
			logType: entityLog,
			token:   token,
		},
		{
			desc: "issue token with failed authentication",
			args: []string{
				tokCmd,
				user.Email,
				invalidPassword,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			logType:       errLog,
			token:         mgsdk.Token{},
		},
		{
			desc: "issue token with invalid args",
			args: []string{
				tokCmd,
				user.Email,
				user.Credentials.Secret,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			lg := mgsdk.Login{
				Username: tc.args[1],
				Password: tc.args[2],
			}
			sdkCall := sdkMock.On("CreateToken", mock.Anything, lg).Return(tc.token, tc.sdkErr)

			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case entityLog:
				err := json.Unmarshal([]byte(out), &tkn)
				assert.Nil(t, err)
				assert.Equal(t, tc.token, tkn, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.token, tkn))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}

			sdkCall.Unset()
		})
	}
}

func TestRefreshIssueTokenCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	var tkn mgsdk.Token

	token := mgsdk.Token{
		AccessToken:  testsutil.GenerateUUID(t),
		RefreshToken: testsutil.GenerateUUID(t),
	}

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		token         mgsdk.Token
		logType       outputLog
	}{
		{
			desc: "issue refresh token successfully without domain id",
			args: []string{
				refTokCmd,
				"token",
			},
			sdkErr:  nil,
			logType: entityLog,
			token:   token,
		},
		{
			desc: "issue refresh token with invalid args",
			args: []string{
				refTokCmd,
				"token",
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "issue refresh token with invalid Username",
			args: []string{
				refTokCmd,
				"invalidToken",
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			logType:       errLog,
			token:         mgsdk.Token{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("RefreshToken", mock.Anything, mock.Anything).Return(tc.token, tc.sdkErr)

			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case entityLog:
				err := json.Unmarshal([]byte(out), &tkn)
				assert.Nil(t, err)
				assert.Equal(t, tc.token, tkn, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.token, tkn))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}

			sdkCall.Unset()
		})
	}
}

func TestUpdateUserCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	var usr mgsdk.User

	userID := testsutil.GenerateUUID(t)

	tagUpdateType := "tags"
	emailUpdateType := "email"
	roleUpdateType := "role"
	newEmail := "newemail@example.com"
	newRole := "administrator"
	newTagsJSON := "[\"tag1\", \"tag2\"]"
	newNameMetadataJSON := "{\"name\":\"new name\", \"metadata\":{\"key\": \"value\"}}"

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		user          mgsdk.User
		logType       outputLog
	}{
		{
			desc: "update user tags successfully",
			args: []string{
				userID,
				updateCmd,
				tagUpdateType,
				newTagsJSON,
				validToken,
			},
			sdkErr:  nil,
			logType: entityLog,
			user:    user,
		},
		{
			desc: "update user tags with invalid json",
			args: []string{
				userID,
				updateCmd,
				tagUpdateType,
				"[\"tag1\", \"tag2\"",
				validToken,
			},
			sdkErr:        errors.NewSDKError(errors.New("unexpected end of JSON input")),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.New("unexpected end of JSON input")),
			logType:       errLog,
		},
		{
			desc: "update user tags with invalid token",
			args: []string{
				userID,
				updateCmd,
				tagUpdateType,
				newTagsJSON,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
		{
			desc: "update user email successfully",
			args: []string{
				userID,
				updateCmd,
				emailUpdateType,
				newEmail,
				validToken,
			},
			logType: entityLog,
			user:    user,
		},
		{
			desc: "update user email with invalid token",
			args: []string{
				userID,
				updateCmd,
				emailUpdateType,
				newEmail,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
		{
			desc: "update user successfully",
			args: []string{
				userID,
				updateCmd,
				newNameMetadataJSON,
				validToken,
			},
			logType: entityLog,
			user:    user,
		},
		{
			desc: "update user with invalid token",
			args: []string{
				userID,
				updateCmd,
				newNameMetadataJSON,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
		{
			desc: "update user with invalid json",
			args: []string{
				userID,
				updateCmd,
				"{\"name\":\"new name\", \"metadata\":{\"key\": \"value\"}",
				validToken,
			},
			sdkErr:        errors.NewSDKError(errors.New("unexpected end of JSON input")),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.New("unexpected end of JSON input")),
			logType:       errLog,
		},
		{
			desc: "update user role successfully",
			args: []string{
				userID,
				updateCmd,
				roleUpdateType,
				newRole,
				validToken,
			},
			logType: entityLog,
			user:    user,
		},
		{
			desc: "update user role with invalid token",
			args: []string{
				userID,
				updateCmd,
				roleUpdateType,
				newRole,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
		{
			desc: "update user with invalid args",
			args: []string{
				userID,
				updateCmd,
				roleUpdateType,
				newRole,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("UpdateUser", mock.Anything, mock.Anything, mock.Anything).Return(tc.user, tc.sdkErr)
			sdkCall1 := sdkMock.On("UpdateUserTags", mock.Anything, mock.Anything, mock.Anything).Return(tc.user, tc.sdkErr)
			sdkCall2 := sdkMock.On("UpdateUserIdentity", mock.Anything, mock.Anything, mock.Anything).Return(tc.user, tc.sdkErr)
			sdkCall3 := sdkMock.On("UpdateUserRole", mock.Anything, mock.Anything, mock.Anything).Return(tc.user, tc.sdkErr)
			switch {
			case tc.args[2] == tagUpdateType:
				var u mgsdk.User
				u.Tags = []string{"tag1", "tag2"}
				u.ID = tc.args[0]

				sdkCall1 = sdkMock.On("UpdateUserTags", mock.Anything, u, tc.args[4]).Return(tc.user, tc.sdkErr)
			case tc.args[2] == emailUpdateType:
				var u mgsdk.User
				u.Email = tc.args[3]
				u.ID = tc.args[0]

				sdkCall2 = sdkMock.On("UpdateUserEmail", mock.Anything, u, tc.args[4]).Return(tc.user, tc.sdkErr)
			case tc.args[2] == roleUpdateType && len(tc.args) >= 5:
				sdkCall3 = sdkMock.On("UpdateUserRole", mock.Anything, mgsdk.User{
					Role: tc.args[3],
				}, tc.args[4]).Return(tc.user, tc.sdkErr)
			case len(tc.args) == 4: // Basic user update
				sdkCall = sdkMock.On("UpdateUser", mock.Anything, mgsdk.User{
					FirstName: "new name",
					Metadata: mgsdk.Metadata{
						"key": "value",
					},
				}, tc.args[3]).Return(tc.user, tc.sdkErr)
			}
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case entityLog:
				err := json.Unmarshal([]byte(out), &usr)
				assert.Nil(t, err)
				assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}

			sdkCall.Unset()
			sdkCall1.Unset()
			sdkCall2.Unset()
			sdkCall3.Unset()
		})
	}
}

func TestGetUserProfileCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	var usr mgsdk.User

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		user          mgsdk.User
		logType       outputLog
	}{
		{
			desc: "get user profile successfully",
			args: []string{
				profCmd,
				validToken,
			},
			sdkErr:  nil,
			logType: entityLog,
		},
		{
			desc: "get user profile with invalid args",
			args: []string{
				profCmd,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "get user profile with invalid token",
			args: []string{
				profCmd,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("UserProfile", mock.Anything, tc.args[1]).Return(tc.user, tc.sdkErr)
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case entityLog:
				err := json.Unmarshal([]byte(out), &usr)
				assert.Nil(t, err)
				assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
			}
			sdkCall.Unset()
		})
	}
}

func TestResetPasswordRequestCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	exampleEmail := "example@mail.com"

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		logType       outputLog
	}{
		{
			desc: "request password reset successfully",
			args: []string{
				resPassReqCmd,
				exampleEmail,
			},
			sdkErr:  nil,
			logType: okLog,
		},
		{
			desc: "request password reset with invalid args",
			args: []string{
				resPassReqCmd,
				exampleEmail,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "failed request password reset",
			args: []string{
				resPassReqCmd,
				exampleEmail,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity).Error()),
			logType:       errLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("ResetPasswordRequest", mock.Anything, tc.args[1]).Return(tc.sdkErr)
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case okLog:
				assert.True(t, strings.Contains(out, "ok"), fmt.Sprintf("%s unexpected response: expected success message, got: %v", tc.desc, out))
			}
			sdkCall.Unset()
		})
	}
}

func TestResetPasswordCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	newPassword := "new-password"

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		logType       outputLog
	}{
		{
			desc: "reset password successfully",
			args: []string{
				resPassCmd,
				newPassword,
				newPassword,
				validToken,
			},
			sdkErr:  nil,
			logType: okLog,
		},
		{
			desc: "reset password with invalid args",
			args: []string{
				resPassCmd,
				newPassword,
				newPassword,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "reset password with invalid token",
			args: []string{
				resPassCmd,
				newPassword,
				newPassword,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("ResetPassword", mock.Anything, tc.args[1], tc.args[2], tc.args[3]).Return(tc.sdkErr)
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}

			sdkCall.Unset()
		})
	}
}

func TestUpdatePasswordCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	oldPassword := "old-password"
	newPassword := "new-password"

	var usr mgsdk.User
	var err error

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		user          mgsdk.User
		logType       outputLog
	}{
		{
			desc: "update password successfully",
			args: []string{
				passCmd,
				oldPassword,
				newPassword,
				validToken,
			},
			sdkErr:  nil,
			logType: entityLog,
			user:    user,
		},
		{
			desc: "reset password with invalid args",
			args: []string{
				passCmd,
				oldPassword,
				newPassword,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			sdkErr:        nil,
			logType:       usageLog,
			user:          user,
		},
		{
			desc: "update password with invalid token",
			args: []string{
				passCmd,
				oldPassword,
				newPassword,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("UpdatePassword", mock.Anything, tc.args[1], tc.args[2], tc.args[3]).Return(tc.user, tc.sdkErr)
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case entityLog:
				err = json.Unmarshal([]byte(out), &usr)
				assert.Nil(t, err)
				assert.Equal(t, tc.user, usr, fmt.Sprintf("%s user mismatch: expected %+v got %+v", tc.desc, tc.user, usr))
			}

			sdkCall.Unset()
		})
	}
}

func TestEnableUserCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)
	var usr mgsdk.User

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		user          mgsdk.User
		logType       outputLog
	}{
		{
			desc: "enable user successfully",
			args: []string{
				user.ID,
				enableCmd,
				validToken,
			},
			sdkErr:  nil,
			user:    user,
			logType: entityLog,
		},
		{
			desc: "enable user with invalid args",
			args: []string{
				user.ID,
				enableCmd,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "enable user with invalid token",
			args: []string{
				user.ID,
				enableCmd,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("EnableUser", mock.Anything, tc.args[0], tc.args[2]).Return(tc.user, tc.sdkErr)
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case entityLog:
				err := json.Unmarshal([]byte(out), &usr)
				assert.Nil(t, err)
				assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
			}

			sdkCall.Unset()
		})
	}
}

func TestDisableUserCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	var usr mgsdk.User

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		user          mgsdk.User
		logType       outputLog
	}{
		{
			desc: "disable user successfully",
			args: []string{
				user.ID,
				disableCmd,
				validToken,
			},
			sdkErr:  nil,
			logType: entityLog,
			user:    user,
		},
		{
			desc: "disable user with invalid args",
			args: []string{
				user.ID,
				disableCmd,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "disable user with invalid token",
			args: []string{
				user.ID,
				disableCmd,
				invalidToken,
			},
			logType:       errLog,
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("DisableUser", mock.Anything, tc.args[0], tc.args[2]).Return(tc.user, tc.sdkErr)
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case entityLog:
				err := json.Unmarshal([]byte(out), &usr)
				if err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}
				assert.Equal(t, tc.user, usr, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.user, usr))
			}

			sdkCall.Unset()
		})
	}
}

func TestDeleteUserCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	usersCmd := cli.NewUsersCmd()
	rootCmd := setFlags(usersCmd)

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		logType       outputLog
	}{
		{
			desc: "delete user successfully",
			args: []string{
				user.ID,
				delCmd,
				validToken,
			},
			logType: okLog,
		},
		{
			desc: "delete user with invalid args",
			args: []string{
				user.ID,
				delCmd,
				validToken,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
		{
			desc: "delete user with invalid token",
			args: []string{
				user.ID,
				delCmd,
				invalidToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			logType:       errLog,
		},
		{
			desc: "delete user with invalid user ID",
			args: []string{
				invalidID,
				delCmd,
				validToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden).Error()),
			logType:       errLog,
		},
		{
			desc: "delete user with failed to delete",
			args: []string{
				user.ID,
				delCmd,
				validToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity).Error()),
			logType:       errLog,
		},
		{
			desc: "delete user with invalid args",
			args: []string{
				user.ID,
				delCmd,
				extraArg,
			},
			errLogMessage: rootCmd.Use,
			logType:       usageLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).Return(tc.sdkErr)
			out := executeCommand(t, rootCmd, tc.args...)

			switch tc.logType {
			case okLog:
				assert.True(t, strings.Contains(out, "ok"), fmt.Sprintf("%s unexpected response: expected success message, got: %v", tc.desc, out))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			}

			sdkCall.Unset()
		})
	}
}

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var group = mgsdk.Group{
	ID:   testsutil.GenerateUUID(&testing.T{}),
	Name: "testgroup",
}

func TestCreateGroupCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	groupJson := "{\"name\":\"testgroup\", \"metadata\":{\"key1\":\"value1\"}}"
	groupCmd := cli.NewGroupsCmd()
	rootCmd := setFlags(groupCmd)

	gp := mgsdk.Group{}
	cases := []struct {
		desc          string
		args          []string
		logType       outputLog
		group         mgsdk.Group
		sdkErr        errors.SDKError
		errLogMessage string
	}{
		{
			desc: "create group successfully",
			args: []string{
				groupJson,
				domainID,
				token,
			},
			group:   group,
			logType: entityLog,
		},
		{
			desc: "create group with invalid args",
			args: []string{
				groupJson,
				domainID,
				token,
				extraArg,
			},
			logType: usageLog,
		},
		{
			desc: "create group with invalid json",
			args: []string{
				"{\"name\":\"testgroup\", \"metadata\":{\"key1\":\"value1\"}",
				domainID,
				token,
			},
			sdkErr:        errors.NewSDKError(errors.New("unexpected end of JSON input")),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.New("unexpected end of JSON input")),
			logType:       errLog,
		},
		{
			desc: "create group with invalid token",
			args: []string{
				groupJson,
				domainID,
				invalidToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusUnauthorized),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusUnauthorized)),
			logType:       errLog,
		},
		{
			desc: "create group with invalid domain",
			args: []string{
				groupJson,
				domainID,
				token,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrDomainAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrDomainAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("CreateGroup", mock.Anything, tc.args[1], tc.args[2]).Return(tc.group, tc.sdkErr)
			out := executeCommand(t, rootCmd, append([]string{createCmd}, tc.args...)...)

			switch tc.logType {
			case entityLog:
				err := json.Unmarshal([]byte(out), &gp)
				assert.Nil(t, err)
				assert.Equal(t, tc.group, gp, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.group, gp))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			}
			sdkCall.Unset()
		})
	}
}

func TestDeletegroupCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	groupCmd := cli.NewGroupsCmd()
	rootCmd := setFlags(groupCmd)

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		logType       outputLog
		errLogMessage string
	}{
		{
			desc: "delete group successfully",
			args: []string{
				group.ID,
				domainID,
				token,
			},
			logType: okLog,
		},
		{
			desc: "delete group with invalid args",
			args: []string{
				group.ID,
				domainID,
				token,
				extraArg,
			},
			logType: usageLog,
		},
		{
			desc: "delete group with invalid id",
			args: []string{
				invalidID,
				domainID,
				token,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
		{
			desc: "delete group with invalid token",
			args: []string{
				group.ID,
				domainID,
				invalidToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("DeleteGroup", tc.args[0], tc.args[1], tc.args[2]).Return(tc.sdkErr)
			out := executeCommand(t, rootCmd, append([]string{delCmd}, tc.args...)...)

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

func TestUpdategroupCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	groupCmd := cli.NewGroupsCmd()
	rootCmd := setFlags(groupCmd)

	newGroupJson := fmt.Sprintf("{\"id\":\"%s\",\"name\" : \"newgroup\"}", group.ID)
	cases := []struct {
		desc          string
		args          []string
		group         mgsdk.Group
		sdkErr        errors.SDKError
		errLogMessage string
		logType       outputLog
	}{
		{
			desc: "update group successfully",
			args: []string{
				newGroupJson,
				domainID,
				token,
			},
			group: mgsdk.Group{
				Name: "newgroup1",
				ID:   group.ID,
			},
			logType: entityLog,
		},
		{
			desc: "update group with invalid args",
			args: []string{
				newGroupJson,
				domainID,
				token,
				extraArg,
			},
			logType: usageLog,
		},
		{
			desc: "update group with invalid group id",
			args: []string{
				fmt.Sprintf("{\"id\":\"%s\",\"name\" : \"group1\"}", invalidID),
				domainID,
				token,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
		{
			desc: "update group with invalid json syntax",
			args: []string{
				fmt.Sprintf("{\"id\":\"%s\",\"name\" : \"group1\"", group.ID),
				domainID,
				token,
			},
			sdkErr:        errors.NewSDKError(errors.New("unexpected end of JSON input")),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.New("unexpected end of JSON input")),
			logType:       errLog,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var ch mgsdk.Group
			sdkCall := sdkMock.On("UpdateGroup", mock.Anything, tc.args[1], tc.args[2]).Return(tc.group, tc.sdkErr)
			out := executeCommand(t, rootCmd, append([]string{updCmd}, tc.args...)...)

			switch tc.logType {
			case entityLog:
				err := json.Unmarshal([]byte(out), &ch)
				assert.Nil(t, err)
				assert.Equal(t, tc.group, ch, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.group, ch))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			}
			sdkCall.Unset()
		})
	}
}

func TestEnablegroupCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	groupCmd := cli.NewGroupsCmd()
	rootCmd := setFlags(groupCmd)
	var ch mgsdk.Group

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		group         mgsdk.Group
		logType       outputLog
	}{
		{
			desc: "enable group successfully",
			args: []string{
				group.ID,
				domainID,
				validToken,
			},
			group:   group,
			logType: entityLog,
		},
		{
			desc: "delete group with invalid token",
			args: []string{
				group.ID,
				domainID,
				invalidToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
		{
			desc: "delete group with invalid group ID",
			args: []string{
				invalidID,
				domainID,
				token,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
		{
			desc: "enable group with invalid args",
			args: []string{
				group.ID,
				domainID,
				validToken,
				extraArg,
			},
			logType: usageLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("EnableGroup", tc.args[0], tc.args[1], tc.args[2]).Return(tc.group, tc.sdkErr)
			out := executeCommand(t, rootCmd, append([]string{enableCmd}, tc.args...)...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case entityLog:
				err := json.Unmarshal([]byte(out), &ch)
				assert.Nil(t, err)
				assert.Equal(t, tc.group, ch, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.group, ch))
			}

			sdkCall.Unset()
		})
	}
}

func TestDisablegroupCmd(t *testing.T) {
	sdkMock := new(sdkmocks.SDK)
	cli.SetSDK(sdkMock)
	groupsCmd := cli.NewGroupsCmd()
	rootCmd := setFlags(groupsCmd)

	var ch mgsdk.Group

	cases := []struct {
		desc          string
		args          []string
		sdkErr        errors.SDKError
		errLogMessage string
		group         mgsdk.Group
		logType       outputLog
	}{
		{
			desc: "disable group successfully",
			args: []string{
				group.ID,
				domainID,
				validToken,
			},
			logType: entityLog,
			group:   group,
		},
		{
			desc: "disable group with invalid token",
			args: []string{
				group.ID,
				domainID,
				invalidToken,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
		{
			desc: "disable group with invalid id",
			args: []string{
				invalidID,
				domainID,
				token,
			},
			sdkErr:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
			errLogMessage: fmt.Sprintf("\nerror: %s\n\n", errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden)),
			logType:       errLog,
		},
		{
			desc: "disable client with invalid args",
			args: []string{
				group.ID,
				domainID,
				validToken,
				extraArg,
			},
			logType: usageLog,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdkMock.On("DisableGroup", tc.args[0], tc.args[1], tc.args[2]).Return(tc.group, tc.sdkErr)
			out := executeCommand(t, rootCmd, append([]string{disableCmd}, tc.args...)...)

			switch tc.logType {
			case errLog:
				assert.Equal(t, tc.errLogMessage, out, fmt.Sprintf("%s unexpected error response: expected %s got errLogMessage:%s", tc.desc, tc.errLogMessage, out))
			case usageLog:
				assert.False(t, strings.Contains(out, rootCmd.Use), fmt.Sprintf("%s invalid usage: %s", tc.desc, out))
			case entityLog:
				err := json.Unmarshal([]byte(out), &ch)
				if err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}
				assert.Equal(t, tc.group, ch, fmt.Sprintf("%s unexpected response: expected: %v, got: %v", tc.desc, tc.group, ch))
			}

			sdkCall.Unset()
		})
	}
}

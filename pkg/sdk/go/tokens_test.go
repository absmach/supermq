// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/absmach/magistrala"
	mgauth "github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/pkg/apiutil"
	pauth "github.com/absmach/magistrala/pkg/auth"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	sdk "github.com/absmach/magistrala/pkg/sdk/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestIssueToken(t *testing.T) {
	ts, svc, auth := setupUsers()
	defer ts.Close()

	client := generateTestUser(t)
	token := generateTestToken()

	conf := sdk.Config{
		UsersURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc     string
		login    sdk.Login
		svcRes   mgclients.Client
		svcErr   error
		response sdk.Token
		err      errors.SDKError
	}{
		{
			desc: "issue token successfully",
			login: sdk.Login{
				Identity: client.Credentials.Identity,
				Secret:   client.Credentials.Secret,
				DomainID: validID,
			},
			svcRes: mgclients.Client{
				ID: client.ID,
			},
			svcErr:   nil,
			response: token,
			err:      nil,
		},
		{
			desc: "issue token with invalid identity",
			login: sdk.Login{
				Identity: invalidIdentity,
				Secret:   client.Credentials.Secret,
				DomainID: validID,
			},
			svcRes:   mgclients.Client{},
			svcErr:   svcerr.ErrAuthentication,
			response: sdk.Token{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc: "issue token with invalid secret",
			login: sdk.Login{
				Identity: client.Credentials.Identity,
				Secret:   "invalid",
				DomainID: validID,
			},
			svcRes:   mgclients.Client{},
			svcErr:   svcerr.ErrLogin,
			response: sdk.Token{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrLogin, http.StatusUnauthorized),
		},
		{
			desc: "issue token with empty identity",
			login: sdk.Login{
				Identity: "",
				Secret:   client.Credentials.Secret,
				DomainID: validID,
			},
			svcRes:   mgclients.Client{},
			svcErr:   nil,
			response: sdk.Token{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingIdentity), http.StatusBadRequest),
		},
		{
			desc: "issue token with empty secret",
			login: sdk.Login{
				Identity: client.Credentials.Identity,
				Secret:   "",
				DomainID: validID,
			},
			svcRes:   mgclients.Client{},
			svcErr:   nil,
			response: sdk.Token{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingPass), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authCall := auth.On("Issue", mock.Anything, mock.Anything).Return(&magistrala.Token{AccessToken: token.AccessToken, RefreshToken: &token.RefreshToken, AccessType: mgauth.AccessKey.String()}, tc.err)
			svcCall := svc.On("IssueToken", mock.Anything, tc.login.Identity, tc.login.Secret, tc.login.DomainID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.CreateToken(tc.login)
			assert.Equal(t, tc.err, err)
			fmt.Println("err", err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "IssueToken", mock.Anything, tc.login.Identity, tc.login.Secret, tc.login.DomainID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestRefreshToken(t *testing.T) {
	ts, svc, auth := setupUsers()
	defer ts.Close()

	token := generateTestToken()
	client := generateTestUser(t)

	conf := sdk.Config{
		UsersURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc        string
		token       string
		login       sdk.Login
		svcRes      mgclients.Client
		svcErr      error
		refreshRes  *magistrala.Token
		refreshErr  error
		identifyRes *magistrala.IdentityRes
		identifyErr error
		response    sdk.Token
		err         errors.SDKError
	}{
		{
			desc:  "refresh token successfully",
			token: token.RefreshToken,
			login: sdk.Login{
				DomainID: validID,
			},
			svcRes: mgclients.Client{
				ID: client.ID,
			},
			svcErr: nil,
			refreshRes: &magistrala.Token{
				AccessToken:  token.AccessToken,
				RefreshToken: &token.RefreshToken,
				AccessType:   token.AccessType,
			},
			response: token,
			err:      nil,
		},
		{
			desc:  "refresh token with invalid token",
			token: invalidToken,
			login: sdk.Login{
				DomainID: validID,
			},
			svcRes:     mgclients.Client{},
			svcErr:     svcerr.ErrAuthentication,
			refreshRes: nil,
			response:   sdk.Token{},
			err:        errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:  "refresh token with empty token",
			token: "",
			login: sdk.Login{
				DomainID: validID,
			},
			svcRes:     mgclients.Client{},
			svcErr:     nil,
			refreshRes: nil,
			response:   sdk.Token{},
			err:        errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authCall := auth.On("Identify", mock.Anything, mock.Anything).Return(&magistrala.IdentityRes{Id: validID, UserId: validID, DomainId: validID}, tc.identifyErr)
			authCall1 := auth.On("Refresh", mock.Anything, mock.Anything).Return(tc.refreshRes, tc.refreshErr)
			svcCall := svc.On("RefreshToken", mock.Anything, pauth.Session{DomainUserID: validID, UserID: validID, DomainID: validID}, tc.login.DomainID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.RefreshToken(tc.login, tc.token)
			fmt.Println("err", err)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "RefreshToken", mock.Anything, pauth.Session{DomainUserID: validID, UserID: validID, DomainID: validID}, tc.login.DomainID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
			authCall1.Unset()
		})
	}
}

func generateTestToken() sdk.Token {
	return sdk.Token{
		AccessToken:  "access_token",
		RefreshToken: "refresh_token",
		AccessType:   mgauth.AccessKey.String(),
	}
}

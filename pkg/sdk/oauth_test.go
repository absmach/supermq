// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	authmocks "github.com/absmach/supermq/auth/mocks"
	smqlog "github.com/absmach/supermq/logger"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	authnmocks "github.com/absmach/supermq/pkg/authn/mocks"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/oauth2"
	oauth2mocks "github.com/absmach/supermq/pkg/oauth2/mocks"
	sdk "github.com/absmach/supermq/pkg/sdk"
	"github.com/absmach/supermq/pkg/uuid"
	"github.com/absmach/supermq/users"
	httpapi "github.com/absmach/supermq/users/api"
	umocks "github.com/absmach/supermq/users/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	goauth2 "golang.org/x/oauth2"
)

const (
	googleProvider  = "google"
	jwtRefreshToken = "jwt-refresh-token"
	jwtAccessToken  = "jwt-access-token"
	testAccessToken = "access-token"
	testCode        = "test-code"
	testState       = "test-state"
	testCallbackURL = "http://localhost:9090/callback"
	testAuthURLBase = "https://accounts.google.com/o/oauth2/auth"
	testUserEmail   = "test@example.com"
	testUsername    = "testuser"
)

func setupOAuthServer() (*httptest.Server, *umocks.Service, *oauth2mocks.Provider, *authmocks.TokenServiceClient) {
	usvc := new(umocks.Service)
	logger := smqlog.NewMock()
	mux := chi.NewRouter()
	idp := uuid.NewMock()
	provider := new(oauth2mocks.Provider)
	provider.On("Name").Return(googleProvider)
	authn := new(authnmocks.Authentication)
	am := smqauthn.NewAuthNMiddleware(authn, smqauthn.WithDomainCheck(false), smqauthn.WithAllowUnverifiedUser(true))
	token := new(authmocks.TokenServiceClient)
	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	p := []oauth2.Provider{provider}
	httpapi.MakeHandler(usvc, am, token, true, mux, logger, "", passRegex, idp, redisClient, p, p)

	return httptest.NewServer(mux), usvc, provider, token
}

func TestOAuthAuthorizationURL(t *testing.T) {
	ts, _, provider, _ := setupOAuthServer()
	defer ts.Close()

	conf := sdk.Config{
		UsersURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		providerName    string
		redirectURL     string
		providerEnabled bool
		getAuthURL      string
		state           string
		err             errors.SDKError
	}{
		{
			desc:            "get authorization URL successfully",
			providerName:    googleProvider,
			redirectURL:     "",
			providerEnabled: true,
			getAuthURL:      testAuthURLBase + "?client_id=test&state=" + testState,
			state:           testState,
			err:             nil,
		},
		{
			desc:            "get authorization URL with custom redirect",
			providerName:    googleProvider,
			redirectURL:     testCallbackURL,
			providerEnabled: true,
			getAuthURL:      testAuthURLBase + "?client_id=test&state=" + testState + "&redirect_uri=" + testCallbackURL,
			state:           testState,
			err:             nil,
		},
		{
			desc:            "get authorization URL with disabled provider",
			providerName:    googleProvider,
			redirectURL:     "",
			providerEnabled: false,
			getAuthURL:      "",
			state:           "",
			err:             errors.NewSDKErrorWithStatus(errors.Wrap(svcerr.ErrNotFound, errors.New("oauth provider is disabled")), http.StatusNotFound),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			provider.On("IsEnabled").Return(tc.providerEnabled)
			if tc.providerEnabled {
				if tc.redirectURL != "" {
					provider.On("GetAuthURLWithRedirect", tc.redirectURL).Return(tc.getAuthURL)
				} else {
					provider.On("GetAuthURL").Return(tc.getAuthURL)
				}
				provider.On("State").Return(tc.state)
			}

			authURL, state, err := mgsdk.OAuthAuthorizationURL(context.Background(), tc.providerName, tc.redirectURL)

			if tc.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.getAuthURL, authURL)
				assert.Equal(t, tc.state, state)
			} else {
				assert.Error(t, err)
				assert.Empty(t, authURL)
				assert.Empty(t, state)
			}

			// Reset mocks
			provider.ExpectedCalls = nil
		})
	}
}

func TestOAuthCallback(t *testing.T) {
	ts, svc, provider, tokenClient := setupOAuthServer()
	defer ts.Close()

	conf := sdk.Config{
		UsersURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	validUser := users.User{
		ID:    generateUUID(t),
		Email: testUserEmail,
		Credentials: users.Credentials{
			Username: testUsername,
		},
		Status: users.EnabledStatus,
	}

	cases := []struct {
		desc            string
		providerName    string
		code            string
		state           string
		redirectURL     string
		providerEnabled bool
		mockSetup       func()
		expectedToken   sdk.Token
		err             errors.SDKError
	}{
		{
			desc:            "successful OAuth callback",
			providerName:    googleProvider,
			code:            testCode,
			state:           testState,
			redirectURL:     testCallbackURL,
			providerEnabled: true,
			mockSetup: func() {
				provider.On("IsEnabled").Return(true)
				provider.On("State").Return(testState)
				provider.On("ExchangeWithRedirect", mock.Anything, testCode, testCallbackURL).
					Return(goauth2.Token{AccessToken: testAccessToken}, nil).Once()
				provider.On("UserInfo", testAccessToken).Return(validUser, nil).Once()
				svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
					return u.Email == validUser.Email && u.AuthProvider == googleProvider
				})).Return(validUser, nil).Once()
				svc.On("OAuthAddUserPolicy", mock.Anything, validUser).Return(nil).Once()
				refreshToken := jwtRefreshToken
				tokenClient.On("Issue", mock.Anything, mock.Anything).
					Return(&grpcTokenV1.Token{
						AccessToken:  jwtAccessToken,
						RefreshToken: &refreshToken,
					}, nil).Once()
			},
			expectedToken: sdk.Token{
				AccessToken:  jwtAccessToken,
				RefreshToken: jwtRefreshToken,
			},
			err: nil,
		},
		{
			desc:            "OAuth callback without redirect URL",
			providerName:    googleProvider,
			code:            testCode,
			state:           testState,
			redirectURL:     "",
			providerEnabled: true,
			mockSetup: func() {
				provider.On("IsEnabled").Return(true)
				provider.On("State").Return(testState)
				provider.On("Exchange", mock.Anything, testCode).
					Return(goauth2.Token{AccessToken: testAccessToken}, nil).Once()
				provider.On("UserInfo", testAccessToken).Return(validUser, nil).Once()
				svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
					return u.Email == validUser.Email && u.AuthProvider == googleProvider
				})).Return(validUser, nil).Once()
				svc.On("OAuthAddUserPolicy", mock.Anything, validUser).Return(nil).Once()
				refreshToken := jwtRefreshToken
				tokenClient.On("Issue", mock.Anything, mock.Anything).
					Return(&grpcTokenV1.Token{
						AccessToken:  jwtAccessToken,
						RefreshToken: &refreshToken,
					}, nil).Once()
			},
			expectedToken: sdk.Token{
				AccessToken:  jwtAccessToken,
				RefreshToken: jwtRefreshToken,
			},
			err: nil,
		},
		{
			desc:            "OAuth callback with disabled provider",
			providerName:    googleProvider,
			code:            testCode,
			state:           testState,
			redirectURL:     "",
			providerEnabled: false,
			mockSetup: func() {
				provider.On("IsEnabled").Return(false)
			},
			expectedToken: sdk.Token{},
			err:           errors.NewSDKErrorWithStatus(errors.Wrap(svcerr.ErrNotFound, errors.New("oauth provider is disabled")), http.StatusNotFound),
		},
		{
			desc:            "OAuth callback with invalid state",
			providerName:    googleProvider,
			code:            testCode,
			state:           "wrong-state",
			redirectURL:     "",
			providerEnabled: true,
			mockSetup: func() {
				provider.On("IsEnabled").Return(true)
				provider.On("State").Return(testState)
			},
			expectedToken: sdk.Token{},
			err:           errors.NewSDKErrorWithStatus(errors.Wrap(errors.ErrMalformedEntity, errors.New("invalid state")), http.StatusBadRequest),
		},
		{
			desc:            "OAuth callback with exchange error",
			providerName:    googleProvider,
			code:            testCode,
			state:           testState,
			redirectURL:     "",
			providerEnabled: true,
			mockSetup: func() {
				provider.On("IsEnabled").Return(true)
				provider.On("State").Return(testState)
				provider.On("Exchange", mock.Anything, testCode).
					Return(goauth2.Token{}, fmt.Errorf("exchange failed")).Once()
			},
			expectedToken: sdk.Token{},
			err:           errors.NewSDKErrorWithStatus(errors.New("exchange failed"), http.StatusUnauthorized),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.mockSetup()

			token, err := mgsdk.OAuthCallback(context.Background(), tc.providerName, tc.code, tc.state, tc.redirectURL)

			if tc.err == nil {
				assert.NoError(t, err)
				assert.NotEmpty(t, token.AccessToken)
				assert.NotEmpty(t, token.RefreshToken)
			} else {
				assert.Error(t, err)
				assert.Empty(t, token.AccessToken)
				assert.Empty(t, token.RefreshToken)
			}

			// Reset mocks
			svc.ExpectedCalls = nil
		})
	}
}

func TestOAuthIntegration(t *testing.T) {
	ts, svc, provider, tokenClient := setupOAuthServer()
	defer ts.Close()

	conf := sdk.Config{
		UsersURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	validUser := users.User{
		ID:    generateUUID(t),
		Email: testUserEmail,
		Credentials: users.Credentials{
			Username: testUsername,
		},
		Status: users.EnabledStatus,
	}

	redirectURL := testCallbackURL

	// Setup mocks for authorization URL
	provider.On("IsEnabled").Return(true)
	provider.On("GetAuthURLWithRedirect", redirectURL).
		Return(testAuthURLBase + "?redirect_uri=" + redirectURL)
	provider.On("State").Return(testState)

	// Step 1: Get authorization URL
	authURL, state, err := mgsdk.OAuthAuthorizationURL(context.Background(), googleProvider, redirectURL)
	assert.NoError(t, err)
	assert.NotEmpty(t, authURL)
	assert.Equal(t, testState, state)
	assert.Contains(t, authURL, redirectURL)

	// Setup mocks for callback
	provider.On("ExchangeWithRedirect", mock.Anything, testCode, redirectURL).
		Return(goauth2.Token{AccessToken: testAccessToken}, nil).Once()
	provider.On("UserInfo", testAccessToken).Return(validUser, nil).Once()
	svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
		return u.Email == validUser.Email && u.AuthProvider == googleProvider
	})).Return(validUser, nil).Once()
	svc.On("OAuthAddUserPolicy", mock.Anything, validUser).Return(nil).Once()
	refreshToken := jwtRefreshToken
	tokenClient.On("Issue", mock.Anything, mock.Anything).
		Return(&grpcTokenV1.Token{
			AccessToken:  jwtAccessToken,
			RefreshToken: &refreshToken,
		}, nil).Once()

	// Step 2: Exchange code for token
	token, err := mgsdk.OAuthCallback(context.Background(), googleProvider, testCode, state, redirectURL)
	assert.NoError(t, err)
	assert.NotEmpty(t, token.AccessToken)
	assert.NotEmpty(t, token.RefreshToken)
}

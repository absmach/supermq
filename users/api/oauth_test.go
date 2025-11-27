// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	authmocks "github.com/absmach/supermq/auth/mocks"
	"github.com/absmach/supermq/internal/testsutil"
	smqlog "github.com/absmach/supermq/logger"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	authnmocks "github.com/absmach/supermq/pkg/authn/mocks"
	oauth2mocks "github.com/absmach/supermq/pkg/oauth2/mocks"
	"github.com/absmach/supermq/pkg/uuid"
	"github.com/absmach/supermq/users"
	usersapi "github.com/absmach/supermq/users/api"
	"github.com/absmach/supermq/users/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	goauth2 "golang.org/x/oauth2"
)

func TestOAuthAuthorizeEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	logger := smqlog.NewMock()
	idp := uuid.NewMock()
	authn := new(authnmocks.Authentication)
	am := smqauthn.NewAuthNMiddleware(authn)
	token := new(authmocks.TokenServiceClient)

	cases := []struct {
		name            string
		provider        string
		redirectURI     string
		providerName    string
		providerEnabled bool
		expectedStatus  int
		checkResponse   func(t *testing.T, res *http.Response)
	}{
		{
			name:            "get authorization URL successfully",
			provider:        "google",
			redirectURI:     "",
			providerName:    "google",
			providerEnabled: true,
			expectedStatus:  http.StatusOK,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "authorization_url")
				assert.Contains(t, resp, "state")
				assert.NotEmpty(t, resp["authorization_url"])
				assert.NotEmpty(t, resp["state"])
			},
		},
		{
			name:            "get authorization URL with custom redirect",
			provider:        "google",
			redirectURI:     "http://localhost:9090/callback",
			providerName:    "google",
			providerEnabled: true,
			expectedStatus:  http.StatusOK,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "authorization_url")
				assert.Contains(t, resp["authorization_url"], "redirect_uri")
			},
		},
		{
			name:            "provider disabled",
			provider:        "google",
			redirectURI:     "",
			providerName:    "google",
			providerEnabled: false,
			expectedStatus:  http.StatusNotFound,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
				assert.Equal(t, "oauth provider is disabled", resp["error"])
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(oauth2mocks.Provider)
			provider.On("Name").Return(tc.providerName)
			provider.On("IsEnabled").Return(tc.providerEnabled)
			provider.On("GetAuthURL").Return("https://accounts.google.com/o/oauth2/auth?client_id=test&state=test")
			provider.On("GetAuthURLWithRedirect", mock.Anything).Return("https://accounts.google.com/o/oauth2/auth?client_id=test&state=test&redirect_uri=" + tc.redirectURI)
			provider.On("State").Return("test-state")

			mux := chi.NewRouter()
			redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
			usersapi.MakeHandler(svc, am, token, true, mux, logger, "", passRegex, idp, redisClient, provider)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			url := fmt.Sprintf("%s/oauth/authorize/%s", ts.URL, tc.provider)
			if tc.redirectURI != "" {
				url = fmt.Sprintf("%s?redirect_uri=%s", url, tc.redirectURI)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, res.StatusCode)
			if tc.checkResponse != nil {
				tc.checkResponse(t, res)
			}
		})
	}
}

func TestOAuthCLICallbackEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	logger := smqlog.NewMock()
	idp := uuid.NewMock()
	authn := new(authnmocks.Authentication)
	am := smqauthn.NewAuthNMiddleware(authn)

	validUserID := testsutil.GenerateUUID(t)
	validUser := users.User{
		ID:    validUserID,
		Email: "test@example.com",
		Credentials: users.Credentials{
			Username: "testuser",
		},
		Status: users.EnabledStatus,
	}

	cases := []struct {
		name            string
		provider        string
		providerName    string
		providerEnabled bool
		requestBody     string
		mockSetup       func(*oauth2mocks.Provider, *mocks.Service, *authmocks.TokenServiceClient)
		expectedStatus  int
		checkResponse   func(t *testing.T, res *http.Response)
	}{
		{
			name:            "successful OAuth callback",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state","redirect_url":"http://localhost:9090/callback"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("ExchangeWithRedirect", mock.Anything, "test-code", "http://localhost:9090/callback").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(validUser, nil)
				svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
					return u.Email == validUser.Email && u.AuthProvider == "google"
				})).Return(validUser, nil)
				svc.On("OAuthAddUserPolicy", mock.Anything, validUser).Return(nil)
				refreshToken := "jwt-refresh-token"
				tokenClient.On("Issue", mock.Anything, mock.Anything).
					Return(&grpcTokenV1.Token{
						AccessToken:  "jwt-access-token",
						RefreshToken: &refreshToken,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Equal(t, "jwt-access-token", resp["access_token"])
				assert.Equal(t, "jwt-refresh-token", resp["refresh_token"])
			},
		},
		{
			name:            "OAuth callback without redirect URL",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(validUser, nil)
				svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
					return u.Email == validUser.Email && u.AuthProvider == "google"
				})).Return(validUser, nil)
				svc.On("OAuthAddUserPolicy", mock.Anything, validUser).Return(nil)
				refreshToken := "jwt-refresh-token"
				tokenClient.On("Issue", mock.Anything, mock.Anything).
					Return(&grpcTokenV1.Token{
						AccessToken:  "jwt-access-token",
						RefreshToken: &refreshToken,
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Equal(t, "jwt-access-token", resp["access_token"])
			},
		},
		{
			name:            "provider disabled",
			provider:        "google",
			providerName:    "google",
			providerEnabled: false,
			requestBody:     `{"code":"test-code","state":"test-state"}`,
			mockSetup:       func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {},
			expectedStatus:  http.StatusNotFound,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
				assert.Equal(t, "oauth provider is disabled", resp["error"])
			},
		},
		{
			name:            "invalid request body",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `invalid json`,
			mockSetup:       func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {},
			expectedStatus:  http.StatusBadRequest,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
			},
		},
		{
			name:            "invalid state",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"wrong-state"}`,
			mockSetup:       func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {},
			expectedStatus:  http.StatusBadRequest,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Equal(t, "invalid state", resp["error"])
			},
		},
		{
			name:            "empty code",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"","state":"test-state"}`,
			mockSetup:       func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {},
			expectedStatus:  http.StatusBadRequest,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Equal(t, "empty code", resp["error"])
			},
		},
		{
			name:            "exchange token error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{}, fmt.Errorf("exchange failed"))
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
			},
		},
		{
			name:            "user info retrieval error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(users.User{}, fmt.Errorf("user info failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
			},
		},
		{
			name:            "OAuth callback service error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(validUser, nil)
				svc.On("OAuthCallback", mock.Anything, mock.Anything).
					Return(users.User{}, fmt.Errorf("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
			},
		},
		{
			name:            "add user policy error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(validUser, nil)
				svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
					return u.Email == validUser.Email
				})).Return(validUser, nil)
				svc.On("OAuthAddUserPolicy", mock.Anything, validUser).
					Return(fmt.Errorf("policy error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
			},
		},
		{
			name:            "token issuance error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *mocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(validUser, nil)
				svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
					return u.Email == validUser.Email
				})).Return(validUser, nil)
				svc.On("OAuthAddUserPolicy", mock.Anything, validUser).Return(nil)
				tokenClient.On("Issue", mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("token issuance failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, res *http.Response) {
				body, err := io.ReadAll(res.Body)
				assert.NoError(t, err)

				var resp map[string]string
				err = json.Unmarshal(body, &resp)
				assert.NoError(t, err)
				assert.Contains(t, resp, "error")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(oauth2mocks.Provider)
			provider.On("Name").Return(tc.providerName)
			provider.On("IsEnabled").Return(tc.providerEnabled)
			provider.On("State").Return("test-state")

			tokenClient := new(authmocks.TokenServiceClient)

			tc.mockSetup(provider, svc, tokenClient)

			mux := chi.NewRouter()
			redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
			usersapi.MakeHandler(svc, am, tokenClient, true, mux, logger, "", passRegex, idp, redisClient, provider)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			url := fmt.Sprintf("%s/oauth/cli/callback/%s", ts.URL, tc.provider)
			req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(tc.requestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, res.StatusCode)
			if tc.checkResponse != nil {
				tc.checkResponse(t, res)
			}

			// Reset mocks for next test
			svc.ExpectedCalls = nil
			tokenClient.ExpectedCalls = nil
		})
	}
}

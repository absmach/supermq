// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
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
	"github.com/absmach/supermq/pkg/oauth2"
	oauthhttp "github.com/absmach/supermq/pkg/oauth2/http"
	oauth2mocks "github.com/absmach/supermq/pkg/oauth2/mocks"
	"github.com/absmach/supermq/pkg/oauth2/store"
	"github.com/absmach/supermq/users"
	usermocks "github.com/absmach/supermq/users/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	goauth2 "golang.org/x/oauth2"
)

func TestOAuthAuthorizeEndpoint(t *testing.T) {
	svc := new(usermocks.Service)
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
			makeHandler(svc, token, mux, redisClient, provider)

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
	svc := new(usermocks.Service)

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
		mockSetup       func(*oauth2mocks.Provider, *usermocks.Service, *authmocks.TokenServiceClient)
		expectedStatus  int
		checkResponse   func(t *testing.T, res *http.Response)
	}{
		{
			name:            "successful OAuth callback",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			requestBody:     `{"code":"test-code","state":"test-state","redirect_url":"http://localhost:9090/callback"}`,
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
			},
			expectedStatus: http.StatusNotFound,
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
			},
			expectedStatus: http.StatusBadRequest,
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
			},
			expectedStatus: http.StatusBadRequest,
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
			},
			expectedStatus: http.StatusBadRequest,
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			makeHandler(svc, tokenClient, mux, redisClient, provider)

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

func TestOAuthCallbackEndpoint(t *testing.T) {
	svc := new(usermocks.Service)

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
		queryParams     map[string]string
		mockSetup       func(*oauth2mocks.Provider, *usermocks.Service, *authmocks.TokenServiceClient)
		expectedStatus  int
		checkResponse   func(t *testing.T, res *http.Response)
	}{
		{
			name:            "successful OAuth web callback",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			expectedStatus: http.StatusFound,
			checkResponse: func(t *testing.T, res *http.Response) {
				// Check that it redirects to the redirect URL
				assert.Contains(t, res.Header.Get("Location"), "http://localhost/ui")
				// Check cookies are set
				cookies := res.Cookies()
				assert.NotEmpty(t, cookies)
				foundAccessToken := false
				foundRefreshToken := false
				for _, cookie := range cookies {
					if cookie.Name == "access_token" {
						foundAccessToken = true
						assert.Equal(t, "jwt-access-token", cookie.Value)
					}
					if cookie.Name == "refresh_token" {
						foundRefreshToken = true
						assert.Equal(t, "jwt-refresh-token", cookie.Value)
					}
				}
				assert.True(t, foundAccessToken, "access_token cookie not found")
				assert.True(t, foundRefreshToken, "refresh_token cookie not found")
			},
		},
		{
			name:            "provider disabled",
			provider:        "google",
			providerName:    "google",
			providerEnabled: false,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
			},
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				location := res.Header.Get("Location")
				assert.Contains(t, location, "error=")
				assert.Contains(t, location, "oauth")
			},
		},
		{
			name:            "invalid state",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "wrong-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
			},
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				location := res.Header.Get("Location")
				assert.Contains(t, location, "error=")
				assert.Contains(t, location, "state")
			},
		},
		{
			name:            "empty code",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
			},
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				location := res.Header.Get("Location")
				assert.Contains(t, location, "error=")
				assert.Contains(t, location, "code")
			},
		},
		{
			name:            "exchange token error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{}, fmt.Errorf("exchange failed"))
			},
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				assert.Contains(t, res.Header.Get("Location"), "error=")
			},
		},
		{
			name:            "user info retrieval error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(users.User{}, fmt.Errorf("user info failed"))
			},
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				assert.Contains(t, res.Header.Get("Location"), "error=")
			},
		},
		{
			name:            "OAuth callback service error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(validUser, nil)
				svc.On("OAuthCallback", mock.Anything, mock.Anything).
					Return(users.User{}, fmt.Errorf("service error"))
			},
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				assert.Contains(t, res.Header.Get("Location"), "error=")
			},
		},
		{
			name:            "add user policy error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
				provider.On("Exchange", mock.Anything, "test-code").
					Return(goauth2.Token{AccessToken: "access-token"}, nil)
				provider.On("UserInfo", "access-token").Return(validUser, nil)
				svc.On("OAuthCallback", mock.Anything, mock.MatchedBy(func(u users.User) bool {
					return u.Email == validUser.Email
				})).Return(validUser, nil)
				svc.On("OAuthAddUserPolicy", mock.Anything, validUser).
					Return(fmt.Errorf("policy error"))
			},
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				assert.Contains(t, res.Header.Get("Location"), "error=")
			},
		},
		{
			name:            "token issuance error",
			provider:        "google",
			providerName:    "google",
			providerEnabled: true,
			queryParams: map[string]string{
				"state": "test-state",
				"code":  "test-code",
			},
			mockSetup: func(provider *oauth2mocks.Provider, svc *usermocks.Service, tokenClient *authmocks.TokenServiceClient) {
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
			expectedStatus: http.StatusSeeOther,
			checkResponse: func(t *testing.T, res *http.Response) {
				assert.Contains(t, res.Header.Get("Location"), "error=")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(oauth2mocks.Provider)
			provider.On("Name").Return(tc.providerName)
			provider.On("IsEnabled").Return(tc.providerEnabled)
			provider.On("State").Return("test-state")
			provider.On("ErrorURL").Return("http://localhost/error")
			provider.On("RedirectURL").Return("http://localhost/ui")

			tokenClient := new(authmocks.TokenServiceClient)

			tc.mockSetup(provider, svc, tokenClient)

			mux := chi.NewRouter()
			redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
			makeHandler(svc, tokenClient, mux, redisClient, provider)

			ts := httptest.NewServer(mux)
			defer ts.Close()

			// Build query string
			url := fmt.Sprintf("%s/oauth/callback/%s", ts.URL, tc.provider)
			if len(tc.queryParams) > 0 {
				url += "?"
				first := true
				for k, v := range tc.queryParams {
					if !first {
						url += "&"
					}
					url += fmt.Sprintf("%s=%s", k, v)
					first = false
				}
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			assert.NoError(t, err)

			// Don't follow redirects to check the redirect URL
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			res, err := client.Do(req)
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

func makeHandler(svc users.Service, tokensvc grpcTokenV1.TokenServiceClient, mux *chi.Mux, cacheClient *redis.Client, providers ...oauth2.Provider) http.Handler {
	ctx := context.Background()

	deviceStore := store.NewRedisDeviceCodeStore(ctx, cacheClient)
	oauthSvc := oauth2.NewOAuthService(deviceStore, svc, tokensvc)

	mux = oauthhttp.Handler(mux, tokensvc, oauthSvc, providers...)
	mux = oauthhttp.DeviceHandler(mux, tokensvc, oauthSvc, providers...)

	return mux
}

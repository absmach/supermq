// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/absmach/supermq/pkg/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthDeviceCode(t *testing.T) {
	tests := []struct {
		name           string
		providerName   string
		serverResponse string
		serverStatus   int
		expectedErr    bool
		checkResponse  func(*testing.T, sdk.DeviceCode)
	}{
		{
			name:         "successful device code request",
			providerName: "google",
			serverResponse: `{
				"device_code": "device123abc",
				"user_code": "ABCD-EFGH",
				"verification_uri": "https://example.com/device",
				"expires_in": 600,
				"interval": 3
			}`,
			serverStatus: http.StatusOK,
			expectedErr:  false,
			checkResponse: func(t *testing.T, deviceCode sdk.DeviceCode) {
				assert.Equal(t, "device123abc", deviceCode.DeviceCode)
				assert.Equal(t, "ABCD-EFGH", deviceCode.UserCode)
				assert.Equal(t, "https://example.com/device", deviceCode.VerificationURI)
				assert.Equal(t, 600, deviceCode.ExpiresIn)
				assert.Equal(t, 3, deviceCode.Interval)
			},
		},
		{
			name:           "provider not found",
			providerName:   "unknown",
			serverResponse: `{"error": "oauth provider is disabled"}`,
			serverStatus:   http.StatusNotFound,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, deviceCode sdk.DeviceCode) {},
		},
		{
			name:           "invalid json response",
			providerName:   "google",
			serverResponse: `{invalid json}`,
			serverStatus:   http.StatusOK,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, deviceCode sdk.DeviceCode) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, fmt.Sprintf("/oauth/device/code/%s", tc.providerName), r.URL.Path)

				w.WriteHeader(tc.serverStatus)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			sdkConf := sdk.Config{
				UsersURL: server.URL,
			}
			mgsdk := sdk.NewSDK(sdkConf)

			deviceCode, err := mgsdk.OAuthDeviceCode(context.Background(), tc.providerName)

			if tc.expectedErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			tc.checkResponse(t, deviceCode)
		})
	}
}

func TestOAuthDeviceToken(t *testing.T) {
	tests := []struct {
		name           string
		providerName   string
		deviceCode     string
		serverResponse string
		serverStatus   int
		expectedErr    bool
		checkResponse  func(*testing.T, sdk.Token)
	}{
		{
			name:         "successful token retrieval",
			providerName: "google",
			deviceCode:   "device123",
			serverResponse: `{
				"access_token": "access_token_123",
				"refresh_token": "refresh_token_456"
			}`,
			serverStatus: http.StatusOK,
			expectedErr:  false,
			checkResponse: func(t *testing.T, token sdk.Token) {
				assert.Equal(t, "access_token_123", token.AccessToken)
				assert.Equal(t, "refresh_token_456", token.RefreshToken)
			},
		},
		{
			name:           "authorization pending",
			providerName:   "google",
			deviceCode:     "device123",
			serverResponse: `{"error": "authorization pending"}`,
			serverStatus:   http.StatusAccepted,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, token sdk.Token) {},
		},
		{
			name:           "device code expired",
			providerName:   "google",
			deviceCode:     "device123",
			serverResponse: `{"error": "device code expired"}`,
			serverStatus:   http.StatusBadRequest,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, token sdk.Token) {},
		},
		{
			name:           "slow down",
			providerName:   "google",
			deviceCode:     "device123",
			serverResponse: `{"error": "slow down"}`,
			serverStatus:   http.StatusBadRequest,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, token sdk.Token) {},
		},
		{
			name:           "access denied",
			providerName:   "google",
			deviceCode:     "device123",
			serverResponse: `{"error": "access denied"}`,
			serverStatus:   http.StatusUnauthorized,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, token sdk.Token) {},
		},
		{
			name:           "invalid device code",
			providerName:   "google",
			deviceCode:     "invalid",
			serverResponse: `{"error": "invalid device code"}`,
			serverStatus:   http.StatusNotFound,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, token sdk.Token) {},
		},
		{
			name:           "provider disabled",
			providerName:   "disabled",
			deviceCode:     "device123",
			serverResponse: `{"error": "oauth provider is disabled"}`,
			serverStatus:   http.StatusNotFound,
			expectedErr:    true,
			checkResponse:  func(t *testing.T, token sdk.Token) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, fmt.Sprintf("/oauth/device/token/%s", tc.providerName), r.URL.Path)

				w.WriteHeader(tc.serverStatus)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			sdkConf := sdk.Config{
				UsersURL: server.URL,
			}
			mgsdk := sdk.NewSDK(sdkConf)

			token, err := mgsdk.OAuthDeviceToken(context.Background(), tc.providerName, tc.deviceCode)

			if tc.expectedErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			tc.checkResponse(t, token)
		})
	}
}

func TestOAuthDeviceFlow(t *testing.T) {
	t.Run("complete device flow integration", func(t *testing.T) {
		var savedDeviceCode string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/oauth/device/code/google":
				response := `{
					"device_code": "device_code_123",
					"user_code": "ABCD-EFGH",
					"verification_uri": "https://example.com/device",
					"expires_in": 600,
					"interval": 3
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
				savedDeviceCode = "device_code_123"

			case "/oauth/device/token/google":
				// Simulate polling: first call returns pending, second returns token
				if savedDeviceCode == "device_code_123" {
					// First call - pending
					w.WriteHeader(http.StatusAccepted)
					w.Write([]byte(`{"error": "authorization pending"}`))
					savedDeviceCode = "approved" // Mark as approved for next call
				} else {
					// Second call - success
					response := `{
						"access_token": "access_token_123",
						"refresh_token": "refresh_token_456"
					}`
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(response))
				}
			}
		}))
		defer server.Close()

		sdkConf := sdk.Config{
			UsersURL: server.URL,
		}
		mgsdk := sdk.NewSDK(sdkConf)

		// Step 1: Get device code
		deviceCode, err := mgsdk.OAuthDeviceCode(context.Background(), "google")
		require.NoError(t, err)
		assert.Equal(t, "device_code_123", deviceCode.DeviceCode)
		assert.Equal(t, "ABCD-EFGH", deviceCode.UserCode)

		// Step 2: First poll - pending
		_, err = mgsdk.OAuthDeviceToken(context.Background(), "google", deviceCode.DeviceCode)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization pending")

		// Step 3: Second poll - success
		token, err := mgsdk.OAuthDeviceToken(context.Background(), "google", deviceCode.DeviceCode)
		require.NoError(t, err)
		assert.Equal(t, "access_token_123", token.AccessToken)
		assert.Equal(t, "refresh_token_456", token.RefreshToken)
	})
}

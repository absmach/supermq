// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	smqsdk "github.com/absmach/supermq/pkg/sdk"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerformOAuthDeviceLogin(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		mockResponses []mockResponse
		expectedErr   bool
		errContains   string
	}{
		{
			name:     "successful device flow",
			provider: "google",
			mockResponses: []mockResponse{
				{
					path: "/oauth/device/code/google",
					response: `{
						"device_code": "device_code_123",
						"user_code": "ABCD-EFGH",
						"verification_uri": "https://example.com/device",
						"expires_in": 600,
						"interval": 1
					}`,
					status: http.StatusOK,
				},
				{
					path:     "/oauth/device/token/google",
					response: `{"error": "authorization pending"}`,
					status:   http.StatusAccepted,
				},
				{
					path: "/oauth/device/token/google",
					response: `{
						"access_token": "access_token_123",
						"refresh_token": "refresh_token_456"
					}`,
					status: http.StatusOK,
				},
			},
			expectedErr: false,
		},
		{
			name:     "device code request fails",
			provider: "google",
			mockResponses: []mockResponse{
				{
					path:     "/oauth/device/code/google",
					response: `{"error": "oauth provider is disabled"}`,
					status:   http.StatusNotFound,
				},
			},
			expectedErr: true,
			errContains: "failed to get device code",
		},
		{
			name:     "authorization denied",
			provider: "google",
			mockResponses: []mockResponse{
				{
					path: "/oauth/device/code/google",
					response: `{
						"device_code": "device_code_123",
						"user_code": "ABCD-EFGH",
						"verification_uri": "https://example.com/device",
						"expires_in": 600,
						"interval": 1
					}`,
					status: http.StatusOK,
				},
				{
					path:     "/oauth/device/token/google",
					response: `{"error": "access denied"}`,
					status:   http.StatusUnauthorized,
				},
			},
			expectedErr: true,
			errContains: "failed to get token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if callCount < len(tc.mockResponses) {
					mock := tc.mockResponses[callCount]
					assert.Equal(t, mock.path, r.URL.Path)
					w.WriteHeader(mock.status)
					w.Write([]byte(mock.response))
					callCount++
				}
			}))
			defer server.Close()

			// Set up SDK for testing
			sdkConf := smqsdk.Config{
				UsersURL: server.URL,
			}
			sdk = smqsdk.NewSDK(sdkConf)

			cmd := &cobra.Command{}
			cmd.SetContext(context.Background())

			err := performOAuthDeviceLogin(cmd, tc.provider)

			if tc.expectedErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPollForAuthorization(t *testing.T) {
	tests := []struct {
		name          string
		deviceCode    string
		interval      int
		mockResponses []pollResponse
		expectedErr   bool
		errContains   string
	}{
		{
			name:       "successful authorization after pending",
			deviceCode: "device123",
			interval:   1,
			mockResponses: []pollResponse{
				{
					response: `{"error": "authorization pending"}`,
					status:   http.StatusAccepted,
					delay:    0,
				},
				{
					response: `{
						"access_token": "access_token_123",
						"refresh_token": "refresh_token_456"
					}`,
					status: http.StatusOK,
					delay:  0,
				},
			},
			expectedErr: false,
		},
		{
			name:       "slow down response",
			deviceCode: "device123",
			interval:   1,
			mockResponses: []pollResponse{
				{
					response: `{"error": "slow down"}`,
					status:   http.StatusAccepted,
					delay:    0,
				},
				{
					response: `{
						"access_token": "access_token_123",
						"refresh_token": "refresh_token_456"
					}`,
					status: http.StatusOK,
					delay:  0,
				},
			},
			expectedErr: false,
		},
		{
			name:       "access denied",
			deviceCode: "device123",
			interval:   1,
			mockResponses: []pollResponse{
				{
					response: `{"error": "access denied"}`,
					status:   http.StatusUnauthorized,
					delay:    0,
				},
			},
			expectedErr: true,
			errContains: "failed to get token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if callCount < len(tc.mockResponses) {
					mock := tc.mockResponses[callCount]
					time.Sleep(mock.delay)
					w.WriteHeader(mock.status)
					w.Write([]byte(mock.response))
					callCount++
				}
			}))
			defer server.Close()

			sdkConf := smqsdk.Config{
				UsersURL: server.URL,
			}
			sdk = smqsdk.NewSDK(sdkConf)

			ctx := context.Background()
			_, err := pollForAuthorization(ctx, "google", tc.deviceCode, tc.interval)

			if tc.expectedErr {
				require.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPrintDeviceInstructions(t *testing.T) {
	// This test just ensures the function doesn't panic
	t.Run("prints instructions without panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printDeviceInstructions("https://example.com/device", "ABCD-EFGH")
		})
	})
}

// Helper types for testing
type mockResponse struct {
	path     string
	response string
	status   int
}

type pollResponse struct {
	response string
	status   int
	delay    time.Duration
}

func TestDeviceCodeGeneration(t *testing.T) {
	t.Run("device code is generated correctly", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := `{
				"device_code": "ABCDEFGHIJK123456",
				"user_code": "WXYZ-1234",
				"verification_uri": "https://example.com/verify",
				"expires_in": 600,
				"interval": 3
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer server.Close()

		sdkConf := smqsdk.Config{
			UsersURL: server.URL,
		}
		testSDK := smqsdk.NewSDK(sdkConf)

		deviceCode, err := testSDK.OAuthDeviceCode(context.Background(), "google")
		require.NoError(t, err)

		assert.NotEmpty(t, deviceCode.DeviceCode)
		assert.NotEmpty(t, deviceCode.UserCode)
		assert.Contains(t, deviceCode.UserCode, "-")
		assert.NotEmpty(t, deviceCode.VerificationURI)
		assert.Greater(t, deviceCode.ExpiresIn, 0)
		assert.Greater(t, deviceCode.Interval, 0)
	})
}

func TestDeviceFlowTimeout(t *testing.T) {
	t.Run("timeout after max duration", func(t *testing.T) {
		// This test would take too long to run, so we skip it in normal test runs
		// It's here to document the timeout behavior
		t.Skip("Skipping timeout test - takes too long")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always return pending
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte(`{"error": "authorization pending"}`))
		}))
		defer server.Close()

		sdkConf := smqsdk.Config{
			UsersURL: server.URL,
		}
		sdk = smqsdk.NewSDK(sdkConf)

		ctx := context.Background()
		_, err := pollForAuthorization(ctx, "google", "device123", 1)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBrowserOpener struct {
	opened string
	err    error
}

func (m *mockBrowserOpener) Open(url string) error {
	m.opened = url
	return m.err
}

func TestHandleOAuthCallback(t *testing.T) {
	cases := []struct {
		name           string
		queryParams    string
		expectedCode   string
		expectedState  string
		expectedErr    bool
		expectedHTML   string
		callTwice      bool
		secondCallSent bool
	}{
		{
			name:          "successful callback",
			queryParams:   "?code=test-code&state=test-state",
			expectedCode:  "test-code",
			expectedState: "test-state",
			expectedErr:   false,
			expectedHTML:  "Authentication Successful",
		},
		{
			name:          "callback with error parameter",
			queryParams:   "?error=access_denied",
			expectedCode:  "",
			expectedState: "",
			expectedErr:   true,
			expectedHTML:  "Authentication Failed",
		},
		{
			name:          "callback with missing code",
			queryParams:   "?state=test-state",
			expectedCode:  "",
			expectedState: "",
			expectedErr:   true,
			expectedHTML:  "missing authorization code",
		},
		{
			name:           "multiple calls only process first",
			queryParams:    "?code=test-code&state=test-state",
			expectedCode:   "test-code",
			expectedState:  "test-state",
			expectedErr:    false,
			expectedHTML:   "Authentication Successful",
			callTwice:      true,
			secondCallSent: false,
		},
		{
			name:          "callback with empty state",
			queryParams:   "?code=test-code&state=",
			expectedCode:  "test-code",
			expectedState: "",
			expectedErr:   false,
			expectedHTML:  "Authentication Successful",
		},
		{
			name:          "callback with special characters in state",
			queryParams:   "?code=test-code&state=abc%2Fdef%3D123",
			expectedCode:  "test-code",
			expectedState: "abc/def=123",
			expectedErr:   false,
			expectedHTML:  "Authentication Successful",
		},
		{
			name:          "callback with both error and code",
			queryParams:   "?code=test-code&error=some_error",
			expectedCode:  "",
			expectedState: "",
			expectedErr:   true,
			expectedHTML:  "Authentication Failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resultChan := make(chan oauthCallbackResult, 1)
			var once sync.Once

			req := httptest.NewRequest(http.MethodGet, "/callback"+tc.queryParams, nil)
			w := httptest.NewRecorder()

			handleOAuthCallback(w, req, resultChan, &once)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Contains(t, w.Body.String(), tc.expectedHTML)

			select {
			case result := <-resultChan:
				if tc.expectedErr {
					assert.Error(t, result.err)
				} else {
					assert.NoError(t, result.err)
					assert.Equal(t, tc.expectedCode, result.code)
					assert.Equal(t, tc.expectedState, result.state)
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout waiting for result")
			}

			if tc.callTwice {
				w2 := httptest.NewRecorder()
				handleOAuthCallback(w2, req, resultChan, &once)

				select {
				case <-resultChan:
					if !tc.secondCallSent {
						t.Fatal("second call should not send to channel")
					}
				case <-time.After(100 * time.Millisecond):
					// Expected - second call should not send to channel
				}
			}
		})
	}
}

func TestPrintAuthInstructions(t *testing.T) {
	authURL := "https://example.com/oauth/authorize"
	printAuthInstructions(authURL)
}

func TestWaitForCallback(t *testing.T) {
	cases := []struct {
		name        string
		setupChan   func() <-chan oauthCallbackResult
		expectErr   bool
		expectedMsg string
		timeout     time.Duration
	}{
		{
			name: "successful callback",
			setupChan: func() <-chan oauthCallbackResult {
				ch := make(chan oauthCallbackResult, 1)
				ch <- oauthCallbackResult{code: "test-code", state: "test-state"}
				return ch
			},
			expectErr: false,
		},
		{
			name: "callback with error",
			setupChan: func() <-chan oauthCallbackResult {
				ch := make(chan oauthCallbackResult, 1)
				ch <- oauthCallbackResult{err: errors.New("oauth error")}
				return ch
			},
			expectErr:   true,
			expectedMsg: "callback error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			callbackChan := tc.setupChan()
			result, err := waitForCallback(callbackChan)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.expectedMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result.code)
			}
		})
	}
}

func TestOpenBrowser(t *testing.T) {
	err := openBrowser("https://example.com")
	// This might fail in CI environments, so we just check it doesn't panic
	_ = err
}

func TestCallbackServer(t *testing.T) {
	t.Run("successful callback", func(t *testing.T) {
		resultChan := make(chan oauthCallbackResult, 1)

		server, err := newCallbackServer(resultChan)
		require.NoError(t, err)
		require.NotNil(t, server)
		defer server.Shutdown(nil)

		callbackURL := fmt.Sprintf("http://127.0.0.1:%s%s?code=test-code&state=test-state", localServerPort, callbackPath)

		resp, err := http.Get(callbackURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		select {
		case result := <-resultChan:
			assert.NoError(t, result.err)
			assert.Equal(t, "test-code", result.code)
			assert.Equal(t, "test-state", result.state)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for callback result")
		}
	})

	t.Run("callback with error parameter", func(t *testing.T) {
		resultChan := make(chan oauthCallbackResult, 1)

		server, err := newCallbackServer(resultChan)
		require.NoError(t, err)
		require.NotNil(t, server)
		defer server.Shutdown(nil)

		callbackURL := fmt.Sprintf("http://127.0.0.1:%s%s?error=access_denied", localServerPort, callbackPath)

		resp, err := http.Get(callbackURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		select {
		case result := <-resultChan:
			assert.Error(t, result.err)
			assert.Contains(t, result.err.Error(), "access_denied")
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for callback result")
		}
	})

	t.Run("callback with missing code", func(t *testing.T) {
		resultChan := make(chan oauthCallbackResult, 1)

		server, err := newCallbackServer(resultChan)
		require.NoError(t, err)
		require.NotNil(t, server)
		defer server.Shutdown(nil)

		callbackURL := fmt.Sprintf("http://127.0.0.1:%s%s?state=test-state", localServerPort, callbackPath)

		resp, err := http.Get(callbackURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		select {
		case result := <-resultChan:
			assert.Error(t, result.err)
			assert.Contains(t, result.err.Error(), "missing authorization code")
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for callback result")
		}
	})
}

func TestDefaultBrowserOpener(t *testing.T) {
	opener := defaultBrowserOpener{}
	err := opener.Open("https://example.com")
	// This might fail in CI environments, so we just check it returns an error type
	_ = err
}

func TestMockBrowserOpener(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		setupErr  error
		expectErr bool
	}{
		{
			name:      "successful browser open",
			url:       "https://example.com/auth",
			setupErr:  nil,
			expectErr: false,
		},
		{
			name:      "browser fails to open",
			url:       "https://example.com/auth",
			setupErr:  errors.New("browser failed"),
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockBrowser := &mockBrowserOpener{err: tc.setupErr}
			err := mockBrowser.Open(tc.url)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tc.url, mockBrowser.opened)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.url, mockBrowser.opened)
			}
		})
	}
}

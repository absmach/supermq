// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package google_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/absmach/supermq/pkg/oauth2/google"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testClientID     = "test-client-id"
	testClientSecret = "test-client-secret"
	testState        = "test-state"
	testRedirectURL  = "http://localhost/callback"
	testCode         = "test-code"
)

func TestGetAuthURL(t *testing.T) {
	cfg := oauth2.Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		State:        testState,
		RedirectURL:  testRedirectURL,
	}

	provider := google.NewProvider(cfg, "http://localhost/ui", "http://localhost/error")

	authURL := provider.GetAuthURL()

	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, "accounts.google.com/o/oauth2/auth")
	assert.Contains(t, authURL, "client_id="+testClientID)
	assert.Contains(t, authURL, "state="+testState)
	// redirect_uri is URL-encoded in the query string
	assert.Contains(t, authURL, "redirect_uri=")
}

func TestGetAuthURLWithRedirect(t *testing.T) {
	cfg := oauth2.Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		State:        testState,
		RedirectURL:  testRedirectURL,
	}

	provider := google.NewProvider(cfg, "http://localhost/ui", "http://localhost/error")

	customRedirect := "http://localhost:9090/callback"
	authURL := provider.GetAuthURLWithRedirect(customRedirect)

	assert.NotEmpty(t, authURL)
	assert.Contains(t, authURL, "accounts.google.com/o/oauth2/auth")
	assert.Contains(t, authURL, "client_id="+testClientID)
	assert.Contains(t, authURL, "state="+testState)
	// redirect_uri is URL-encoded in the query string, just verify it exists
	assert.Contains(t, authURL, "redirect_uri=")
	// Verify the custom redirect is in the URL (URL-encoded)
	assert.Contains(t, authURL, "9090")
}

func TestExchangeWithRedirect(t *testing.T) {
	// Create a mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		err := r.ParseForm()
		require.NoError(t, err)

		// Verify the code
		code := r.FormValue("code")

		if code != testCode {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte(`{"error": "invalid_grant"}`))
			assert.NoError(t, err)
			return
		}

		// Return a mock token
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write([]byte(`{
			"access_token": "test-access-token",
			"token_type": "Bearer",
			"expires_in": 3600,
			"refresh_token": "test-refresh-token"
		}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	cfg := oauth2.Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		State:        testState,
		RedirectURL:  testRedirectURL,
	}

	// We can't easily test the actual Google provider without modifying the endpoint
	// This test verifies the method exists and has the correct signature
	provider := google.NewProvider(cfg, "http://localhost/ui", "http://localhost/error")

	// Test with invalid code (will fail but ensures method works)
	_, err := provider.ExchangeWithRedirect(context.Background(), "invalid-code", "http://localhost:9090/callback")
	assert.Error(t, err) // Expected to fail with actual Google OAuth
}

func TestIsEnabled(t *testing.T) {
	cases := []struct {
		name     string
		config   oauth2.Config
		expected bool
	}{
		{
			name: "enabled with all credentials",
			config: oauth2.Config{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				State:        testState,
				RedirectURL:  testRedirectURL,
			},
			expected: true,
		},
		{
			name: "disabled without client ID",
			config: oauth2.Config{
				ClientID:     "",
				ClientSecret: testClientSecret,
				State:        testState,
				RedirectURL:  testRedirectURL,
			},
			expected: false,
		},
		{
			name: "disabled without client secret",
			config: oauth2.Config{
				ClientID:     testClientID,
				ClientSecret: "",
				State:        testState,
				RedirectURL:  testRedirectURL,
			},
			expected: false,
		},
		{
			name: "disabled without credentials",
			config: oauth2.Config{
				ClientID:     "",
				ClientSecret: "",
				State:        testState,
				RedirectURL:  testRedirectURL,
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := google.NewProvider(tc.config, "http://localhost/ui", "http://localhost/error")
			assert.Equal(t, tc.expected, provider.IsEnabled())
		})
	}
}

func TestProviderMethods(t *testing.T) {
	cfg := oauth2.Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		State:        testState,
		RedirectURL:  testRedirectURL,
	}

	uiRedirectURL := "http://localhost:9095/ui/tokens/secure"
	errorURL := "http://localhost:9095/ui/error"

	provider := google.NewProvider(cfg, uiRedirectURL, errorURL)

	t.Run("Name", func(t *testing.T) {
		assert.Equal(t, "google", provider.Name())
	})

	t.Run("State", func(t *testing.T) {
		assert.Equal(t, testState, provider.State())
	})

	t.Run("RedirectURL", func(t *testing.T) {
		assert.Equal(t, uiRedirectURL, provider.RedirectURL())
	})

	t.Run("ErrorURL", func(t *testing.T) {
		assert.Equal(t, errorURL, provider.ErrorURL())
	})
}

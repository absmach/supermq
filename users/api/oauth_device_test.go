// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	useroauth "github.com/absmach/supermq/users/oauth"
	"github.com/absmach/supermq/users/oauth/store"
	"github.com/absmach/supermq/users"
	"github.com/absmach/supermq/users/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	goauth2 "golang.org/x/oauth2"
	"google.golang.org/grpc"
)

// func TestGenerateUserCode(t *testing.T) {
	code, err := generateUserCode()
	assert.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Contains(t, code, "-")
	assert.Len(t, code, 9) // XXXX-XXXX format
}

// func TestGenerateDeviceCode(t *testing.T) {
	code, err := generateDeviceCode()
	assert.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Greater(t, len(code), 40) // Base32 encoded 32 bytes
}

func TestInMemoryDeviceCodeStore(t *testing.T) {
	store := store.NewInMemoryDeviceCodeStore()

	code := store.DeviceCode{
		DeviceCode:      "device123",
		UserCode:        "ABCD-EFGH",
		VerificationURI: "http://example.com/verify",
		ExpiresIn:       600,
		Interval:        3,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "state123",
	}

	t.Run("Save and Get", func(t *testing.T) {
		err := store.Save(code)
		assert.NoError(t, err)

		retrieved, err := store.Get(code.DeviceCode)
		assert.NoError(t, err)
		assert.Equal(t, code.DeviceCode, retrieved.DeviceCode)
		assert.Equal(t, code.UserCode, retrieved.UserCode)
	})

	t.Run("GetByUserCode", func(t *testing.T) {
		retrieved, err := store.GetByUserCode(code.UserCode)
		assert.NoError(t, err)
		assert.Equal(t, code.DeviceCode, retrieved.DeviceCode)
	})

	t.Run("Update", func(t *testing.T) {
		code.Approved = true
		code.AccessToken = "access_token_123"
		err := store.Update(code)
		assert.NoError(t, err)

		retrieved, err := store.Get(code.DeviceCode)
		assert.NoError(t, err)
		assert.True(t, retrieved.Approved)
		assert.Equal(t, "access_token_123", retrieved.AccessToken)
	})

	t.Run("Delete", func(t *testing.T) {
		err := store.Delete(code.DeviceCode)
		assert.NoError(t, err)

		_, err = store.Get(code.DeviceCode)
		assert.Error(t, err)

		_, err = store.GetByUserCode(code.UserCode)
		assert.Error(t, err)
	})

	t.Run("Get non-existent", func(t *testing.T) {
		_, err := store.Get("nonexistent")
		assert.Error(t, err)
	})

	t.Run("Update non-existent", func(t *testing.T) {
		err := store.Update(store.DeviceCode{DeviceCode: "nonexistent"})
		assert.Error(t, err)
	})
}

func TestDeviceCodeHandler(t *testing.T) {
	provider := new(MockOAuthProvider)
	provider.On("Name").Return("google")
	provider.On("IsEnabled").Return(true)

	store := store.NewInMemoryDeviceCodeStore()

	tests := []struct {
		name           string
		providerName   string
		enabled        bool
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "successful device code generation",
			providerName:   "google",
			enabled:        true,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var deviceCode DeviceCode
				err := json.NewDecoder(rec.Body).Decode(&deviceCode)
				assert.NoError(t, err)
				assert.NotEmpty(t, deviceCode.DeviceCode)
				assert.NotEmpty(t, deviceCode.UserCode)
				assert.NotEmpty(t, deviceCode.VerificationURI)
				assert.Greater(t, deviceCode.ExpiresIn, 0)
				assert.Greater(t, deviceCode.Interval, 0)
			},
		},
		{
			name:           "provider disabled",
			providerName:   "google",
			enabled:        false,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "oauth provider is disabled", resp.Error)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(MockOAuthProvider)
			provider.On("Name").Return(tc.providerName)
			provider.On("IsEnabled").Return(tc.enabled)
			provider.On("State").Return("state123")

			handler := deviceCodeHandler(provider, store)

			req := httptest.NewRequest(http.MethodPost, "/oauth/device/code/google", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.checkResponse(t, rec)
		})
	}
}

func TestDeviceTokenHandler(t *testing.T) {
	svc := new(mocks.Service)
	tokenClient := new(MockTokenServiceClient)
	store := store.NewInMemoryDeviceCodeStore()

	// Save a device code
	deviceCode := store.DeviceCode{
		DeviceCode:      "device123",
		UserCode:        "ABCD-EFGH",
		VerificationURI: "http://example.com/verify",
		ExpiresIn:       600,
		Interval:        3,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "state123",
		LastPoll:        time.Now().Add(-5 * time.Second),
	}
	store.Save(deviceCode)

	tests := []struct {
		name           string
		deviceCode     string
		setupCode      func()
		enabled        bool
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "authorization pending",
			deviceCode: deviceCode.DeviceCode,
			setupCode: func() {
				// Code is pending (not approved)
			},
			enabled:        true,
			expectedStatus: http.StatusAccepted,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "authorization pending", resp.Error)
			},
		},
		{
			name:       "invalid device code",
			deviceCode: "invalid",
			setupCode:  func() {},
			enabled:    true,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Contains(t, resp.Error, "invalid device code")
			},
		},
		{
			name:       "provider disabled",
			deviceCode: deviceCode.DeviceCode,
			setupCode:  func() {},
			enabled:    false,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "oauth provider is disabled", resp.Error)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupCode()

			provider := new(MockOAuthProvider)
			provider.On("Name").Return("google")
			provider.On("IsEnabled").Return(tc.enabled)

			handler := deviceTokenHandler(provider, oauthSvc, store, svc, tokenClient)

			reqBody, _ := json.Marshal(map[string]string{
				"device_code": tc.deviceCode,
			})
			req := httptest.NewRequest(http.MethodPost, "/oauth/device/token/google", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.checkResponse(t, rec)
		})
	}
}

func TestDeviceVerifyHandler(t *testing.T) {
	svc := new(mocks.Service)
	tokenClient := new(MockTokenServiceClient)
	store := store.NewInMemoryDeviceCodeStore()

	// Save a device code
	deviceCode := store.DeviceCode{
		DeviceCode:      "device123",
		UserCode:        "ABCD-EFGH",
		VerificationURI: "http://example.com/verify",
		ExpiresIn:       600,
		Interval:        3,
		Provider:        "google",
		CreatedAt:       time.Now(),
		State:           "state123",
	}
	store.Save(deviceCode)

	tests := []struct {
		name           string
		userCode       string
		code           string
		approve        bool
		setupMocks     func(*MockOAuthProvider)
		enabled        bool
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "deny authorization",
			userCode: deviceCode.UserCode,
			code:     "",
			approve:  false,
			setupMocks: func(provider *MockOAuthProvider) {
				provider.On("IsEnabled").Return(true)
			},
			enabled:        true,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "denied", resp["status"])
			},
		},
		{
			name:     "invalid user code",
			userCode: "INVALID",
			code:     "",
			approve:  false,
			setupMocks: func(provider *MockOAuthProvider) {
				// No setup needed
			},
			enabled:        true,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Contains(t, resp.Error, "invalid user code")
			},
		},
		{
			name:     "provider disabled",
			userCode: deviceCode.UserCode,
			code:     "",
			approve:  false,
			setupMocks: func(provider *MockOAuthProvider) {
				provider.On("IsEnabled").Return(false)
			},
			enabled:        false,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, "oauth provider is disabled", resp.Error)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(MockOAuthProvider)
			provider.On("Name").Return("google")
			tc.setupMocks(provider)

			handler := deviceVerifyHandler(store, svc, tokenClient, provider)

			reqBody, _ := json.Marshal(map[string]interface{}{
				"user_code": tc.userCode,
				"code":      tc.code,
				"approve":   tc.approve,
			})
			req := httptest.NewRequest(http.MethodPost, "/oauth/device/verify", bytes.NewReader(reqBody))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.checkResponse(t, rec)
		})
	}
}

// Mock OAuth provider for testing
type MockOAuthProvider struct {
	mock.Mock
}

func (m *MockOAuthProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOAuthProvider) State() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOAuthProvider) RedirectURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOAuthProvider) ErrorURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOAuthProvider) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockOAuthProvider) Exchange(ctx context.Context, code string) (goauth2.Token, error) {
	args := m.Called(ctx, code)
	return args.Get(0).(goauth2.Token), args.Error(1)
}

func (m *MockOAuthProvider) ExchangeWithRedirect(ctx context.Context, code, redirectURL string) (goauth2.Token, error) {
	args := m.Called(ctx, code, redirectURL)
	return args.Get(0).(goauth2.Token), args.Error(1)
}

func (m *MockOAuthProvider) UserInfo(accessToken string) (users.User, error) {
	args := m.Called(accessToken)
	return args.Get(0).(users.User), args.Error(1)
}

func (m *MockOAuthProvider) GetAuthURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOAuthProvider) GetAuthURLWithRedirect(redirectURL string) string {
	args := m.Called(redirectURL)
	return args.String(0)
}

// Mock token service client for testing
type MockTokenServiceClient struct {
	mock.Mock
}

func (m *MockTokenServiceClient) Issue(ctx context.Context, req *grpcTokenV1.IssueReq, opts ...grpc.CallOption) (*grpcTokenV1.Token, error) {
	useroauth "github.com/absmach/supermq/users/oauth"
	"github.com/absmach/supermq/users/oauth/store"
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*grpcTokenV1.Token), args.Error(1)
	useroauth "github.com/absmach/supermq/users/oauth"
	"github.com/absmach/supermq/users/oauth/store"
}

func (m *MockTokenServiceClient) Refresh(ctx context.Context, req *grpcTokenV1.RefreshReq, opts ...grpc.CallOption) (*grpcTokenV1.Token, error) {
	useroauth "github.com/absmach/supermq/users/oauth"
	"github.com/absmach/supermq/users/oauth/store"
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*grpcTokenV1.Token), args.Error(1)
	useroauth "github.com/absmach/supermq/users/oauth"
	"github.com/absmach/supermq/users/oauth/store"
}

// Add these to users/mocks package
func init() {
	// Register mocks to ensure they're available for testing
}

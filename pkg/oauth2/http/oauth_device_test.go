// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/absmach/supermq/pkg/oauth2"
	useroauth "github.com/absmach/supermq/pkg/oauth2"
	oauthhttp "github.com/absmach/supermq/pkg/oauth2/http"

	"github.com/absmach/supermq/pkg/oauth2/mocks"
	"github.com/absmach/supermq/pkg/oauth2/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const errProviderDisabled = "oauth provider is disabled"

type errorResponse struct {
	Error string `json:"error"`
}

func TestInMemoryDeviceCodeStore(t *testing.T) {
	deviceStore := store.NewInMemoryDeviceCodeStore()

	code := oauth2.DeviceCode{
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
		err := deviceStore.Save(code)
		assert.NoError(t, err)

		retrieved, err := deviceStore.Get(code.DeviceCode)
		assert.NoError(t, err)
		assert.Equal(t, code.DeviceCode, retrieved.DeviceCode)
		assert.Equal(t, code.UserCode, retrieved.UserCode)
	})

	t.Run("GetByUserCode", func(t *testing.T) {
		retrieved, err := deviceStore.GetByUserCode(code.UserCode)
		assert.NoError(t, err)
		assert.Equal(t, code.DeviceCode, retrieved.DeviceCode)
	})

	t.Run("Update", func(t *testing.T) {
		code.Approved = true
		code.AccessToken = "access_token_123"
		err := deviceStore.Update(code)
		assert.NoError(t, err)

		retrieved, err := deviceStore.Get(code.DeviceCode)
		assert.NoError(t, err)
		assert.True(t, retrieved.Approved)
		assert.Equal(t, "access_token_123", retrieved.AccessToken)
	})

	t.Run("Delete", func(t *testing.T) {
		err := deviceStore.Delete(code.DeviceCode)
		assert.NoError(t, err)

		_, err = deviceStore.Get(code.DeviceCode)
		assert.Error(t, err)

		_, err = deviceStore.GetByUserCode(code.UserCode)
		assert.Error(t, err)
	})

	t.Run("Get non-existent", func(t *testing.T) {
		_, err := deviceStore.Get("nonexistent")
		assert.Error(t, err)
	})

	t.Run("Update non-existent", func(t *testing.T) {
		err := deviceStore.Update(oauth2.DeviceCode{DeviceCode: "nonexistent"})
		assert.Error(t, err)
	})
}

func TestDeviceCodeHandler(t *testing.T) {
	tests := []struct {
		name           string
		providerName   string
		enabled        bool
		setupMocks     func(*mocks.Service, *mocks.Provider)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "successful device code generation",
			providerName: "google",
			enabled:      true,
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				provider.On("IsEnabled").Return(true)
				mockCode := oauth2.DeviceCode{
					DeviceCode:      "mock-device-code",
					UserCode:        "ABCD-EFGH",
					VerificationURI: "http://example.com/verify",
					ExpiresIn:       600,
					Interval:        5,
				}
				oauthSvc.On("CreateDeviceCode", mock.Anything, provider, mock.Anything).
					Return(mockCode, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var deviceCode oauth2.DeviceCode
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
			name:         "provider disabled",
			providerName: "google",
			enabled:      false,
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				provider.On("IsEnabled").Return(false)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, errProviderDisabled, resp.Error)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(mocks.Provider)
			provider.On("Name").Return(tc.providerName)
			oauthSvc := new(mocks.Service)

			tc.setupMocks(oauthSvc, provider)

			handler := oauthhttp.DeviceCodeHandler(provider, oauthSvc)

			req := httptest.NewRequest(http.MethodPost, "/oauth/device/code/google", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
			tc.checkResponse(t, rec)
		})
	}
}

func TestDeviceTokenHandler(t *testing.T) {
	tests := []struct {
		name           string
		deviceCode     string
		setupMocks     func(*mocks.Service, *mocks.Provider)
		enabled        bool
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "authorization pending",
			deviceCode: "device123",
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				provider.On("IsEnabled").Return(true)
				oauthSvc.On("PollDeviceToken", mock.Anything, provider, "device123").
					Return(nil, oauth2.ErrDeviceCodePending)
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
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				provider.On("IsEnabled").Return(true)
				oauthSvc.On("PollDeviceToken", mock.Anything, provider, "invalid").
					Return(nil, oauth2.ErrDeviceCodeNotFound)
			},
			enabled:        true,
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
			deviceCode: "device123",
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				provider.On("IsEnabled").Return(false)
			},
			enabled:        false,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, errProviderDisabled, resp.Error)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(mocks.Provider)
			provider.On("Name").Return("google")
			oauthSvc := new(mocks.Service)

			tc.setupMocks(oauthSvc, provider)

			handler := oauthhttp.DeviceTokenHandler(provider, oauthSvc)

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
	tests := []struct {
		name           string
		userCode       string
		code           string
		approve        bool
		setupMocks     func(*mocks.Service, *mocks.Provider)
		enabled        bool
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "deny authorization",
			userCode: "ABCD-EFGH",
			code:     "",
			approve:  false,
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				oauthSvc.On("GetDeviceCodeByUserCode", mock.Anything, "ABCD-EFGH").
					Return(oauth2.DeviceCode{Provider: "google"}, nil)
				provider.On("IsEnabled").Return(true)
				oauthSvc.On("VerifyDevice", mock.Anything, provider, "ABCD-EFGH", "", false).
					Return(nil)
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
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				oauthSvc.On("GetDeviceCodeByUserCode", mock.Anything, "INVALID").
					Return(oauth2.DeviceCode{}, useroauth.ErrUserCodeNotFound)
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
			userCode: "ABCD-EFGH",
			code:     "",
			approve:  false,
			setupMocks: func(oauthSvc *mocks.Service, provider *mocks.Provider) {
				oauthSvc.On("GetDeviceCodeByUserCode", mock.Anything, "ABCD-EFGH").
					Return(oauth2.DeviceCode{Provider: "google"}, nil)
				provider.On("IsEnabled").Return(false)
			},
			enabled:        false,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var resp errorResponse
				err := json.NewDecoder(rec.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, errProviderDisabled, resp.Error)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			provider := new(mocks.Provider)
			provider.On("Name").Return("google")
			oauthSvc := new(mocks.Service)

			tc.setupMocks(oauthSvc, provider)

			handler := oauthhttp.DeviceVerifyHandler(oauthSvc, provider)

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

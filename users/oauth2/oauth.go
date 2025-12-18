// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package oauth provides OAuth2 authentication implementation for users service.
// It handles both web-based OAuth flow and device authorization flow for CLI clients.
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"strings"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	"github.com/absmach/supermq/pkg/oauth2"
	goauth2 "golang.org/x/oauth2"
)

// Re-export constants from pkg/oauth2 for backward compatibility.
const (
	DeviceCodeLength      = oauth2.DeviceCodeLength
	DeviceCodePollTimeout = oauth2.DeviceCodePollTimeout
	CodeCheckInterval     = oauth2.CodeCheckInterval
	DeviceStatePrefix     = oauth2.DeviceStatePrefix
)

// Re-export errors from pkg/oauth2 for backward compatibility.
var (
	ErrDeviceCodeExpired  = oauth2.ErrDeviceCodeExpired
	ErrDeviceCodePending  = oauth2.ErrDeviceCodePending
	ErrSlowDown           = oauth2.ErrSlowDown
	ErrAccessDenied       = oauth2.ErrAccessDenied
	ErrInvalidState       = oauth2.ErrInvalidState
	ErrEmptyCode          = oauth2.ErrEmptyCode
	ErrInvalidProvider    = oauth2.ErrInvalidProvider
	ErrDeviceCodeNotFound = oauth2.ErrDeviceCodeNotFound
	ErrUserCodeNotFound   = oauth2.ErrUserCodeNotFound
)

// Service provides OAuth authentication operations for the users service.
type Service interface {
	// Device flow operations

	// CreateDeviceCode initiates the device authorization flow.
	// It generates device and user codes, and returns the verification URI.
	CreateDeviceCode(ctx context.Context, provider oauth2.Provider, verificationURI string) (oauth2.DeviceCode, error)

	// PollDeviceToken polls for device authorization completion.
	// Returns the JWT token once the user has authorized the device.
	PollDeviceToken(ctx context.Context, provider oauth2.Provider, deviceCode string) (*grpcTokenV1.Token, error)

	// VerifyDevice handles user verification of device codes.
	// It exchanges the OAuth authorization code for a token and marks the device as approved.
	VerifyDevice(ctx context.Context, provider oauth2.Provider, userCode, oauthCode string, approve bool) error

	// GetDeviceCodeByUserCode retrieves a device code by its user code.
	GetDeviceCodeByUserCode(ctx context.Context, userCode string) (oauth2.DeviceCode, error)

	// Web flow operations

	// ProcessWebCallback handles OAuth callback for web flow.
	// It exchanges the authorization code for a token and creates/updates the user.
	ProcessWebCallback(ctx context.Context, provider oauth2.Provider, code, redirectURL string) (*grpcTokenV1.Token, error)

	// ProcessDeviceCallback handles OAuth callback for device flow.
	// It's called when a user authorizes a device through the web interface.
	ProcessDeviceCallback(ctx context.Context, provider oauth2.Provider, userCode, oauthCode string) error
}

// generateUserCode generates a human-friendly code like "ABCD-EFGH".
func generateUserCode() (string, error) {
	b := make([]byte, DeviceCodeLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	code := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	code = strings.ToUpper(code[:DeviceCodeLength])
	// Format as XXXX-XXXX
	if len(code) >= 8 {
		code = code[:4] + "-" + code[4:8]
	}
	return code, nil
}

// generateDeviceCode generates a random device code.
func generateDeviceCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

// IsDeviceFlowState checks if the state parameter indicates a device flow.
func IsDeviceFlowState(state string) bool {
	return strings.HasPrefix(state, DeviceStatePrefix)
}

// ExtractUserCodeFromState extracts the user code from a device flow state.
func ExtractUserCodeFromState(state string) string {
	return strings.TrimPrefix(state, DeviceStatePrefix)
}

// ExchangeCode exchanges an authorization code for an access token.
// If redirectURL is provided, it uses ExchangeWithRedirect, otherwise uses Exchange.
func ExchangeCode(ctx context.Context, provider oauth2.Provider, code, redirectURL string) (goauth2.Token, error) {
	if redirectURL != "" {
		return provider.ExchangeWithRedirect(ctx, code, redirectURL)
	}
	return provider.Exchange(ctx, code)
}

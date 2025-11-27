// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package oauth provides OAuth2 authentication functionality for users service.
// It handles both web-based OAuth flow and device authorization flow for CLI clients.
package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"
	"time"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/absmach/supermq/users/oauth/store"
	goauth2 "golang.org/x/oauth2"
)

const (
	// DeviceCodeLength is the length of the user code (e.g., "ABCD-EFGH").
	DeviceCodeLength = 8

	// DeviceCodePollTimeout is the timeout for polling device code status.
	DeviceCodePollTimeout = 5 * time.Second

	// CodeCheckInterval is the minimum interval between polling requests.
	CodeCheckInterval = 3 * time.Second

	// DeviceStatePrefix is the prefix used in state parameter for device flow.
	DeviceStatePrefix = "device:"
)

var (
	// ErrDeviceCodeExpired indicates that the device code has expired.
	ErrDeviceCodeExpired = errors.New("device code expired")

	// ErrDeviceCodePending indicates that the user hasn't authorized the device yet.
	ErrDeviceCodePending = errors.New("authorization pending")

	// ErrSlowDown indicates that the client is polling too frequently.
	ErrSlowDown = errors.New("slow down")

	// ErrAccessDenied indicates that the user denied the authorization request.
	ErrAccessDenied = errors.New("access denied")

	// ErrInvalidState indicates that the OAuth state parameter is invalid.
	ErrInvalidState = errors.New("invalid state")

	// ErrEmptyCode indicates that the authorization code is empty.
	ErrEmptyCode = errors.New("empty code")

	// ErrInvalidProvider indicates that the OAuth provider is not found or disabled.
	ErrInvalidProvider = errors.New("invalid provider")

	// ErrDeviceCodeNotFound is an alias for store.ErrDeviceCodeNotFound for backward compatibility.
	ErrDeviceCodeNotFound = store.ErrDeviceCodeNotFound

	// ErrUserCodeNotFound is an alias for store.ErrUserCodeNotFound for backward compatibility.
	ErrUserCodeNotFound = store.ErrUserCodeNotFound
)

// Service provides OAuth authentication operations.
type Service interface {
	// Device flow operations

	// CreateDeviceCode initiates the device authorization flow.
	// It generates device and user codes, and returns the verification URI.
	CreateDeviceCode(ctx context.Context, provider oauth2.Provider, verificationURI string) (store.DeviceCode, error)

	// PollDeviceToken polls for device authorization completion.
	// Returns the JWT token once the user has authorized the device.
	PollDeviceToken(ctx context.Context, provider oauth2.Provider, deviceCode string) (*grpcTokenV1.Token, error)

	// VerifyDevice handles user verification of device codes.
	// It exchanges the OAuth authorization code for a token and marks the device as approved.
	VerifyDevice(ctx context.Context, provider oauth2.Provider, userCode, oauthCode string, approve bool) error

	// GetDeviceCodeByUserCode retrieves a device code by its user code.
	GetDeviceCodeByUserCode(ctx context.Context, userCode string) (store.DeviceCode, error)

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

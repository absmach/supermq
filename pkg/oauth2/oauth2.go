// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth2

import (
	"context"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	"github.com/absmach/supermq/users"
	"golang.org/x/oauth2"
)

// Config is the configuration for the OAuth2 provider.
// This is kept for backward compatibility but deprecated in favor of
// DeviceConfig and UserConfig.
type Config struct {
	ClientID     string `env:"CLIENT_ID"       envDefault:""`
	ClientSecret string `env:"CLIENT_SECRET"   envDefault:""`
	State        string `env:"STATE"           envDefault:""`
	RedirectURL  string `env:"REDIRECT_URL"    envDefault:""`
}

// DeviceConfig is the configuration for the OAuth2 device flow (CLI).
type DeviceConfig struct {
	ClientID     string `env:"DEVICE_CLIENT_ID"       envDefault:""`
	ClientSecret string `env:"DEVICE_CLIENT_SECRET"   envDefault:""`
	State        string `env:"DEVICE_STATE"           envDefault:""`
	RedirectURL  string `env:"DEVICE_REDIRECT_URL"    envDefault:""`
}

// UserConfig is the configuration for the OAuth2 user flow (web).
type UserConfig struct {
	ClientID     string `env:"USER_CLIENT_ID"       envDefault:""`
	ClientSecret string `env:"USER_CLIENT_SECRET"   envDefault:""`
	State        string `env:"USER_STATE"           envDefault:""`
	RedirectURL  string `env:"USER_REDIRECT_URL"    envDefault:""`
}

// ToConfig converts DeviceConfig to Config.
func (dc DeviceConfig) ToConfig() Config {
	return Config{
		ClientID:     dc.ClientID,
		ClientSecret: dc.ClientSecret,
		State:        dc.State,
		RedirectURL:  dc.RedirectURL,
	}
}

// ToConfig converts UserConfig to Config.
func (uc UserConfig) ToConfig() Config {
	return Config{
		ClientID:     uc.ClientID,
		ClientSecret: uc.ClientSecret,
		State:        uc.State,
		RedirectURL:  uc.RedirectURL,
	}
}

// Provider is an interface that provides the OAuth2 flow for a specific provider
// (e.g. Google, GitHub, etc.)
type Provider interface {
	// Name returns the name of the OAuth2 provider.
	Name() string

	// State returns the current state for the OAuth2 flow.
	State() string

	// RedirectURL returns the URL to redirect the user to after completing the OAuth2 flow.
	RedirectURL() string

	// ErrorURL returns the URL to redirect the user to in case of an error during the OAuth2 flow.
	ErrorURL() string

	// IsEnabled checks if the OAuth2 provider is enabled.
	IsEnabled() bool

	// Exchange converts an authorization code into a token.
	Exchange(ctx context.Context, code string) (oauth2.Token, error)

	// ExchangeWithRedirect converts an authorization code into a token using a custom redirect URL.
	ExchangeWithRedirect(ctx context.Context, code, redirectURL string) (oauth2.Token, error)

	// UserInfo retrieves the user's information using the access token.
	UserInfo(accessToken string) (users.User, error)

	// GetAuthURL returns the authorization URL for the OAuth2 flow.
	GetAuthURL() string

	// GetAuthURLWithRedirect returns the authorization URL with a custom redirect URL.
	GetAuthURLWithRedirect(redirectURL string) string
}

// Service provides OAuth authentication operations for the users service.
type Service interface {
	// Device flow operations

	// CreateDeviceCode initiates the device authorization flow.
	// It generates device and user codes, and returns the verification URI.
	CreateDeviceCode(ctx context.Context, provider Provider, verificationURI string) (DeviceCode, error)

	// PollDeviceToken polls for device authorization completion.
	// Returns the JWT token once the user has authorized the device.
	PollDeviceToken(ctx context.Context, provider Provider, deviceCode string) (*grpcTokenV1.Token, error)

	// VerifyDevice handles user verification of device codes.
	// It exchanges the OAuth authorization code for a token and marks the device as approved.
	VerifyDevice(ctx context.Context, provider Provider, userCode, oauthCode string, approve bool) error

	// GetDeviceCodeByUserCode retrieves a device code by its user code.
	GetDeviceCodeByUserCode(ctx context.Context, userCode string) (DeviceCode, error)

	// Web flow operations

	// ProcessWebCallback handles OAuth callback for web flow.
	// It exchanges the authorization code for a token and creates/updates the user.
	ProcessWebCallback(ctx context.Context, provider Provider, code, redirectURL string) (*grpcTokenV1.Token, error)

	// ProcessDeviceCallback handles OAuth callback for device flow.
	// It's called when a user authorizes a device through the web interface.
	ProcessDeviceCallback(ctx context.Context, provider Provider, userCode, oauthCode string) error
}

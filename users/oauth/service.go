// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth

import (
	"context"
	"fmt"
	"time"

	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	smqauth "github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/absmach/supermq/users"
	"github.com/absmach/supermq/users/oauth/store"
)

var _ Service = (*oauthService)(nil)

// oauthService implements the OAuth Service interface.
type oauthService struct {
	deviceStore store.DeviceCodeStore
	userService users.Service
	tokenClient grpcTokenV1.TokenServiceClient
}

// NewOAuthService creates a new OAuth service instance.
func NewOAuthService(deviceStore store.DeviceCodeStore, userService users.Service, tokenClient grpcTokenV1.TokenServiceClient) Service {
	return &oauthService{
		deviceStore: deviceStore,
		userService: userService,
		tokenClient: tokenClient,
	}
}

// CreateDeviceCode initiates the device authorization flow.
func (s *oauthService) CreateDeviceCode(ctx context.Context, provider oauth2.Provider, verificationURI string) (store.DeviceCode, error) {
	if !provider.IsEnabled() {
		return store.DeviceCode{}, ErrInvalidProvider
	}

	userCode, err := generateUserCode()
	if err != nil {
		return store.DeviceCode{}, fmt.Errorf("failed to generate user code: %w", err)
	}

	deviceCode, err := generateDeviceCode()
	if err != nil {
		return store.DeviceCode{}, fmt.Errorf("failed to generate device code: %w", err)
	}

	code := store.DeviceCode{
		DeviceCode:      deviceCode,
		UserCode:        userCode,
		VerificationURI: verificationURI,
		ExpiresIn:       int(store.DeviceCodeExpiry.Seconds()),
		Interval:        int(CodeCheckInterval.Seconds()),
		Provider:        provider.Name(),
		CreatedAt:       time.Now(),
		State:           provider.State(),
	}

	if err := s.deviceStore.Save(code); err != nil {
		return store.DeviceCode{}, fmt.Errorf("failed to save device code: %w", err)
	}

	return code, nil
}

// PollDeviceToken polls for device authorization completion.
func (s *oauthService) PollDeviceToken(ctx context.Context, provider oauth2.Provider, deviceCode string) (*grpcTokenV1.Token, error) {
	if !provider.IsEnabled() {
		return nil, ErrInvalidProvider
	}

	code, err := s.deviceStore.Get(deviceCode)
	if err != nil {
		return nil, ErrDeviceCodeNotFound
	}

	// Check expiration
	if time.Since(code.CreatedAt) > store.DeviceCodeExpiry {
		s.deviceStore.Delete(deviceCode)
		return nil, ErrDeviceCodeExpired
	}

	// Check polling rate
	if time.Since(code.LastPoll) < CodeCheckInterval {
		return nil, ErrSlowDown
	}

	// Update last poll time
	code.LastPoll = time.Now()
	s.deviceStore.Update(code)

	// Check if denied
	if code.Denied {
		s.deviceStore.Delete(deviceCode)
		return nil, ErrAccessDenied
	}

	// Check if approved
	if !code.Approved || code.AccessToken == "" {
		return nil, ErrDeviceCodePending
	}

	// Process the OAuth user and issue tokens
	jwt, err := s.processOAuthUser(ctx, provider, code.AccessToken)
	if err != nil {
		s.deviceStore.Delete(deviceCode)
		return nil, fmt.Errorf("failed to process oauth user: %w", err)
	}

	s.deviceStore.Delete(deviceCode)
	jwt.AccessType = ""
	return jwt, nil
}

// VerifyDevice handles user verification of device codes.
func (s *oauthService) VerifyDevice(ctx context.Context, provider oauth2.Provider, userCode, oauthCode string, approve bool) error {
	if !provider.IsEnabled() {
		return ErrInvalidProvider
	}

	code, err := s.deviceStore.GetByUserCode(userCode)
	if err != nil {
		return err
	}

	// Check expiration
	if time.Since(code.CreatedAt) > store.DeviceCodeExpiry {
		s.deviceStore.Delete(code.DeviceCode)
		return ErrDeviceCodeExpired
	}

	if !approve {
		code.Denied = true
		s.deviceStore.Update(code)
		return nil
	}

	// Exchange authorization code for access token
	token, err := provider.Exchange(ctx, oauthCode)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	code.Approved = true
	code.AccessToken = token.AccessToken
	if err := s.deviceStore.Update(code); err != nil {
		return fmt.Errorf("failed to update device code: %w", err)
	}

	return nil
}

// GetDeviceCodeByUserCode retrieves a device code by its user code.
func (s *oauthService) GetDeviceCodeByUserCode(ctx context.Context, userCode string) (store.DeviceCode, error) {
	return s.deviceStore.GetByUserCode(userCode)
}

// ProcessWebCallback handles OAuth callback for web flow.
func (s *oauthService) ProcessWebCallback(ctx context.Context, provider oauth2.Provider, code, redirectURL string) (*grpcTokenV1.Token, error) {
	if !provider.IsEnabled() {
		return nil, ErrInvalidProvider
	}

	if code == "" {
		return nil, ErrEmptyCode
	}

	token, err := ExchangeCode(ctx, provider, code, redirectURL)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return s.processOAuthUser(ctx, provider, token.AccessToken)
}

// ProcessDeviceCallback handles OAuth callback for device flow.
func (s *oauthService) ProcessDeviceCallback(ctx context.Context, provider oauth2.Provider, userCode, oauthCode string) error {
	return s.VerifyDevice(ctx, provider, userCode, oauthCode, true)
}

// processOAuthUser retrieves user info from the OAuth provider, creates or updates the user,
// adds user policies, and issues a JWT token.
func (s *oauthService) processOAuthUser(ctx context.Context, provider oauth2.Provider, accessToken string) (*grpcTokenV1.Token, error) {
	user, err := provider.UserInfo(accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	user.AuthProvider = provider.Name()
	if user.AuthProvider == "" {
		user.AuthProvider = "oauth"
	}

	user, err = s.userService.OAuthCallback(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to handle oauth callback: %w", err)
	}

	if err := s.userService.OAuthAddUserPolicy(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to add user policy: %w", err)
	}

	return s.tokenClient.Issue(ctx, &grpcTokenV1.IssueReq{
		UserId:   user.ID,
		Type:     uint32(smqauth.AccessKey),
		UserRole: uint32(smqauth.UserRole),
		Verified: !user.VerifiedAt.IsZero(),
	})
}

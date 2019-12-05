// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"context"
	"errors"
	"time"
)

const (
	loginDuration = 10 * time.Hour
	resetDuration = 5 * time.Minute
	issuerName    = "mainflux.authn"
)

var (
	// ErrUnauthorizedAccess represents unauthorized access.
	ErrUnauthorizedAccess = errors.New("unauthorized access")

	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid owner or ID).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrNotFound indicates a non-existing entity request.
	ErrNotFound = errors.New("entity not found")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.New("entity already exists")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Issue issues a new Key.
	Issue(context.Context, string, Key) (Key, error)

	// Revoke removes the Key with the provided id that is
	// issued by the user identified by the provided key.
	Revoke(context.Context, string, string) error

	// Retrieve retrieves data for the Key identified by the provided
	// ID, that is issued by the user identified by the provided key.
	Retrieve(context.Context, string, string) (Key, error)

	// Identify validates token token. If token is valid, content
	// is returned. If token is invalid, or invocation failed for some
	// other reason, non-nil error value is returned in response.
	Identify(context.Context, string) (string, error)
}

type authService struct {
	keys     KeyRepository
	idp      IdentityProvider
	secret   string
	duration time.Duration
}

// New instantiates the auth service implementation.
func New(keys KeyRepository, idp IdentityProvider, secret string) Service {
	return &authService{
		keys:   keys,
		idp:    idp,
		secret: secret,
	}
}

func (svc authService) Issue(ctx context.Context, issuer string, key Key) (Key, error) {
	if key.IssuedAt.UTC().Nanosecond() == 0 {
		return Key{}, ErrInvalidKeyIssuedAt
	}
	switch key.Type {
	case UserKey:
		return svc.userKey(ctx, issuer, key)
	case ResetKey:
		return svc.resetKey(ctx, issuer, key)
	default:
		return svc.loginKey(issuer, key)
	}
}

func (svc authService) Revoke(ctx context.Context, issuer, id string) error {
	email, err := svc.login(issuer)
	if err != nil {
		return err
	}

	return svc.keys.Remove(ctx, email, id)
}

func (svc authService) Retrieve(ctx context.Context, issuer, id string) (Key, error) {
	email, err := svc.login(issuer)
	if err != nil {
		return Key{}, err
	}

	return svc.keys.Retrieve(ctx, email, id)
}

func (svc authService) Identify(ctx context.Context, token string) (string, error) {
	c, err := svc.parse(token)
	if err != nil {
		return "", err
	}

	if c.Type == nil {
		return "", ErrUnauthorizedAccess
	}
	switch *c.Type {
	case UserKey:
		k, err := svc.keys.Retrieve(ctx, c.Issuer, c.Id)
		if err != nil {
			return "", err
		}
		// Auto revoke expired key.
		if k.Expired() {
			svc.keys.Remove(ctx, c.Issuer, c.Id)
			return "", ErrKeyExpired
		}
		return c.Issuer, nil
	case ResetKey, LoginKey:
		if c.Issuer != issuerName {
			return "", ErrUnauthorizedAccess
		}
		return c.Subject, nil
	default:
		return "", ErrUnauthorizedAccess
	}
}

func (svc authService) loginKey(issuer string, key Key) (Key, error) {
	key.Secret = issuer
	return svc.tempKey(loginDuration, key)
}

func (svc authService) resetKey(ctx context.Context, issuer string, key Key) (Key, error) {
	issuer, err := svc.login(issuer)
	if err != nil {
		return Key{}, err
	}
	key.Secret = issuer

	return svc.tempKey(resetDuration, key)
}

func (svc authService) tempKey(duration time.Duration, key Key) (Key, error) {
	key.Issuer = issuerName
	exp := key.IssuedAt.Add(duration)
	key.ExpiresAt = &exp
	val, err := svc.issue(key)
	if err != nil {
		return Key{}, err
	}

	key.Secret = val
	return key, nil
}

func (svc authService) userKey(ctx context.Context, issuer string, key Key) (Key, error) {
	email, err := svc.login(issuer)
	if err != nil {
		return Key{}, err
	}
	key.Issuer = email

	id, err := svc.idp.ID()
	if err != nil {
		return Key{}, err
	}
	key.ID = id

	value, err := svc.issue(key)
	if err != nil {
		return Key{}, err
	}
	key.Secret = value

	if _, err := svc.keys.Save(ctx, key); err != nil {
		return Key{}, err
	}

	return key, nil
}

func (svc authService) login(token string) (string, error) {
	c, err := svc.parse(token)
	if err != nil {
		return "", err
	}
	// Only login token is valid token type.
	if c.Type == nil || *c.Type != LoginKey {
		return "", ErrUnauthorizedAccess
	}

	if c.Subject == "" {
		return "", ErrUnauthorizedAccess
	}
	return c.Subject, nil
}

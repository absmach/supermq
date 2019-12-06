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

var _ Service = (*authnService)(nil)

type authnService struct {
	keys KeyRepository
	idp  IdentityProvider
	t    Tokenizer
}

// New instantiates the auth service implementation.
func New(keys KeyRepository, idp IdentityProvider, tokenizer Tokenizer) Service {
	return &authnService{
		t:    tokenizer,
		keys: keys,
		idp:  idp,
	}
}

func (svc authnService) Issue(ctx context.Context, issuer string, key Key) (Key, error) {
	if key.IssuedAt.IsZero() {
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

func (svc authnService) Revoke(ctx context.Context, issuer, id string) error {
	email, err := svc.login(issuer)
	if err != nil {
		return err
	}

	return svc.keys.Remove(ctx, email, id)
}

func (svc authnService) Retrieve(ctx context.Context, issuer, id string) (Key, error) {
	email, err := svc.login(issuer)
	if err != nil {
		return Key{}, err
	}

	return svc.keys.Retrieve(ctx, email, id)
}

func (svc authnService) Identify(ctx context.Context, token string) (string, error) {
	c, err := svc.t.Parse(token)
	if err != nil {
		return "", err
	}

	switch c.Type {
	case UserKey:
		k, err := svc.keys.Retrieve(ctx, c.Issuer, c.ID)
		if err != nil {
			return "", err
		}
		// Auto revoke expired key.
		if k.Expired() {
			svc.keys.Remove(ctx, c.Issuer, c.ID)
			return "", ErrKeyExpired
		}
		return c.Issuer, nil
	case ResetKey, LoginKey:
		if c.Issuer != issuerName {
			return "", ErrUnauthorizedAccess
		}
		return c.Secret, nil
	default:
		return "", ErrUnauthorizedAccess
	}
}

func (svc authnService) loginKey(issuer string, key Key) (Key, error) {
	key.Secret = issuer
	return svc.tempKey(loginDuration, key)
}

func (svc authnService) resetKey(ctx context.Context, issuer string, key Key) (Key, error) {
	issuer, err := svc.login(issuer)
	if err != nil {
		return Key{}, err
	}
	key.Secret = issuer

	return svc.tempKey(resetDuration, key)
}

func (svc authnService) tempKey(duration time.Duration, key Key) (Key, error) {
	key.Issuer = issuerName
	key.ExpiresAt = key.IssuedAt.Add(duration)
	val, err := svc.t.Issue(key)
	if err != nil {
		return Key{}, err
	}

	key.Secret = val
	return key, nil
}

func (svc authnService) userKey(ctx context.Context, issuer string, key Key) (Key, error) {
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

	value, err := svc.t.Issue(key)
	if err != nil {
		return Key{}, err
	}
	key.Secret = value

	if _, err := svc.keys.Save(ctx, key); err != nil {
		return Key{}, err
	}

	return key, nil
}

func (svc authnService) login(token string) (string, error) {
	c, err := svc.t.Parse(token)
	if err != nil {
		return "", err
	}
	// Only login token is valid token type.
	if c.Type != LoginKey {
		return "", ErrUnauthorizedAccess
	}

	if c.Secret == "" {
		return "", ErrUnauthorizedAccess
	}
	return c.Secret, nil
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	sessionDuration = 10 * time.Hour
	resetDuration   = 5 * time.Minute
	issuerName      = "mainflux.auth"
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

type claims struct {
	jwt.StandardClaims
	Type *uint32 `json:"type,omitempty"`
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
		return svc.sessionKey(ctx, issuer, resetDuration, key)
	default:
		return svc.sessionKey(ctx, issuer, sessionDuration, key)
	}
}

func (svc authService) Revoke(ctx context.Context, issuer, id string) error {
	email, err := svc.login(ctx, issuer)
	if err != nil {
		return err
	}
	return svc.keys.Remove(ctx, email, id)
}

func (svc authService) Retrieve(ctx context.Context, issuer, id string) (Key, error) {
	email, err := svc.login(ctx, issuer)
	if err != nil {
		return Key{}, err
	}

	return svc.keys.Retrieve(ctx, email, id)
}

func (svc authService) Identify(ctx context.Context, token string) (string, error) {
	c, err := svc.parseClaims(token)
	if err != nil {
		return "", err
	}

	return svc.identify(ctx, c)
}

func (svc authService) identify(ctx context.Context, cl claims) (string, error) {
	if cl.Type == nil {
		return "", ErrUnauthorizedAccess
	}
	switch *cl.Type {
	case UserKey:
		k, err := svc.keys.Retrieve(ctx, cl.Issuer, cl.Id)
		if err != nil {
			return "", err
		}
		// Auto revoke expired key.
		if k.ExpiresAt != nil && k.ExpiresAt.Before(time.Now()) {
			svc.keys.Remove(ctx, cl.Issuer, cl.Id)
			return "", ErrUnauthorizedAccess
		}
		return cl.Issuer, nil
	case LoginKey:
		if cl.Subject == "" {
			return "", ErrUnauthorizedAccess
		}
		return cl.Subject, nil
	case ResetKey:
		if cl.Issuer == "" {
			return "", ErrUnauthorizedAccess
		}
		return cl.Issuer, nil
	default:
		return "", ErrUnauthorizedAccess
	}
}

func (svc authService) issueJwt(key Key) (string, error) {
	claims := claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:   key.Issuer,
			Subject:  key.Secret,
			IssuedAt: key.IssuedAt.Unix(),
		},
		Type: &key.Type,
	}

	if key.ExpiresAt != nil {
		claims.ExpiresAt = key.ExpiresAt.Unix()
	}
	if key.ID != "" {
		claims.Id = key.ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(svc.secret))
}

func (svc authService) sessionKey(ctx context.Context, issuer string, duration time.Duration, key Key) (Key, error) {
	key.Issuer = issuerName
	key.Secret = issuer
	exp := key.IssuedAt.Add(duration)
	key.ExpiresAt = &exp
	val, err := svc.issueJwt(key)
	if err != nil {
		return Key{}, err
	}
	key.Secret = val
	return key, nil
}

func (svc authService) userKey(ctx context.Context, issuer string, key Key) (Key, error) {
	email, err := svc.login(ctx, issuer)
	if err != nil {
		return Key{}, err
	}
	key.Issuer = email

	id, err := svc.idp.ID()
	if err != nil {
		return Key{}, err
	}
	key.ID = id

	value, err := svc.issueJwt(key)
	if err != nil {
		return Key{}, err
	}
	key.Secret = value
	if _, err := svc.keys.Save(ctx, key); err != nil {
		return Key{}, err
	}

	return key, nil
}

func (svc authService) parseJwt(key string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(key, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorizedAccess
		}
		return []byte(svc.secret), nil
	})

	if err != nil {
		return nil, ErrUnauthorizedAccess
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrUnauthorizedAccess
}

func (svc authService) parseClaims(key string) (claims, error) {
	c := claims{}
	_, err := jwt.ParseWithClaims(key, &c, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorizedAccess
		}
		return []byte(svc.secret), nil
	})

	if err != nil {
		if c.Type != nil && *c.Type == UserKey {
			if e, ok := err.(*jwt.ValidationError); ok && e.Errors == jwt.ValidationErrorExpired {
				return c, nil
			}
		}
		return claims{}, ErrUnauthorizedAccess
	}

	return c, nil
}

func (svc authService) login(ctx context.Context, token string) (string, error) {
	c, err := svc.parseClaims(token)
	if err != nil {
		return "", err
	}
	// Only user can perform actions over API keys.
	if c.Type == nil || *c.Type != LoginKey {
		return "", ErrUnauthorizedAccess
	}
	return svc.identify(ctx, c)
}

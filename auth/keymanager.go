// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// PublicKeyStatus represents the status of a public key.
type PublicKeyStatus int

const (
	// ActiveKeyStatus indicates the key is active and can be used for verification.
	ActiveKeyStatus PublicKeyStatus = iota
	// RetiredKeyStatus indicates the key is retired but still valid for verification during grace period.
	RetiredKeyStatus
)

// PublicKey represents a public key stored in the database.
type PublicKey struct {
	Kid       string          `json:"kid" db:"kid"`
	JWKData   jwk.Key         `json:"jwk_data" db:"jwk_data"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	RetiredAt *time.Time      `json:"retired_at,omitempty" db:"retired_at"`
	Status    PublicKeyStatus `json:"status" db:"status"`
}

// KeyManager represents a manager for JWT keys.
type KeyManager interface {
	SignJWT(token jwt.Token) ([]byte, error)

	ParseJWT(token string) (jwt.Token, error)

	PublicJWKS() []jwk.Key

	Rotate(ctx context.Context) error
}

// PublicKeyRepository represents a repository for storing and retrieving public keys.
type PublicKeyRepository interface {
	// Save stores a public key in the database.
	Save(ctx context.Context, key PublicKey) error

	// Retrieve gets a public key by its ID.
	Retrieve(ctx context.Context, kid string) (PublicKey, error)

	// RetrieveActive gets all active public keys.
	RetrieveActive(ctx context.Context) ([]PublicKey, error)

	// Retire marks a public key as retired.
	Retire(ctx context.Context, kid string) error

	// PurgeExpired removes all expired public keys from the database.
	PurgeExpired(ctx context.Context, expiredBefore time.Time) error
}

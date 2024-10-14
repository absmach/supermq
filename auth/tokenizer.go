// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import "context"

// Tokenizer specifies API for encoding and decoding between string and Key.
//
//go:generate mockery --name Tokenizer --output=./mocks --filename tokenizer.go --quiet --note "Copyright (c) Abstract Machines"
type Tokenizer interface {
	// Issue converts API Key to its string representation.
	Issue(key Key) (token string, err error)

	// Parse extracts API Key data from string token.
	Parse(ctx context.Context, token string) (key Key, err error)

	// Revoke revokes the token.
	Revoke(ctx context.Context, token string) error

	// RetrieveJWKS returns the JSON Web Key Set.
	RetrieveJWKS() (JWKS, error)
}

// TokenRepository specifies token persistence API.
//
//go:generate mockery --name TokenRepository --output=./mocks --filename token.go --quiet --note "Copyright (c) Abstract Machines"
type TokenRepository interface {
	// Save persists the token.
	Save(ctx context.Context, id string) (err error)

	// Contains checks if token with provided ID exists.
	Contains(ctx context.Context, id string) (ok bool)
}

// Cache represents a cache repository. It exposes functionalities
// through `auth` to perform caching.
//
//go:generate mockery --name Cache --output=./mocks --filename cache.go --quiet --note "Copyright (c) Abstract Machines"
type Cache interface {
	// Save saves the key-value pair in the cache.
	Save(ctx context.Context, key, value string) error

	// Contains checks if the key-value pair exists in the cache.
	Contains(ctx context.Context, key, value string) bool

	// Remove removes the key from the cache.
	Remove(ctx context.Context, key string) error
}

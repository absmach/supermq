// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrInvalidKeyIssuedAt indicates that the Key is being used before it's issued.
	ErrInvalidKeyIssuedAt = errors.New("invalid issue time")

	// ErrKeyExpired indicates that the Key is expired.
	ErrKeyExpired = errors.New("use o expired key")
)

const (
	// LoginKey is temporary User key received on successfull login.
	LoginKey uint32 = iota
	// ResetKey represents a key for resseting password.
	ResetKey
	// UserKey enables the one to act on behalf of the user.
	UserKey
)

// Key represents API key.
type Key struct {
	ID        string
	Type      uint32
	Issuer    string
	Secret    string
	IssuedAt  time.Time
	ExpiresAt *time.Time
}

// KeyRepository specifies Key persistence API.
type KeyRepository interface {
	// Save persists the Key. A non-nil error is returned to indicate
	// operation failure
	Save(context.Context, Key) (string, error)

	// Retrieve retrieves Key by its unique identifier.
	Retrieve(context.Context, string, string) (Key, error)

	// Remove removes Key with provided ID.
	Remove(context.Context, string, string) error
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package store provides storage implementations for OAuth device codes.
package store

import (
	"errors"
	"time"
)

const (
	// DeviceCodeExpiry is the time after which device codes expire.
	DeviceCodeExpiry = 10 * time.Minute
)

var (
	// ErrDeviceCodeNotFound indicates that the device code was not found.
	ErrDeviceCodeNotFound = errors.New("device code not found")

	// ErrUserCodeNotFound indicates that the user code was not found.
	ErrUserCodeNotFound = errors.New("user code not found")
)

// DeviceCode represents an OAuth2 device authorization code.
type DeviceCode struct {
	DeviceCode      string    `json:"device_code"`
	UserCode        string    `json:"user_code"`
	VerificationURI string    `json:"verification_uri"`
	ExpiresIn       int       `json:"expires_in"`
	Interval        int       `json:"interval"`
	Provider        string    `json:"provider,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	State           string    `json:"state,omitempty"`
	AccessToken     string    `json:"access_token,omitempty"`
	Approved        bool      `json:"approved,omitempty"`
	Denied          bool      `json:"denied,omitempty"`
	LastPoll        time.Time `json:"last_poll,omitempty"`
}

// DeviceCodeStore manages device authorization codes.
// It provides operations to save, retrieve, update, and delete device codes
// used in the OAuth2 device authorization flow.
type DeviceCodeStore interface {
	// Save stores a new device code.
	Save(code DeviceCode) error

	// Get retrieves a device code by its device code value.
	Get(deviceCode string) (DeviceCode, error)

	// GetByUserCode retrieves a device code by its user code.
	GetByUserCode(userCode string) (DeviceCode, error)

	// Update updates an existing device code.
	Update(code DeviceCode) error

	// Delete removes a device code.
	Delete(deviceCode string) error
}

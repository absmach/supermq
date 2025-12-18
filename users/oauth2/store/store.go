// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Package store provides storage implementations for OAuth device codes.
package store

import (
	"github.com/absmach/supermq/pkg/oauth2"
)

// Re-export constants from pkg/oauth2 for backward compatibility.
const (
	DeviceCodeExpiry = oauth2.DeviceCodeExpiry
)

// Re-export errors from pkg/oauth2 for backward compatibility.
var (
	ErrDeviceCodeNotFound = oauth2.ErrDeviceCodeNotFound
	ErrUserCodeNotFound   = oauth2.ErrUserCodeNotFound
)

// DeviceCode is an alias for oauth2.DeviceCode for backward compatibility.
type DeviceCode = oauth2.DeviceCode

// DeviceCodeStore is an alias for oauth2.DeviceCodeStore for backward compatibility.
type DeviceCodeStore = oauth2.DeviceCodeStore

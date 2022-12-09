// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	// ErrFailedFetch indicates that fetching of entity data failed.
	ErrFailedFetch = errors.NewSDKError(errors.New("failed to fetch entity"))

	// ErrFailedConnect indicates that connecting thing to channel failed.
	ErrFailedConnect = errors.NewSDKError(errors.New("failed to connect thing to channel"))

	// ErrFailedDisconnect indicates that disconnecting thing from a channel failed.
	ErrFailedDisconnect = errors.NewSDKError(errors.New("failed to disconnect thing from channel"))

	// ErrInvalidContentType indicates that non-existent message content type was passed.
	ErrInvalidContentType = errors.NewSDKError(errors.New("Unknown Content Type"))

	// ErrFailedWhitelist failed to whitelist configs
	ErrFailedWhitelist = errors.NewSDKError(errors.New("failed to whitelist"))

	// ErrCerts indicates error fetching certificates.
	ErrCerts = errors.NewSDKError(errors.New("failed to fetch certs data"))

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.NewSDKError(errors.New("failed to remove certificate"))

	// ErrFailedCertUpdate failed to update certs in bootstrap config
	ErrFailedCertUpdate = errors.NewSDKError(errors.New("failed to update certs in bootstrap config"))
)

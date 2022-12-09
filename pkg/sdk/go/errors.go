// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	// ErrFailedFetch indicates that fetching of entity data failed.
	ErrFailedFetch = errors.NewSDKError("failed to fetch entity")

	// ErrFailedConnect indicates that connecting thing to channel failed.
	ErrFailedConnect = errors.NewSDKError("failed to connect thing to channel")

	// ErrFailedDisconnect indicates that disconnecting thing from a channel failed.
	ErrFailedDisconnect = errors.NewSDKError("failed to disconnect thing from channel")

	// ErrInvalidContentType indicates that non-existent message content type was passed.
	ErrInvalidContentType = errors.NewSDKError("Unknown Content Type")

	// ErrFailedWhitelist failed to whitelist configs
	ErrFailedWhitelist = errors.NewSDKError("failed to whitelist")

	// ErrCerts indicates error fetching certificates.
	ErrCerts = errors.NewSDKError("failed to fetch certs data")

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.NewSDKError("failed to remove certificate")

	// ErrFailedCertUpdate failed to update certs in bootstrap config
	ErrFailedCertUpdate = errors.NewSDKError("failed to update certs in bootstrap config")
)

// func encodeError(body []byte, status int) errors.SDKError {
// 	e := struct {
// 		Err string `json:"error"`
// 	}{}

// 	if err := json.Unmarshal(body, &e); err != nil {
// 		return errors.NewSDKError(errors.Wrap(errEncodeError, err).Error())
// 	}

// 	if status != 0 {
// 		return errors.NewSDKErrorWithStatus(e.Err, status)
// 	}

// 	return errors.NewSDKError(e.Err)
// }

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"

	"github.com/mainflux/mainflux/pkg/errors"
)

var errEncodeError = errors.NewSDKError("failed to encode response error")

var (
	// ErrFailedCreation indicates that entity creation failed.
	ErrFailedCreation = errors.NewSDKError("failed to create entity")

	// ErrFailedUpdate indicates that entity update failed.
	ErrFailedUpdate = errors.NewSDKError("failed to update entity")

	// ErrFailedFetch indicates that fetching of entity data failed.
	ErrFailedFetch = errors.NewSDKError("failed to fetch entity")

	// ErrFailedRemoval indicates that entity removal failed.
	ErrFailedRemoval = errors.NewSDKError("failed to remove entity")

	// ErrFailedConnect indicates that connecting thing to channel failed.
	ErrFailedConnect = errors.NewSDKError("failed to connect thing to channel")

	// ErrFailedDisconnect indicates that disconnecting thing from a channel failed.
	ErrFailedDisconnect = errors.NewSDKError("failed to disconnect thing from channel")

	// ErrFailedPublish indicates that publishing message failed.
	ErrFailedPublish = errors.NewSDKError("failed to publish message")

	// ErrFailedRead indicates that read messages failed.
	ErrFailedRead = errors.NewSDKError("failed to read messages")

	// ErrInvalidContentType indicates that non-existent message content type
	// was passed.
	ErrInvalidContentType = errors.NewSDKError("Unknown Content Type")

	// ErrFetchHealth indicates that fetching of health check failed.
	ErrFetchHealth = errors.NewSDKError("failed to fetch health check")

	// ErrFailedWhitelist failed to whitelist configs
	ErrFailedWhitelist = errors.NewSDKError("failed to whitelist")

	// ErrCerts indicates error fetching certificates.
	ErrCerts = errors.NewSDKError("failed to fetch certs data")

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.NewSDKError("failed to remove certificate")

	// ErrFailedCertUpdate failed to update certs in bootstrap config
	ErrFailedCertUpdate = errors.NewSDKError("failed to update certs in bootstrap config")

	// ErrMemberAdd failed to add member to a group.
	ErrMemberAdd = errors.NewSDKError("failed to add member to group")
)

func encodeError(body []byte, status int) error {
	e := struct {
		Err string `json:"error"`
	}{}

	if err := json.Unmarshal(body, &e); err != nil {
		return errors.Wrap(errEncodeError, err)
	}

	if status != 0 {
		return errors.NewSDKErrorWithStatus(e.Err, status)
	}

	return errors.NewSDKError(e.Err)
}

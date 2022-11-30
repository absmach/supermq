// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux/pkg/errors"
)

var errEncodeError = errors.New("failed to encode response error")

var (
	// ErrFailedCreation indicates that entity creation failed.
	ErrFailedCreation = errors.New("failed to create entity")

	// ErrFailedUpdate indicates that entity update failed.
	ErrFailedUpdate = errors.New("failed to update entity")

	// ErrFailedFetch indicates that fetching of entity data failed.
	ErrFailedFetch = errors.New("failed to fetch entity")

	// ErrFailedRemoval indicates that entity removal failed.
	ErrFailedRemoval = errors.New("failed to remove entity")

	// ErrFailedConnect indicates that connecting thing to channel failed.
	ErrFailedConnect = errors.New("failed to connect thing to channel")

	// ErrFailedDisconnect indicates that disconnecting thing from a channel failed.
	ErrFailedDisconnect = errors.New("failed to disconnect thing from channel")

	// ErrFailedPublish indicates that publishing message failed.
	ErrFailedPublish = errors.New("failed to publish message")

	// ErrFailedRead indicates that read messages failed.
	ErrFailedRead = errors.New("failed to read messages")

	// ErrInvalidContentType indicates that non-existent message content type
	// was passed.
	ErrInvalidContentType = errors.New("Unknown Content Type")

	// ErrFetchHealth indicates that fetching of health check failed.
	ErrFetchHealth = errors.New("failed to fetch health check")

	// ErrFailedWhitelist failed to whitelist configs
	ErrFailedWhitelist = errors.New("failed to whitelist")

	// ErrCerts indicates error fetching certificates.
	ErrCerts = errors.New("failed to fetch certs data")

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.New("failed to remove certificate")

	// ErrFailedCertUpdate failed to update certs in bootstrap config
	ErrFailedCertUpdate = errors.New("failed to update certs in bootstrap config")

	// ErrMemberAdd failed to add member to a group.
	ErrMemberAdd = errors.New("failed to add member to group")
)

func encodeError(body []byte, status int) error {
	e := struct {
		Err string `json:"error"`
	}{}

	if err := json.Unmarshal(body, &e); err != nil {
		return errors.Wrap(errEncodeError, err)
	}

	if status != 0 {
		httpStatus := fmt.Sprintf("%d %s", status, http.StatusText(status))
		return errors.Wrap(errors.New(e.Err), errors.New(httpStatus))
	}

	return errors.New(e.Err)
}

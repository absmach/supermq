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

	// ErrInvalidContentType indicates that non-existent message content type was passed.
	ErrInvalidContentType = errors.NewSDKError(errors.New("Unknown Content Type"))

	// ErrCertsRemove indicates failure while cleaning up from the Certs service.
	ErrCertsRemove = errors.NewSDKError(errors.New("failed to remove certificate"))
)

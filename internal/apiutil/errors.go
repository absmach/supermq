// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package apiutil

import (
	"github.com/mainflux/mainflux/pkg/errors"
)

var (
	// ErrMissingToken indicates missing user token.
	ErrMissingToken = errors.New("missing user token")

	// ErrMissingID indicates missing entity ID.
	ErrMissingID = errors.New("missing entity id")

	// ErrMissingKey indicates missing entity key.
	ErrMissingKey = errors.New("missing entity key")

	// ErrInvalidIDFormat indicates an invalid ID format.
	ErrInvalidIDFormat = errors.New("invalid id format provided")

	// ErrNameSize indicates that name size exceeds the max.
	ErrNameSize = errors.New("name size exceeds the max")

	// ErrLimitSize indicates that limit size exceeds the max.
	ErrLimitSize = errors.New("limit size exceeds the max")

	// ErrInvalidOrder indicates an invalid list order.
	ErrInvalidOrder = errors.New("invalid list order provided")

	// ErrInvalidDirection indicates an invalid list direction.
	ErrInvalidDirection = errors.New("invalid list direction provided")

	// ErrEmptyList indicates that entity data is empty.
	ErrEmptyList = errors.New("empty list provided")

	// ErrMalformedPolicy indicates that policies are malformed.
	ErrMalformedPolicy = errors.New("falmormed policy")

	// ErrMissingCertData indicates missing cert data (ttl, key_type or key_bits).
	ErrMissingCertData = errors.New("missing certificate data")

	// ErrInvalidTopic indicates an invalid subscription topic.
	ErrInvalidTopic = errors.New("invalid Subscription topic")

	// ErrInvalidContact indicates an invalid subscription contract.
	ErrInvalidContact = errors.New("invalid Subscription contact")

	// ErrMissingEmail indicates missing email.
	ErrMissingEmail = errors.New("missing email")

	// ErrMissingHost indicates missing host.
	ErrMissingHost = errors.New("missing host")

	// ErrMissingPassword indicates missing password.
	ErrMissingPass = errors.New("missing password")

	// ErrMissingConfPassword indicates missing conf password.
	ErrMissingConfPass = errors.New("missing conf password")

	// ErrInvalidResetPass indicates an invalid reset password.
	ErrInvalidResetPass = errors.New("invalid reset password")
)

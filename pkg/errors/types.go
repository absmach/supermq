// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import "errors"

var (
	// ErrMalformedEntity indicates a malformed entity specification.
	ErrMalformedEntity = New("malformed entity specification")

	// ErrUnsupportedContentType indicates invalid content type.
	ErrUnsupportedContentType = errors.New("invalid content type")

	// ErrUnidentified indicates unidentified error.
	ErrUnidentified = errors.New("unidentified error")

	// ErrEmptyPath indicates empty file path.
	ErrEmptyPath = errors.New("empty file path")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")

	// ErrRollbackTx indicates failed to rollback transaction.
	ErrRollbackTx = errors.New("failed to rollback transaction")

	// ErrAuthentication indicates failure occurred while authenticating the entity.
	ErrAuthentication = NewAuthNError("failed to perform authentication over the entity")

	// ErrAuthorization indicates failure occurred while authorizing the entity.
	ErrAuthorization = NewAuthZError("failed to perform authorization over the entity")

	// ErrDomainAuthorization indicates failure occurred while authorizing the domain.
	ErrDomainAuthorization = NewAuthZError("failed to perform authorization over the domain")

	// ErrMissingDomainMember indicates member is not part of a domain.
	ErrMissingDomainMember = errors.New("member id is not member of domain")

	// ErrMissingMember indicates member is not found.
	ErrMissingMember = errors.New("member id is not found")
)

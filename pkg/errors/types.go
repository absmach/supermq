// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

var (
	// ErrMalformedEntity indicates a malformed entity specification.
	ErrMalformedEntity = New("malformed entity specification")

	// ErrUnsupportedContentType indicates invalid content type.
	ErrUnsupportedContentType = New("invalid content type")

	// ErrUnidentified indicates unidentified error.
	ErrUnidentified = New("unidentified error")

	// ErrEmptyPath indicates empty file path.
	ErrEmptyPath = New("empty file path")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = New("status already assigned")

	// ErrRollbackTx indicates failed to rollback transaction.
	ErrRollbackTx = New("failed to rollback transaction")

	// ErrAuthentication indicates failure occurred while authenticating the entity.
	ErrAuthentication = New("failed to perform authentication over the entity")

	// ErrAuthorization indicates failure occurred while authorizing the entity.
	ErrAuthorization = New("failed to perform authorization over the entity")

	// ErrMissingDomainMember indicates member is not part of a domain.
	ErrMissingDomainMember = New("member id is not member of domain")

	// ErrMissingMember indicates member is not found.
	ErrMissingMember = New("member id is not found")

	// ErrEmailAlreadyExists indicates that the email id already exists.
	ErrEmailAlreadyExists = New("email id already exists")

	// ErrUsernameNotAvailable indicates that the username is not available.
	ErrUsernameNotAvailable = New("username not available")

	// ErrDomainRouteNotAvailable indicates that the domain route is not available.
	ErrDomainRouteNotAvailable = New("domain route not available")

	// ErrChannelRouteNotAvailable indicates that the channel route is not available.
	ErrChannelRouteNotAvailable = New("channel route not available")

	// ErrRouteNotAvailable indicates that the username is not available.
	ErrRouteNotAvailable = New("route not available")
)

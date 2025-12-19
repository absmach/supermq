// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

var (
	// ErrMalformedEntity indicates a malformed entity specification.
	ErrMalformedEntity = NewRequestError("malformed entity specification")

	// ErrUnsupportedContentType indicates invalid content type.
	ErrUnsupportedContentType = New("invalid content type")

	// ErrUnidentified indicates unidentified error.
	ErrUnidentified = New("unidentified error")

	// ErrEmptyPath indicates empty file path.
	ErrEmptyPath = New("empty file path")

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
	ErrEmailAlreadyExists = New("email id already registered")

	// ErrUsernameNotAvailable indicates that the username is not available.
	ErrUsernameNotAvailable = New("username not available")

	// ErrDomainRouteNotAvailable indicates that the domain route is not available.
	ErrDomainRouteNotAvailable = New("domain route not available")

	// ErrChannelRouteNotAvailable indicates that the channel route is not available.
	ErrChannelRouteNotAvailable = New("channel route not available")

	// ErrTryAgain indicates to try the operation again.
	ErrTryAgain = New("Something went wrong, please try again")

	// ErrRouteNotAvailable indicates that the username is not available.
	ErrRouteNotAvailable = NewRequestError("route not available")
)

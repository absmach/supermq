// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

var (
	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = New("unsupported content type")
	// ErrInvalidQueryParams indicates invalid query parameters
	ErrInvalidQueryParams = New("invalid query parameters")
	// ErrNotInQuery indicates boolean parameter missing in the query
	ErrNotInQuery = New("missing in the query")
	// ErrMalformedEntity indicates a malformed entity specification
	ErrMalformedEntity = New("malformed entity specification")
)

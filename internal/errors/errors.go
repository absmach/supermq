// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

import "github.com/mainflux/mainflux/pkg/errors"

var (
	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = errors.New("unsupported content type")
	// ErrInvalidQueryParams indicates invalid query parameters
	ErrInvalidQueryParams = errors.New("invalid query parameters")
	// ErrNotInQuery indicates boolean parameter missing in the query
	ErrNotInQuery = errors.New("missing in the query")
	// ErrMalformedEntity indicates a malformed entity specification
	ErrMalformedEntity = errors.New("malformed entity specification")
)

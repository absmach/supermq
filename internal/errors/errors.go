// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import "github.com/mainflux/mainflux/pkg/errors"

var (
	// ErrUnsupportedContentType indicates unacceptable or lack of Content-Type
	ErrUnsupportedContentType = errors.New("unsupported content type")
	// ErrInvalidQueryParams indicates invalid query parameters
	ErrInvalidQueryParams = errors.New("invalid query parameters")
	// ErrNotInQuery indicates boolean parameter missing in the query
	ErrNotInQuery = errors.New("ignore parameter")
	// ErrMalformedEntity indicates malformed entity
	ErrMalformedEntity = errors.New("malformed entity specification")
)

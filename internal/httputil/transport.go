// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"context"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/mainflux/mainflux/logger"
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
)

// LoggingErrorEncoder is a go-kit error encoder logging decorator.
func LoggingErrorEncoder(logger logger.Logger, enc kithttp.ErrorEncoder) kithttp.ErrorEncoder {
	return func(ctx context.Context, err error, w http.ResponseWriter) {
		switch err {
		case ErrMissingToken,
			ErrMissingID,
			ErrMissingKey,
			ErrInvalidIDFormat,
			ErrNameSize,
			ErrLimitSize,
			ErrInvalidOrder,
			ErrInvalidDirection,
			ErrEmptyList,
			ErrMalformedPolicy:
			logger.Error(err.Error())
		}

		enc(ctx, err, w)
	}
}

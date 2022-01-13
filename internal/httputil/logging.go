// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"context"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/mainflux/mainflux/logger"
)

// Middleware is an ErrorEncoder middleware
type Middleware func(kithttp.ErrorEncoder) kithttp.ErrorEncoder

// LoggingErrorEncoder is a go-kit error encoder logging decorator.
func LoggingErrorEncoder(logger logger.Logger) Middleware {
	return func(encode kithttp.ErrorEncoder) kithttp.ErrorEncoder {
		return func(ctx context.Context, err error, w http.ResponseWriter) {
			logger.Error(err.Error())
			encode(ctx, err, w)
		}
	}
}

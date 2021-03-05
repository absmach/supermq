// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	internalerr "github.com/mainflux/mainflux/internal/errors"
	"github.com/mainflux/mainflux/readers"
)

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	chanID   string
	pageMeta readers.PageMetadata
}

func (req listMessagesReq) validate() error {
	if req.pageMeta.Limit < 1 || req.pageMeta.Offset < 0 {
		return internalerr.ErrInvalidQueryParams
	}
	if req.pageMeta.Comparator != "" &&
		req.pageMeta.Comparator != readers.EqualKey &&
		req.pageMeta.Comparator != readers.LowerThanKey &&
		req.pageMeta.Comparator != readers.LowerThanEqualKey &&
		req.pageMeta.Comparator != readers.GreaterThanKey &&
		req.pageMeta.Comparator != readers.GreaterThanEqualKey {
		return internalerr.ErrInvalidQueryParams
	}

	return nil
}

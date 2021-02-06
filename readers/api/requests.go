// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import "github.com/mainflux/mainflux/readers"

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	chanID   string
	pageMeta readers.PageMetadata
}

func (req listMessagesReq) validate() error {
	if req.pageMeta.Limit < 1 || req.pageMeta.Offset < 0 {
		return errInvalidRequest
	}
	if req.pageMeta.Comparison != "" &&
		req.pageMeta.Comparison != readers.EqualKey &&
		req.pageMeta.Comparison != readers.LowerThanKey &&
		req.pageMeta.Comparison != readers.LowerEqualThanKey &&
		req.pageMeta.Comparison != readers.GreaterThanKey &&
		req.pageMeta.Comparison != readers.GreaterEqualThanKey {
		return errInvalidRequest
	}

	return nil
}

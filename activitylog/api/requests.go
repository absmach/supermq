// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/internal/apiutil"
)

const maxLimitSize = 1000

type listActivitiesReq struct {
	token string
	page  activitylog.Page
}

func (req listActivitiesReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.page.Limit > maxLimitSize {
		return apiutil.ErrLimitSize
	}
	if req.page.Direction != "" && req.page.Direction != api.AscDir && req.page.Direction != api.DescDir {
		return apiutil.ErrInvalidDirection
	}

	return nil
}

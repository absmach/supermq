// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/stretchr/testify/assert"
)

var (
	token        = "token"
	limit uint64 = 10
)

func TestListActivitiesReqValidate(t *testing.T) {
	cases := []struct {
		desc string
		req  listActivitiesReq
		err  error
	}{
		{
			desc: "valid",
			req: listActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit: limit,
				},
			},
			err: nil,
		},
		{
			desc: "missing token",
			req: listActivitiesReq{
				page: activitylog.Page{
					Limit: limit,
				},
			},
			err: apiutil.ErrBearerToken,
		},
		{
			desc: "invalid limit size",
			req: listActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit: maxLimitSize + 1,
				},
			},
			err: apiutil.ErrLimitSize,
		},
		{
			desc: "invalid sorting direction",
			req: listActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit:     limit,
					Direction: "invalid",
				},
			},
			err: apiutil.ErrInvalidDirection,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.req.validate()
			assert.Equal(t, c.err, err)
		})
	}
}

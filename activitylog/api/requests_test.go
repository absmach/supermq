// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"testing"

	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/stretchr/testify/assert"
)

var (
	token        = "token"
	limit uint64 = 10
)

func TestRetrieveActivitiesReqValidate(t *testing.T) {
	cases := []struct {
		desc string
		req  retrieveActivitiesReq
		err  error
	}{
		{
			desc: "valid",
			req: retrieveActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit: limit,
				},
			},
			err: nil,
		},
		{
			desc: "missing token",
			req: retrieveActivitiesReq{
				page: activitylog.Page{
					Limit: limit,
				},
			},
			err: apiutil.ErrBearerToken,
		},
		{
			desc: "invalid limit size",
			req: retrieveActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit: api.DefLimit + 1,
				},
			},
			err: apiutil.ErrLimitSize,
		},
		{
			desc: "invalid sorting direction",
			req: retrieveActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit:     limit,
					Direction: "invalid",
				},
			},
			err: apiutil.ErrInvalidDirection,
		},
		{
			desc: "valid id and entity type",
			req: retrieveActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit:      limit,
					EntityID:   "id",
					EntityType: activitylog.UserEntity,
				},
			},
			err: nil,
		},
		{
			desc: "valid id and empty entity type",
			req: retrieveActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit:      limit,
					EntityID:   "id",
					EntityType: activitylog.EmptyEntity,
				},
			},
			err: apiutil.ErrMissingEntityType,
		},
		{
			desc: "empty id and empty entity type",
			req: retrieveActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit:      limit,
					EntityType: activitylog.EmptyEntity,
				},
			},
			err: nil,
		},
		{
			desc: "empty id and valid entity type",
			req: retrieveActivitiesReq{
				token: token,
				page: activitylog.Page{
					Limit:      limit,
					EntityType: activitylog.UserEntity,
				},
			},
			err: apiutil.ErrMissingID,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			err := c.req.validate()
			assert.Equal(t, c.err, err)
		})
	}
}

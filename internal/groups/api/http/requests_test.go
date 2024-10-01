// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"
	"strings"
	"testing"

	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/stretchr/testify/assert"
)

var valid = "valid"

func TestCreateGroupReqValidation(t *testing.T) {
	cases := []struct {
		desc string
		req  createGroupReq
		err  error
	}{
		{
			desc: "valid request",
			req: createGroupReq{
				token: valid,
				Group: groups.Group{
					Name: valid,
				},
			},
			err: nil,
		},
		{
			desc: "empty token",
			req: createGroupReq{
				Group: groups.Group{
					Name: valid,
				},
			},
			err: apiutil.ErrBearerToken,
		},
		{
			desc: "long name",
			req: createGroupReq{
				token: valid,
				Group: groups.Group{
					Name: strings.Repeat("a", api.MaxNameSize+1),
				},
			},
			err: apiutil.ErrNameSize,
		},
		{
			desc: "empty name",
			req: createGroupReq{
				token: valid,
				Group: groups.Group{},
			},
			err: apiutil.ErrNameSize,
		},
	}

	for _, tc := range cases {
		err := tc.req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestUpdateGroupReqValidation(t *testing.T) {
	cases := []struct {
		desc string
		req  updateGroupReq
		err  error
	}{
		{
			desc: "valid request",
			req: updateGroupReq{
				token: valid,
				id:    valid,
				Name:  valid,
			},
			err: nil,
		},
		{
			desc: "empty token",
			req: updateGroupReq{
				id:   valid,
				Name: valid,
			},
			err: apiutil.ErrBearerToken,
		},
		{
			desc: "long name",
			req: updateGroupReq{
				token: valid,
				id:    valid,
				Name:  strings.Repeat("a", api.MaxNameSize+1),
			},
			err: apiutil.ErrNameSize,
		},
		{
			desc: "empty id",
			req: updateGroupReq{
				token: valid,
				Name:  valid,
			},
			err: apiutil.ErrMissingID,
		},
	}

	for _, tc := range cases {
		err := tc.req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestListGroupReqValidation(t *testing.T) {
	cases := []struct {
		desc string
		req  listGroupsReq
		err  error
	}{
		{
			desc: "valid request",
			req: listGroupsReq{
				token: valid,
				PageMeta: groups.PageMeta{
					Limit: 10,
				},
			},
			err: nil,
		},
		{
			desc: "empty token",
			req: listGroupsReq{
				PageMeta: groups.PageMeta{
					Limit: 10,
				},
			},
			err: apiutil.ErrBearerToken,
		},

		{
			desc: "invalid upper level",
			req: listGroupsReq{
				token: valid,
				PageMeta: groups.PageMeta{
					Limit: 10,
				},
			},
			err: apiutil.ErrInvalidLevel,
		},
		{
			desc: "invalid lower limit",
			req: listGroupsReq{
				token: valid,
				PageMeta: groups.PageMeta{
					Limit: 0,
				},
			},
			err: apiutil.ErrLimitSize,
		},
		{
			desc: "invalid upper limit",
			req: listGroupsReq{
				token: valid,
				PageMeta: groups.PageMeta{
					Limit: api.MaxLimitSize + 1,
				},
			},
			err: apiutil.ErrLimitSize,
		},
	}

	for _, tc := range cases {
		err := tc.req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestGroupReqValidation(t *testing.T) {
	cases := []struct {
		desc string
		req  groupReq
		err  error
	}{
		{
			desc: "valid request",
			req: groupReq{
				token: valid,
				id:    valid,
			},
			err: nil,
		},
		{
			desc: "empty token",
			req: groupReq{
				id: valid,
			},
			err: apiutil.ErrBearerToken,
		},
		{
			desc: "empty id",
			req: groupReq{
				token: valid,
			},
			err: apiutil.ErrMissingID,
		},
	}

	for _, tc := range cases {
		err := tc.req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestChangeGroupStatusReqValidation(t *testing.T) {
	cases := []struct {
		desc string
		req  changeGroupStatusReq
		err  error
	}{
		{
			desc: "valid request",
			req: changeGroupStatusReq{
				token: valid,
				id:    valid,
			},
			err: nil,
		},
		{
			desc: "empty token",
			req: changeGroupStatusReq{
				id: valid,
			},
			err: apiutil.ErrBearerToken,
		},
		{
			desc: "empty id",
			req: changeGroupStatusReq{
				token: valid,
			},
			err: apiutil.ErrMissingID,
		},
	}

	for _, tc := range cases {
		err := tc.req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

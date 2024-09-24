// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/internal/testsutil"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/absmach/magistrala/pkg/groups/mocks"
	"github.com/stretchr/testify/assert"
)

var validGroupResp = groups.Group{
	ID:          testsutil.GenerateUUID(&testing.T{}),
	Name:        valid,
	Description: valid,
	Domain:      testsutil.GenerateUUID(&testing.T{}),
	Parent:      testsutil.GenerateUUID(&testing.T{}),
	Metadata: clients.Metadata{
		"name": "test",
	},
	Children:  []*groups.Group{},
	CreatedAt: time.Now().Add(-1 * time.Second),
	UpdatedAt: time.Now(),
	UpdatedBy: testsutil.GenerateUUID(&testing.T{}),
	Status:    clients.EnabledStatus,
}

func TestCreateGroupEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	cases := []struct {
		desc    string
		kind    string
		req     createGroupReq
		svcResp groups.Group
		svcErr  error
		resp    createGroupRes
		err     error
	}{
		{
			desc: "successfully with groups kind",
			kind: auth.NewGroupKind,
			req: createGroupReq{
				token: valid,
				Group: groups.Group{
					Name: valid,
				},
			},
			svcResp: validGroupResp,
			svcErr:  nil,
			resp:    createGroupRes{created: true, Group: validGroupResp},
			err:     nil,
		},
		{
			desc: "successfully with channels kind",
			kind: auth.NewChannelKind,
			req: createGroupReq{
				token: valid,
				Group: groups.Group{
					Name: valid,
				},
			},
			svcResp: validGroupResp,
			svcErr:  nil,
			resp:    createGroupRes{created: true, Group: validGroupResp},
			err:     nil,
		},
		{
			desc: "unsuccessfully with invalid request",
			kind: auth.NewGroupKind,
			req: createGroupReq{
				Group: groups.Group{
					Name: valid,
				},
			},
			resp: createGroupRes{created: false},
			err:  apiutil.ErrValidation,
		},
		{
			desc: "unsuccessfully with repo error",
			kind: auth.NewGroupKind,
			req: createGroupReq{
				token: valid,
				Group: groups.Group{
					Name: valid,
				},
			},
			svcResp: groups.Group{},
			svcErr:  svcerr.ErrAuthorization,
			resp:    createGroupRes{created: false},
			err:     svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := svc.On("CreateGroup", context.Background(), tc.req.token, tc.kind, tc.req.Group).Return(tc.svcResp, tc.svcErr)
		resp, err := CreateGroupEndpoint(svc, tc.kind)(context.Background(), tc.req)
		assert.Equal(t, tc.resp, resp, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.resp, resp))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("expected error %v to contain %v", err, tc.err))
		response := resp.(createGroupRes)
		switch err {
		case nil:
			assert.Equal(t, response.Code(), http.StatusCreated)
			assert.Equal(t, response.Headers()["Location"], fmt.Sprintf("/groups/%s", response.ID))
		default:
			assert.Equal(t, response.Code(), http.StatusOK)
			assert.Empty(t, response.Headers())
		}
		assert.False(t, response.Empty())
		repoCall.Unset()
	}
}

func TestViewGroupEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	cases := []struct {
		desc    string
		req     groupReq
		svcResp groups.Group
		svcErr  error
		resp    viewGroupRes
		err     error
	}{
		{
			desc: "successfully",
			req: groupReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcResp: validGroupResp,
			svcErr:  nil,
			resp:    viewGroupRes{Group: validGroupResp},
			err:     nil,
		},
		{
			desc: "unsuccessfully with invalid request",
			req: groupReq{
				id: testsutil.GenerateUUID(t),
			},
			resp: viewGroupRes{},
			err:  apiutil.ErrValidation,
		},
		{
			desc: "unsuccessfully with repo error",
			req: groupReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcResp: groups.Group{},
			svcErr:  svcerr.ErrAuthorization,
			resp:    viewGroupRes{},
			err:     svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := svc.On("ViewGroup", context.Background(), tc.req.token, tc.req.id).Return(tc.svcResp, tc.svcErr)
		resp, err := ViewGroupEndpoint(svc)(context.Background(), tc.req)
		assert.Equal(t, tc.resp, resp, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.resp, resp))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("expected error %v to contain %v", err, tc.err))
		response := resp.(viewGroupRes)
		assert.Equal(t, response.Code(), http.StatusOK)
		assert.Empty(t, response.Headers())
		assert.False(t, response.Empty())
		repoCall.Unset()
	}
}

func TestEnableGroupEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	cases := []struct {
		desc    string
		req     changeGroupStatusReq
		svcResp groups.Group
		svcErr  error
		resp    changeStatusRes
		err     error
	}{
		{
			desc: "successfully",
			req: changeGroupStatusReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcResp: validGroupResp,
			svcErr:  nil,
			resp:    changeStatusRes{Group: validGroupResp},
			err:     nil,
		},
		{
			desc: "unsuccessfully with invalid request",
			req: changeGroupStatusReq{
				id: testsutil.GenerateUUID(t),
			},
			resp: changeStatusRes{},
			err:  apiutil.ErrValidation,
		},
		{
			desc: "unsuccessfully with repo error",
			req: changeGroupStatusReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcResp: groups.Group{},
			svcErr:  svcerr.ErrAuthorization,
			resp:    changeStatusRes{},
			err:     svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := svc.On("EnableGroup", context.Background(), tc.req.token, tc.req.id).Return(tc.svcResp, tc.svcErr)
		resp, err := EnableGroupEndpoint(svc)(context.Background(), tc.req)
		assert.Equal(t, tc.resp, resp, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.resp, resp))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("expected error %v to contain %v", err, tc.err))
		response := resp.(changeStatusRes)
		assert.Equal(t, response.Code(), http.StatusOK)
		assert.Empty(t, response.Headers())
		assert.False(t, response.Empty())
		repoCall.Unset()
	}
}

func TestDisableGroupEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	cases := []struct {
		desc    string
		req     changeGroupStatusReq
		svcResp groups.Group
		svcErr  error
		resp    changeStatusRes
		err     error
	}{
		{
			desc: "successfully",
			req: changeGroupStatusReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcResp: validGroupResp,
			svcErr:  nil,
			resp:    changeStatusRes{Group: validGroupResp},
			err:     nil,
		},
		{
			desc: "unsuccessfully with invalid request",
			req: changeGroupStatusReq{
				id: testsutil.GenerateUUID(t),
			},
			resp: changeStatusRes{},
			err:  apiutil.ErrValidation,
		},
		{
			desc: "unsuccessfully with repo error",
			req: changeGroupStatusReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcResp: groups.Group{},
			svcErr:  svcerr.ErrAuthorization,
			resp:    changeStatusRes{},
			err:     svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := svc.On("DisableGroup", context.Background(), tc.req.token, tc.req.id).Return(tc.svcResp, tc.svcErr)
		resp, err := DisableGroupEndpoint(svc)(context.Background(), tc.req)
		assert.Equal(t, tc.resp, resp, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.resp, resp))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("expected error %v to contain %v", err, tc.err))
		response := resp.(changeStatusRes)
		assert.Equal(t, response.Code(), http.StatusOK)
		assert.Empty(t, response.Headers())
		assert.False(t, response.Empty())
		repoCall.Unset()
	}
}

func TestDeleteGroupEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	cases := []struct {
		desc   string
		req    groupReq
		svcErr error
		resp   deleteGroupRes
		err    error
	}{
		{
			desc: "successfully",
			req: groupReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcErr: nil,
			resp:   deleteGroupRes{deleted: true},
			err:    nil,
		},
		{
			desc: "unsuccessfully with invalid request",
			req: groupReq{
				id: testsutil.GenerateUUID(t),
			},
			resp: deleteGroupRes{},
			err:  apiutil.ErrValidation,
		},
		{
			desc: "unsuccessfully with repo error",
			req: groupReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
			},
			svcErr: svcerr.ErrAuthorization,
			resp:   deleteGroupRes{},
			err:    svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := svc.On("DeleteGroup", context.Background(), tc.req.token, tc.req.id).Return(tc.svcErr)
		resp, err := DeleteGroupEndpoint(svc)(context.Background(), tc.req)
		assert.Equal(t, tc.resp, resp, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.resp, resp))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("expected error %v to contain %v", err, tc.err))
		response := resp.(deleteGroupRes)
		switch err {
		case nil:
			assert.Equal(t, response.Code(), http.StatusNoContent)
		default:
			assert.Equal(t, response.Code(), http.StatusBadRequest)
		}
		assert.Empty(t, response.Headers())
		assert.True(t, response.Empty())
		repoCall.Unset()
	}
}

func TestUpdateGroupEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	cases := []struct {
		desc    string
		req     updateGroupReq
		svcResp groups.Group
		svcErr  error
		resp    updateGroupRes
		err     error
	}{
		{
			desc: "successfully",
			req: updateGroupReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
				Name:  valid,
			},
			svcResp: validGroupResp,
			svcErr:  nil,
			resp:    updateGroupRes{Group: validGroupResp},
			err:     nil,
		},
		{
			desc: "unsuccessfully with invalid request",
			req: updateGroupReq{
				id:   testsutil.GenerateUUID(t),
				Name: valid,
			},
			resp: updateGroupRes{},
			err:  apiutil.ErrValidation,
		},
		{
			desc: "unsuccessfully with repo error",
			req: updateGroupReq{
				token: valid,
				id:    testsutil.GenerateUUID(t),
				Name:  valid,
			},
			svcResp: groups.Group{},
			svcErr:  svcerr.ErrAuthorization,
			resp:    updateGroupRes{},
			err:     svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		group := groups.Group{
			ID:          tc.req.id,
			Name:        tc.req.Name,
			Description: tc.req.Description,
			Metadata:    tc.req.Metadata,
		}
		repoCall := svc.On("UpdateGroup", context.Background(), tc.req.token, group).Return(tc.svcResp, tc.svcErr)
		resp, err := UpdateGroupEndpoint(svc)(context.Background(), tc.req)
		assert.Equal(t, tc.resp, resp, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.resp, resp))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("expected error %v to contain %v", err, tc.err))
		response := resp.(updateGroupRes)
		assert.Equal(t, response.Code(), http.StatusOK)
		assert.Empty(t, response.Headers())
		assert.False(t, response.Empty())
		repoCall.Unset()
	}
}

func TestListGroupsEndpoint(t *testing.T) {
	svc := new(mocks.Service)
	childGroup := groups.Group{
		ID:          testsutil.GenerateUUID(t),
		Name:        valid,
		Description: valid,
		Domain:      testsutil.GenerateUUID(t),
		Parent:      validGroupResp.ID,
		Metadata: clients.Metadata{
			"name": "test",
		},
		Level:     -1,
		Children:  []*groups.Group{},
		CreatedAt: time.Now().Add(-1 * time.Second),
		UpdatedAt: time.Now(),
		UpdatedBy: testsutil.GenerateUUID(t),
		Status:    clients.EnabledStatus,
	}
	parentGroup := groups.Group{
		ID:          testsutil.GenerateUUID(t),
		Name:        valid,
		Description: valid,
		Domain:      testsutil.GenerateUUID(t),
		Metadata: clients.Metadata{
			"name": "test",
		},
		Level:     1,
		Children:  []*groups.Group{},
		CreatedAt: time.Now().Add(-1 * time.Second),
		UpdatedAt: time.Now(),
		UpdatedBy: testsutil.GenerateUUID(t),
		Status:    clients.EnabledStatus,
	}

	validGroupResp.Children = append(validGroupResp.Children, &childGroup)
	parentGroup.Children = append(parentGroup.Children, &validGroupResp)

	cases := []struct {
		desc       string
		memberKind string
		req        listGroupsReq
		svcResp    groups.Page
		svcErr     error
		resp       groupPageRes
		err        error
	}{
		{
			desc:       "successfully",
			memberKind: auth.ThingsKind,
			req: listGroupsReq{
				Page: groups.Page{
					PageMeta: groups.PageMeta{
						Limit: 10,
					},
				},
				token: valid,
			},
			svcResp: groups.Page{
				Groups: []groups.Group{validGroupResp},
			},
			svcErr: nil,
			resp: groupPageRes{
				Groups: []viewGroupRes{
					{
						Group: validGroupResp,
					},
				},
			},
			err: nil,
		},
		{
			desc: "successfully with empty member kind",
			req: listGroupsReq{
				Page: groups.Page{
					PageMeta: groups.PageMeta{
						Limit: 10,
					},
				},
				token: valid,
			},
			svcResp: groups.Page{
				Groups: []groups.Group{validGroupResp},
			},
			svcErr: nil,
			resp: groupPageRes{
				Groups: []viewGroupRes{
					{
						Group: validGroupResp,
					},
				},
			},
			err: nil,
		},
		{
			desc:       "successfully with tree",
			memberKind: auth.ThingsKind,
			req: listGroupsReq{
				Page: groups.Page{
					PageMeta: groups.PageMeta{
						Limit: 10,
					},
					Tree: true,
				},
				token: valid,
			},
			svcResp: groups.Page{
				Groups: []groups.Group{validGroupResp, childGroup},
			},
			svcErr: nil,
			resp: groupPageRes{
				Groups: []viewGroupRes{
					{
						Group: validGroupResp,
					},
				},
			},
			err: nil,
		},
		{
			desc:       "list children groups successfully without tree",
			memberKind: auth.UsersKind,
			req: listGroupsReq{
				Page: groups.Page{
					PageMeta: groups.PageMeta{
						Limit: 10,
					},
					ParentID:           validGroupResp.ID,
					HierarchyDirection: -1,
					Tree:               false,
				},
				token: valid,
			},
			svcResp: groups.Page{
				Groups: []groups.Group{validGroupResp, childGroup},
			},
			svcErr: nil,
			resp: groupPageRes{
				Groups: []viewGroupRes{
					{
						Group: childGroup,
					},
				},
			},
			err: nil,
		},
		{
			desc:       "list parent group successfully without tree",
			memberKind: auth.UsersKind,
			req: listGroupsReq{
				Page: groups.Page{
					PageMeta: groups.PageMeta{
						Limit: 10,
					},
					ParentID:           validGroupResp.ID,
					HierarchyDirection: 1,
					Tree:               false,
				},
				token: valid,
			},
			svcResp: groups.Page{
				Groups: []groups.Group{parentGroup, validGroupResp},
			},
			svcErr: nil,
			resp: groupPageRes{
				Groups: []viewGroupRes{
					{
						Group: parentGroup,
					},
				},
			},
			err: nil,
		},
		{
			desc:       "unsuccessfully with invalid request",
			memberKind: auth.ThingsKind,
			req:        listGroupsReq{},
			resp:       groupPageRes{},
			err:        apiutil.ErrValidation,
		},
		{
			desc:       "unsuccessfully with repo error",
			memberKind: auth.ThingsKind,
			req: listGroupsReq{
				Page: groups.Page{
					PageMeta: groups.PageMeta{
						Limit: 10,
					},
				},
				token: valid,
			},
			svcResp: groups.Page{},
			svcErr:  svcerr.ErrAuthorization,
			resp:    groupPageRes{},
			err:     svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {

		repoCall := svc.On("ListGroups", context.Background(), tc.req.token, tc.req.Page).Return(tc.svcResp, tc.svcErr)
		resp, err := ListGroupsEndpoint(svc)(context.Background(), tc.req)
		assert.Equal(t, tc.resp, resp, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.resp, resp))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("expected error %v to contain %v", err, tc.err))
		response := resp.(groupPageRes)
		assert.Equal(t, response.Code(), http.StatusOK)
		assert.Empty(t, response.Headers())
		assert.False(t, response.Empty())
		repoCall.Unset()
	}
}

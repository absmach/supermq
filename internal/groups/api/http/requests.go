// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/groups"
	mggroups "github.com/absmach/magistrala/pkg/groups"
)

type createGroupReq struct {
	mggroups.Group
	token string
}

func (req createGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if len(req.Name) > api.MaxNameSize || req.Name == "" {
		return apiutil.ErrNameSize
	}

	return nil
}

type updateGroupReq struct {
	token       string
	id          string
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if len(req.Name) > api.MaxNameSize {
		return apiutil.ErrNameSize
	}
	return nil
}

type listGroupsReq struct {
	mggroups.PageMeta
	token string
}

func (req listGroupsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.Limit > api.MaxLimitSize || req.Limit < 1 {
		return apiutil.ErrLimitSize
	}

	return nil
}

type groupReq struct {
	token string
	id    string
}

func (req groupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type changeGroupStatusReq struct {
	token string
	id    string
}

func (req changeGroupStatusReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type retrieveGroupHierarchyReq struct {
	mggroups.HierarchyPageMeta
	id    string
	token string
}

func (req retrieveGroupHierarchyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.Level > groups.MaxLevel {
		return apiutil.ErrLevel
	}

	return nil
}

type addParentGroupReq struct {
	token    string
	id       string
	ParentID string `json:"parent_id"`
}

func (req addParentGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if err := api.ValidateUUID(req.ParentID); err != nil {
		return err
	}
	if req.id == req.ParentID {
		return apiutil.ErrSelfParentingNotAllowed
	}
	return nil
}

type removeParentGroupReq struct {
	token string
	id    string
}

func (req removeParentGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type viewParentGroupReq struct {
	token string
	id    string
}

func (req viewParentGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type addChildrenGroupsReq struct {
	token       string
	id          string
	ChildrenIDs []string `json:"children_ids"`
}

func (req addChildrenGroupsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if len(req.ChildrenIDs) == 0 {
		return apiutil.ErrMissingChildrenGroupIDs
	}
	for _, childID := range req.ChildrenIDs {
		if err := api.ValidateUUID(childID); err != nil {
			return err
		}
		if req.id == childID {
			return apiutil.ErrSelfParentingNotAllowed
		}
	}
	return nil
}

type removeChildrenGroupsReq struct {
	token       string
	id          string
	ChildrenIDs []string `json:"children_ids"`
}

func (req removeChildrenGroupsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if len(req.ChildrenIDs) == 0 {
		return apiutil.ErrMissingChildrenGroupIDs
	}
	for _, childID := range req.ChildrenIDs {
		if err := api.ValidateUUID(childID); err != nil {
			return err
		}
	}
	return nil
}

type removeAllChildrenGroupsReq struct {
	token string
	id    string
}

func (req removeAllChildrenGroupsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type listChildrenGroupsReq struct {
	token string
	id    string
	mggroups.PageMeta
}

func (req listChildrenGroupsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if req.Limit > api.MaxLimitSize || req.Limit < 1 {
		return apiutil.ErrLimitSize
	}
	return nil
}

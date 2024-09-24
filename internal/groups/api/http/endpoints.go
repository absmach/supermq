// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/go-kit/kit/endpoint"
)

const groupTypeChannels = "channels"

func CreateGroupEndpoint(svc groups.Service, kind string) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createGroupReq)
		if err := req.validate(); err != nil {
			return createGroupRes{created: false}, errors.Wrap(apiutil.ErrValidation, err)
		}

		group, err := svc.CreateGroup(ctx, req.token, kind, req.Group)
		if err != nil {
			return createGroupRes{created: false}, err
		}

		return createGroupRes{created: true, Group: group}, nil
	}
}

func ViewGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return viewGroupRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}

		group, err := svc.ViewGroup(ctx, req.token, req.id)
		if err != nil {
			return viewGroupRes{}, err
		}

		return viewGroupRes{Group: group}, nil
	}
}

func UpdateGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateGroupReq)
		if err := req.validate(); err != nil {
			return updateGroupRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}

		group := groups.Group{
			ID:          req.id,
			Name:        req.Name,
			Description: req.Description,
			Metadata:    req.Metadata,
		}

		group, err := svc.UpdateGroup(ctx, req.token, group)
		if err != nil {
			return updateGroupRes{}, err
		}

		return updateGroupRes{Group: group}, nil
	}
}

func EnableGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeGroupStatusReq)
		if err := req.validate(); err != nil {
			return changeStatusRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		group, err := svc.EnableGroup(ctx, req.token, req.id)
		if err != nil {
			return changeStatusRes{}, err
		}
		return changeStatusRes{Group: group}, nil
	}
}

func DisableGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeGroupStatusReq)
		if err := req.validate(); err != nil {
			return changeStatusRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		group, err := svc.DisableGroup(ctx, req.token, req.id)
		if err != nil {
			return changeStatusRes{}, err
		}
		return changeStatusRes{Group: group}, nil
	}
}

func ListGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)

		if err := req.validate(); err != nil {
			return groupPageRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		page, err := svc.ListGroups(ctx, req.token, req.Page)
		if err != nil {
			return groupPageRes{}, err
		}

		if req.Page.Tree {
			return buildGroupsResponseTree(page), nil
		}
		filterByID := req.Page.ParentID != ""

		return buildGroupsResponse(page, filterByID), nil
	}
}

func DeleteGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return deleteGroupRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.DeleteGroup(ctx, req.token, req.id); err != nil {
			return deleteGroupRes{}, err
		}
		return deleteGroupRes{deleted: true}, nil
	}
}

func listParentGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listParentGroupsReq)
		if err := req.validate(); err != nil {
			return listParentGroupsRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		gp, err := svc.ListParentGroups(ctx, req.token, req.id, req.Page)
		if err != nil {
			return listParentGroupsRes{}, err
		}
		viewGroups := []viewGroupRes{}

		for _, group := range gp.Groups {
			viewGroups = append(viewGroups, toViewGroupRes(group))
		}
		return listParentGroupsRes{
			pageRes: pageRes{
				Limit:  gp.Limit,
				Offset: gp.Offset,
				Total:  gp.Total,
				Level:  gp.Level,
			},
			Groups: viewGroups,
		}, nil
	}
}

func addParentGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addParentGroupReq)
		if err := req.validate(); err != nil {
			return addParentGroupRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.AddParentGroup(ctx, req.token, req.id, req.ParentID); err != nil {
			return addParentGroupRes{}, err
		}
		return addParentGroupRes{}, nil
	}
}

func removeParentGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeParentGroupReq)
		if err := req.validate(); err != nil {
			return removeParentGroupRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.RemoveParentGroup(ctx, req.token, req.id); err != nil {
			return removeParentGroupRes{}, err
		}
		return removeParentGroupRes{}, nil
	}
}
func viewParentGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewParentGroupReq)
		if err := req.validate(); err != nil {
			return viewParentGroupRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		pg, err := svc.ViewParentGroup(ctx, req.token, req.id)
		if err != nil {
			return viewParentGroupRes{}, err
		}
		return viewParentGroupRes{pg}, nil
	}
}

func addChildrenGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addChildrenGroupsReq)
		if err := req.validate(); err != nil {
			return addChildrenGroupsRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.AddChildrenGroups(ctx, req.token, req.id, req.ChildrenIDs); err != nil {
			return addChildrenGroupsRes{}, err
		}
		return addChildrenGroupsRes{}, nil
	}
}

func removeChildrenGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeChildrenGroupsReq)
		if err := req.validate(); err != nil {
			return removeChildrenGroupsReq{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.RemoveChildrenGroups(ctx, req.token, req.id, req.ChildrenIDs); err != nil {
			return removeChildrenGroupsReq{}, err
		}
		return removeChildrenGroupsReq{}, nil
	}
}

func removeAllChildrenGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeAllChildrenGroupsReq)
		if err := req.validate(); err != nil {
			return removeChildrenGroupsReq{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		if err := svc.RemoveAllChildrenGroups(ctx, req.token, req.id); err != nil {
			return removeAllChildrenGroupsReq{}, err
		}
		return removeAllChildrenGroupsReq{}, nil
	}
}

func listChildrenGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listChildrenGroupsReq)
		if err := req.validate(); err != nil {
			return listChildrenGroupsRes{}, errors.Wrap(apiutil.ErrValidation, err)
		}
		gp, err := svc.ListChildrenGroups(ctx, req.token, req.id, req.Page)
		if err != nil {
			return listChildrenGroupsRes{}, err
		}
		viewGroups := []viewGroupRes{}

		for _, group := range gp.Groups {
			viewGroups = append(viewGroups, toViewGroupRes(group))
		}
		return listChildrenGroupsRes{
			pageRes: pageRes{
				Limit:  gp.Limit,
				Offset: gp.Offset,
				Total:  gp.Total,
				Level:  gp.Level,
			},
			Groups: viewGroups,
		}, nil
	}
}

func buildGroupsResponseTree(page groups.Page) groupPageRes {
	groupsMap := map[string]*groups.Group{}
	// Parents' map keeps its array of children.
	parentsMap := map[string][]*groups.Group{}
	for i := range page.Groups {
		if _, ok := groupsMap[page.Groups[i].ID]; !ok {
			groupsMap[page.Groups[i].ID] = &page.Groups[i]
			parentsMap[page.Groups[i].ID] = make([]*groups.Group, 0)
		}
	}

	for _, group := range groupsMap {
		if children, ok := parentsMap[group.Parent]; ok {
			children = append(children, group)
			parentsMap[group.Parent] = children
		}
	}

	res := groupPageRes{
		pageRes: pageRes{
			Limit:  page.Limit,
			Offset: page.Offset,
			Total:  page.Total,
			Level:  page.Level,
		},
		Groups: []viewGroupRes{},
	}

	for _, group := range groupsMap {
		if children, ok := parentsMap[group.ID]; ok {
			group.Children = children
		}
	}

	for _, group := range groupsMap {
		view := toViewGroupRes(*group)
		if children, ok := parentsMap[group.Parent]; len(children) == 0 || !ok {
			res.Groups = append(res.Groups, view)
		}
	}

	return res
}

func toViewGroupRes(group groups.Group) viewGroupRes {
	view := viewGroupRes{
		Group: group,
	}
	return view
}

func buildGroupsResponse(gp groups.Page, filterByID bool) groupPageRes {
	res := groupPageRes{
		pageRes: pageRes{
			Total: gp.Total,
			Level: gp.Level,
		},
		Groups: []viewGroupRes{},
	}

	for _, group := range gp.Groups {
		view := viewGroupRes{
			Group: group,
		}
		if filterByID && group.Level == 0 {
			continue
		}
		res.Groups = append(res.Groups, view)
	}

	return res
}

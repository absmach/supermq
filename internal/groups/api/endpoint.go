package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/internal/groups"
)

func createGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createGroupReq)
		if err := req.validate(); err != nil {
			return groupRes{}, err
		}

		group := groups.Group{
			Name:        req.Name,
			Description: req.Description,
			ParentID:    req.ParentID,
			Metadata:    req.Metadata,
		}

		group, err := svc.CreateGroup(ctx, req.token, group)
		if err != nil {
			return groupRes{}, err
		}

		return groupRes{created: true, id: group.ID}, nil
	}
}

func viewGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return viewGroupRes{}, err
		}

		group, err := svc.ViewGroup(ctx, req.token, req.id)
		if err != nil {
			return viewGroupRes{}, err
		}

		res := viewGroupRes{
			ID:          group.ID,
			Name:        group.Name,
			Description: group.Description,
			Metadata:    group.Metadata,
			ParentID:    group.ParentID,
			OwnerID:     group.OwnerID,
			CreatedAt:   group.CreatedAt,
			UpdatedAt:   group.UpdatedAt,
		}

		return res, nil
	}
}

func updateGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateGroupReq)
		if err := req.validate(); err != nil {
			return groupRes{}, err
		}

		group := groups.Group{
			ID:          req.id,
			Name:        req.Name,
			Description: req.Description,
			Metadata:    req.Metadata,
		}

		_, err := svc.UpdateGroup(ctx, req.token, group)
		if err != nil {
			return groupRes{}, err
		}

		res := groupRes{created: false}
		return res, nil
	}
}

func deleteGroupEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(groupReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.RemoveGroup(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return deleteRes{}, nil
	}
}

func listGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}
		pm := groups.PageMetadata{
			Level:    req.level,
			Metadata: req.metadata,
		}
		page, err := svc.ListGroups(ctx, req.token, pm)
		if err != nil {
			return groupPageRes{}, err
		}

		if req.tree {
			return buildGroupsResponseTree(page), nil
		}

		return buildGroupsResponse(page), nil
	}
}

func listMemberships(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMembershipsReq)
		if err := req.validate(); err != nil {
			return memberPageRes{}, err
		}

		pm := groups.PageMetadata{
			Offset:   req.offset,
			Limit:    req.limit,
			Metadata: req.metadata,
		}

		page, err := svc.ListMemberships(ctx, req.token, req.id, pm)
		if err != nil {
			return memberPageRes{}, err
		}

		return buildGroupsResponse(page), nil
	}
}

func listChildrenEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}

		pm := groups.PageMetadata{
			Level:    req.level,
			Metadata: req.metadata,
		}
		page, err := svc.ListChildren(ctx, req.token, req.id, pm)
		if err != nil {
			return groupPageRes{}, err
		}

		if req.tree {
			return buildGroupsResponseTree(page), nil
		}

		return buildGroupsResponse(page), nil
	}
}

func listParentsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		if err := req.validate(); err != nil {
			return groupPageRes{}, err
		}
		pm := groups.PageMetadata{
			Level:    req.level,
			Metadata: req.metadata,
		}

		page, err := svc.ListParents(ctx, req.token, req.id, pm)
		if err != nil {
			return groupPageRes{}, err
		}

		if req.tree {
			return buildGroupsResponseTree(page), nil
		}

		return buildGroupsResponse(page), nil
	}
}

func assignEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.Assign(ctx, req.token, req.groupID, req.Members...); err != nil {
			return nil, err
		}

		return assignRes{}, nil
	}
}

func unassignEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(unassignReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.Unassign(ctx, req.token, req.groupID, req.Members...); err != nil {
			return nil, err
		}

		return unassignRes{}, nil
	}
}

func buildGroupsResponseTree(page groups.GroupPage) groupPageRes {
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
		if children, ok := parentsMap[group.ParentID]; ok {
			children = append(children, group)
			parentsMap[group.ParentID] = children
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
		if children, ok := parentsMap[group.ParentID]; len(children) == 0 || !ok {
			res.Groups = append(res.Groups, view)
		}
	}

	return res
}

func toViewGroupRes(group groups.Group) viewGroupRes {
	view := viewGroupRes{
		ID:          group.ID,
		ParentID:    group.ParentID,
		OwnerID:     group.OwnerID,
		Name:        group.Name,
		Description: group.Description,
		Metadata:    group.Metadata,
		Level:       group.Level,
		Path:        group.Path,
		Children:    make([]*viewGroupRes, 0),
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}

	for _, ch := range group.Children {
		child := toViewGroupRes(*ch)
		view.Children = append(view.Children, &child)
	}

	return view
}

func buildGroupsResponse(gp groups.GroupPage) groupPageRes {
	res := groupPageRes{
		pageRes: pageRes{
			Total: gp.Total,
			Level: gp.Level,
		},
		Groups: []viewGroupRes{},
	}

	for _, group := range gp.Groups {
		view := viewGroupRes{
			ID:          group.ID,
			ParentID:    group.ParentID,
			OwnerID:     group.OwnerID,
			Name:        group.Name,
			Description: group.Description,
			Metadata:    group.Metadata,
			Level:       group.Level,
			Path:        group.Path,
			CreatedAt:   group.CreatedAt,
			UpdatedAt:   group.UpdatedAt,
		}
		res.Groups = append(res.Groups, view)
	}

	return res
}

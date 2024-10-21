// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/authn"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/absmach/magistrala/pkg/policies"
	"github.com/absmach/magistrala/things"
	"github.com/go-kit/kit/endpoint"
)

func createClientEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createClientReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		client, err := svc.CreateThings(ctx, session, req.client)
		if err != nil {
			return nil, err
		}

		return createClientRes{
			Client:  client[0],
			created: true,
		}, nil
	}
}

func createClientsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createClientsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}

		session.UpdateSession(req.domainID)
		page, err := svc.CreateThings(ctx, session, req.Clients...)
		if err != nil {
			return nil, err
		}

		res := clientsPageRes{
			pageRes: pageRes{
				Total: uint64(len(page)),
			},
			Clients: []viewClientRes{},
		}
		for _, c := range page {
			res.Clients = append(res.Clients, viewClientRes{Client: c})
		}

		return res, nil
	}
}

func viewClientEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewClientReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}

		session.UpdateSession(req.domainID)
		c, err := svc.ViewClient(ctx, session, req.id)
		if err != nil {
			return nil, err
		}

		return viewClientRes{Client: c}, nil
	}
}

func viewClientPermsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewClientPermsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}

		session.UpdateSession(req.domainID)
		p, err := svc.ViewClientPerms(ctx, session, req.id)
		if err != nil {
			return nil, err
		}

		return viewClientPermsRes{Permissions: p}, nil
	}
}

func listClientsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listClientsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		pm := mgclients.Page{
			Status:     req.status,
			Offset:     req.offset,
			Limit:      req.limit,
			Name:       req.name,
			Tag:        req.tag,
			Permission: req.permission,
			Metadata:   req.metadata,
			ListPerms:  req.listPerms,
			Role:       mgclients.AllRole, // retrieve all things since things don't have roles
			Id:         req.id,
		}
		page, err := svc.ListClients(ctx, session, req.userID, pm)
		if err != nil {
			return nil, err
		}

		res := clientsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Clients: []viewClientRes{},
		}
		for _, c := range page.Clients {
			res.Clients = append(res.Clients, viewClientRes{Client: c})
		}

		return res, nil
	}
}

func listMembersEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listMembersReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		req.Page.Role = mgclients.AllRole // retrieve all things since things don't have roles
		page, err := svc.ListClientsByGroup(ctx, session, req.groupID, req.Page)
		if err != nil {
			return nil, err
		}

		return buildClientsResponse(page), nil
	}
}

func updateClientEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		cli := mgclients.Client{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		client, err := svc.UpdateClient(ctx, session, cli)
		if err != nil {
			return nil, err
		}

		return updateClientRes{Client: client}, nil
	}
}

func updateClientTagsEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientTagsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		cli := mgclients.Client{
			ID:   req.id,
			Tags: req.Tags,
		}
		client, err := svc.UpdateClientTags(ctx, session, cli)
		if err != nil {
			return nil, err
		}

		return updateClientRes{Client: client}, nil
	}
}

func updateClientSecretEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClientCredentialsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		client, err := svc.UpdateClientSecret(ctx, session, req.id, req.Secret)
		if err != nil {
			return nil, err
		}

		return updateClientRes{Client: client}, nil
	}
}

func enableClientEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeClientStatusReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		client, err := svc.EnableClient(ctx, session, req.id)
		if err != nil {
			return nil, err
		}

		return changeClientStatusRes{Client: client}, nil
	}
}

func disableClientEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(changeClientStatusReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		client, err := svc.DisableClient(ctx, session, req.id)
		if err != nil {
			return nil, err
		}

		return changeClientStatusRes{Client: client}, nil
	}
}

func buildClientsResponse(cp mgclients.MembersPage) clientsPageRes {
	res := clientsPageRes{
		pageRes: pageRes{
			Total:  cp.Total,
			Offset: cp.Offset,
			Limit:  cp.Limit,
		},
		Clients: []viewClientRes{},
	}
	for _, c := range cp.Members {
		res.Clients = append(res.Clients, viewClientRes{Client: c})
	}

	return res
}

func assignUsersEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignUsersRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Assign(ctx, session, req.groupID, req.Relation, policies.UsersKind, req.UserIDs...); err != nil {
			return nil, err
		}

		return assignUsersRes{}, nil
	}
}

func unassignUsersEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignUsersRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Unassign(ctx, session, req.groupID, req.Relation, policies.UsersKind, req.UserIDs...); err != nil {
			return nil, err
		}

		return unassignUsersRes{}, nil
	}
}

func assignUserGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignUserGroupsRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Assign(ctx, session, req.groupID, policies.ParentGroupRelation, policies.ChannelsKind, req.UserGroupIDs...); err != nil {
			return nil, err
		}

		return assignUserGroupsRes{}, nil
	}
}

func unassignUserGroupsEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(assignUserGroupsRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Unassign(ctx, session, req.groupID, policies.ParentGroupRelation, policies.ChannelsKind, req.UserGroupIDs...); err != nil {
			return nil, err
		}

		return unassignUserGroupsRes{}, nil
	}
}

func connectChannelThingEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(connectChannelThingRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Assign(ctx, session, req.ChannelID, policies.GroupRelation, policies.ThingsKind, req.ThingID); err != nil {
			return nil, err
		}

		return connectChannelThingRes{}, nil
	}
}

func disconnectChannelThingEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(connectChannelThingRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Unassign(ctx, session, req.ChannelID, policies.GroupRelation, policies.ThingsKind, req.ThingID); err != nil {
			return nil, err
		}

		return disconnectChannelThingRes{}, nil
	}
}

func connectEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(connectChannelThingRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Assign(ctx, session, req.ChannelID, policies.GroupRelation, policies.ThingsKind, req.ThingID); err != nil {
			return nil, err
		}

		return connectChannelThingRes{}, nil
	}
}

func disconnectEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(connectChannelThingRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Unassign(ctx, session, req.ChannelID, policies.GroupRelation, policies.ThingsKind, req.ThingID); err != nil {
			return nil, err
		}

		return disconnectChannelThingRes{}, nil
	}
}

func thingShareEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(thingShareRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Share(ctx, session, req.thingID, req.Relation, req.UserIDs...); err != nil {
			return nil, err
		}

		return thingShareRes{}, nil
	}
}

func thingUnshareEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(thingShareRequest)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.Unshare(ctx, session, req.thingID, req.Relation, req.UserIDs...); err != nil {
			return nil, err
		}

		return thingUnshareRes{}, nil
	}
}

func deleteClientEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteClientReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}
		session.UpdateSession(req.domainID)
		if err := svc.DeleteClient(ctx, session, req.id); err != nil {
			return nil, err
		}

		return deleteClientRes{}, nil
	}
}

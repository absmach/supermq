// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/ui"
)

func indexEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(indexReq)
		res, err := svc.Index(ctx, req.token)
		return uiRes{
			html: res,
		}, err
	}
}

func createThingEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createThingsReq)

		th := sdk.Thing{
			Key:      req.Key,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		res, err := svc.CreateThings(ctx, req.token, th)
		if err != nil {
			return nil, err
		}

		return uiRes{
			html: res,
		}, err
	}
}

func viewThingEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		res, err := svc.ViewThing(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html: res,
		}, err
	}
}

func listThingsEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listThingsReq)
		res, err := svc.ListThings(ctx, req.token)
		return uiRes{
			html: res,
		}, err
	}
}

func updateThingEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateThingReq)
		uth := sdk.Thing{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}

		res, err := svc.UpdateThing(ctx, req.token, req.id, uth)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html: res,
		}, err
	}
}

func removeThingEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		res, err := svc.RemoveThing(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html:    res,
			headers: map[string]string{"location": redirectURL + "things"},
			code:    http.StatusPermanentRedirect,
		}, err
	}
}

func createChannelEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createChannelsReq)

		ch := sdk.Channel{
			Key:      req.Key,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		res, err := svc.CreateChannels(ctx, "123", ch)
		if err != nil {
			return nil, err
		}

		return uiRes{
			html: res,
		}, err
	}
}

func viewChannelEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		res, err := svc.ViewChannel(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html: res,
		}, err
	}
}

func updateChannelEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateChannelReq)

		uch := sdk.Channel{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}

		res, err := svc.UpdateChannel(ctx, req.token, req.id, uch)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html: res,
		}, err
	}
}

func listChannelsEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listChannelsReq)
		res, err := svc.ListChannels(ctx, req.token)
		return uiRes{
			html: res,
		}, err
	}
}

func removeChannelEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		res, err := svc.RemoveChannel(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html:    res,
			headers: map[string]string{"location": redirectURL + "channels"},
			code:    http.StatusPermanentRedirect,
		}, err
	}
}

func createGroupEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createGroupsReq)

		gr := sdk.Group{
			Name:        req.Name,
			Description: req.Description,
			ParentID:    req.ParentID,
			Metadata:    req.Metadata,
		}

		res, err := svc.CreateGroups(ctx, req.token, gr)
		if err != nil {
			return nil, err
		}

		return uiRes{
			html: res,
		}, err
	}
}

func listGroupsEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listGroupsReq)
		res, err := svc.ListGroups(ctx, req.token)
		return uiRes{
			html: res,
		}, err
	}
}

func viewGroupEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		res, err := svc.ViewGroup(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}

		return uiRes{
			html: res,
		}, err
	}
}

func updateGroupEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateGroupReq)

		uch := sdk.Group{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}

		res, err := svc.UpdateGroup(ctx, req.token, req.id, uch)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html: res,
		}, err
	}
}

func removeGroupEndpoint(svc ui.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewResourceReq)

		res, err := svc.RemoveGroup(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}
		return uiRes{
			html:    res,
			headers: map[string]string{"location": redirectURL + "groups"},
			code:    http.StatusPermanentRedirect,
		}, err
	}
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/absmach/magistrala/things"
	pThings "github.com/absmach/magistrala/things/private"
	"github.com/go-kit/kit/endpoint"
)

func authorizeEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(authorizeReq)

		thingID, err := svc.Authorize(ctx, things.AuthzReq{
			ChannelID:  req.ChannelID,
			ThingID:    req.ThingID,
			ThingKey:   req.ThingKey,
			Permission: req.Permission,
		})
		if err != nil {
			return authorizeRes{}, err
		}
		return authorizeRes{
			authorized: true,
			id:         thingID,
		}, err
	}
}

func retrieveEntityEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(retrieveEntityReq)
		thing, err := svc.RetrieveById(ctx, req.Id)

		if err != nil {
			return retrieveEntityRes{}, err
		}

		return retrieveEntityRes{id: thing.ID, domain: thing.Domain, status: uint8(thing.Status)}, nil

	}
}
func retrieveEntitiesEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(retrieveEntitiesReq)
		tp, err := svc.RetrieveByIds(ctx, req.Ids)

		if err != nil {
			return retrieveEntitiesRes{}, err
		}
		thingsBasic := []thingBasic{}
		for _, thing := range tp.Clients {
			thingsBasic = append(thingsBasic, thingBasic{id: thing.ID, domain: thing.Domain, status: uint8(thing.Status)})
		}
		return retrieveEntitiesRes{
			total:  tp.Total,
			limit:  tp.Limit,
			offset: tp.Offset,
			things: thingsBasic,
		}, nil

	}
}

func addConnectionsEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(connectionsReq)

		var conns []things.Connection

		for _, c := range req.connections {
			conns = append(conns, things.Connection{
				ThingID:   c.thingID,
				ChannelID: c.channelID,
				DomainID:  c.domainID,
			})
		}
		if err := svc.AddConnections(ctx, conns); err != nil {
			return connectionsRes{ok: false}, err
		}

		return connectionsRes{ok: true}, nil

	}
}

func removeConnectionsEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(connectionsReq)

		var conns []things.Connection

		for _, c := range req.connections {
			conns = append(conns, things.Connection{
				ThingID:   c.thingID,
				ChannelID: c.channelID,
				DomainID:  c.domainID,
			})
		}
		if err := svc.RemoveConnections(ctx, conns); err != nil {
			return connectionsRes{ok: false}, err
		}

		return connectionsRes{ok: true}, nil

	}
}
func removeChannelConnectionsEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeChannelConnectionsReq)

		if err := svc.RemoveChannelConnections(ctx, req.channelID); err != nil {
			return removeChannelConnectionsRes{}, err
		}

		return removeChannelConnectionsRes{}, nil
	}
}

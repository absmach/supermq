// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/absmach/magistrala/things"
	"github.com/go-kit/kit/endpoint"
)

func authorizeEndpoint(svc things.Service) endpoint.Endpoint {
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

func getEntityBasicEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(getEntityBasicReq)
		thing, err := svc.RetrieveById(ctx, req.Id)

		if err != nil {
			return thingBasic{}, err
		}

		return thingBasic{id: thing.ID, domain: thing.Domain, status: uint8(thing.Status)}, nil

	}
}
func getEntitiesBasicEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {

		req := request.(getEntitiesBasicReq)
		tp, err := svc.RetrieveByIds(ctx, req.Ids)

		if err != nil {
			return getEntitiesBasicRes{}, err
		}
		thingsBasic := []thingBasic{}
		for _, thing := range tp.Clients {
			thingsBasic = append(thingsBasic, thingBasic{id: thing.ID, domain: thing.Domain, status: uint8(thing.Status)})
		}
		return getEntitiesBasicRes{
			total:  tp.Total,
			limit:  tp.Limit,
			offset: tp.Offset,
			things: thingsBasic,
		}, nil

	}
}

func addConnectionsEndpoint(svc things.Service) endpoint.Endpoint {
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
		err := svc.AddConnections(ctx, conns)

		if err != nil {
			return connectionsRes{ok: false}, err
		}

		return connectionsRes{ok: true}, nil

	}
}

func removeConnectionsEndpoint(svc things.Service) endpoint.Endpoint {
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
		err := svc.RemoveConnections(ctx, conns)

		if err != nil {
			return connectionsRes{ok: false}, err
		}

		return connectionsRes{ok: true}, nil

	}
}

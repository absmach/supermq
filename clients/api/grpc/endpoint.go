// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/absmach/magistrala/clients"
	pThings "github.com/absmach/magistrala/clients/private"
	"github.com/go-kit/kit/endpoint"
)

func authenticateEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(authenticateReq)

		thingID, err := svc.Authenticate(ctx, req.ThingKey)
		if err != nil {
			return authenticateRes{}, err
		}
		return authenticateRes{
			authenticated: true,
			id:            thingID,
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
		thingsBasic := []enitity{}
		for _, client := range tp.Clients {
			thingsBasic = append(thingsBasic, enitity{id: client.ID, domain: client.Domain, status: uint8(client.Status)})
		}
		return retrieveEntitiesRes{
			total:   tp.Total,
			limit:   tp.Limit,
			offset:  tp.Offset,
			clients: thingsBasic,
		}, nil
	}
}

func addConnectionsEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(connectionsReq)

		var conns []clients.Connection

		for _, c := range req.connections {
			conns = append(conns, clients.Connection{
				ThingID:   c.clientID,
				ChannelID: c.channelID,
				DomainID:  c.domainID,
				Type:      c.connType,
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

		var conns []clients.Connection

		for _, c := range req.connections {
			conns = append(conns, clients.Connection{
				ThingID:   c.clientID,
				ChannelID: c.channelID,
				DomainID:  c.domainID,
				Type:      c.connType,
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

func UnsetParentGroupFromClientEndpoint(svc pThings.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UnsetParentGroupFromClientReq)

		if err := svc.UnsetParentGroupFromClient(ctx, req.parentGroupID); err != nil {
			return UnsetParentGroupFromClientRes{}, err
		}

		return UnsetParentGroupFromClientRes{}, nil
	}
}

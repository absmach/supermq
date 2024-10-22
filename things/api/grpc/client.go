// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/things"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const svcName = "magistrala.ThingsService"

var _ magistrala.ThingsServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	timeout           time.Duration
	authorize         endpoint.Endpoint
	getEntityBasic    endpoint.Endpoint
	getEntitiesBasic  endpoint.Endpoint
	addConnections    endpoint.Endpoint
	removeConnections endpoint.Endpoint
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn, timeout time.Duration) magistrala.ThingsServiceClient {
	return &grpcClient{
		authorize: kitgrpc.NewClient(
			conn,
			svcName,
			"Authorize",
			encodeAuthorizeRequest,
			decodeAuthorizeResponse,
			magistrala.ThingsAuthzRes{},
		).Endpoint(),

		getEntityBasic: kitgrpc.NewClient(
			conn,
			svcName,
			"GetEntityBasic",
			encodeGetEntityBasicRequest,
			decodeGetEntityBasicResponse,
			magistrala.EntityBasic{},
		).Endpoint(),

		getEntitiesBasic: kitgrpc.NewClient(
			conn,
			svcName,
			"GetEntitiesBasic",
			encodeGetEntitiesBasicRequest,
			decodeGetEntitiesBasicResponse,
			magistrala.EntitiesBasicRes{},
		).Endpoint(),

		addConnections: kitgrpc.NewClient(
			conn,
			svcName,
			"AddConnections",
			encodeConnectionsRequest,
			decodeConnectionsResponse,
			magistrala.ConnectionsRes{},
		).Endpoint(),

		removeConnections: kitgrpc.NewClient(
			conn,
			svcName,
			"RemoveConnections",
			encodeConnectionsRequest,
			decodeConnectionsResponse,
			magistrala.ConnectionsRes{},
		).Endpoint(),
		timeout: timeout,
	}
}

func (client grpcClient) Authorize(ctx context.Context, req *magistrala.ThingsAuthzReq, _ ...grpc.CallOption) (r *magistrala.ThingsAuthzRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.authorize(ctx, things.AuthzReq{
		ThingID:    req.GetThingID(),
		ThingKey:   req.GetThingKey(),
		ChannelID:  req.GetChannelID(),
		Permission: req.GetPermission(),
	})
	if err != nil {
		return &magistrala.ThingsAuthzRes{}, decodeError(err)
	}

	ar := res.(authorizeRes)
	return &magistrala.ThingsAuthzRes{Authorized: ar.authorized, Id: ar.id}, nil
}

func (client grpcClient) GetEntityBasic(ctx context.Context, req *magistrala.EntityReq, _ ...grpc.CallOption) (r *magistrala.EntityBasic, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.getEntityBasic(ctx, req.GetId())
	if err != nil {
		return &magistrala.EntityBasic{}, decodeError(err)
	}

	thing := res.(thingBasic)

	return &magistrala.EntityBasic{Id: thing.id, DomainId: thing.domain, Status: uint32(thing.status)}, nil
}

func (client grpcClient) GetEntitiesBasic(ctx context.Context, req *magistrala.EntitiesReq, _ ...grpc.CallOption) (r *magistrala.EntitiesBasicRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.getEntitiesBasic(ctx, req.GetIds())
	if err != nil {
		return &magistrala.EntitiesBasicRes{}, decodeError(err)
	}

	ep := res.(getEntitiesBasicRes)

	entities := []*magistrala.EntityBasic{}
	for _, thing := range ep.things {
		entities = append(entities, &magistrala.EntityBasic{
			Id:       thing.id,
			DomainId: thing.domain,
			Status:   uint32(thing.status),
		})
	}
	return &magistrala.EntitiesBasicRes{Total: ep.total, Limit: ep.limit, Offset: ep.offset, Entities: entities}, nil
}

func (client grpcClient) AddConnections(ctx context.Context, req *magistrala.ConnectionsReq, _ ...grpc.CallOption) (r *magistrala.ConnectionsRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()
	conns := []things.Connection{}

	for _, c := range req.Connections {
		conns = append(conns, things.Connection{
			ThingID:   c.GetThingId(),
			ChannelID: c.GetChannelId(),
			DomainID:  c.GetDomainId(),
		})
	}

	res, err := client.addConnections(ctx, conns)
	if err != nil {
		return &magistrala.ConnectionsRes{}, decodeError(err)
	}

	cr := res.(connectionsRes)

	return &magistrala.ConnectionsRes{Ok: cr.ok}, nil
}

func (client grpcClient) RemoveConnections(ctx context.Context, req *magistrala.ConnectionsReq, _ ...grpc.CallOption) (r *magistrala.ConnectionsRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()
	conns := []things.Connection{}

	for _, c := range req.Connections {
		conns = append(conns, things.Connection{
			ThingID:   c.GetThingId(),
			ChannelID: c.GetChannelId(),
			DomainID:  c.GetDomainId(),
		})
	}

	res, err := client.removeConnections(ctx, conns)
	if err != nil {
		return &magistrala.ConnectionsRes{}, decodeError(err)
	}

	cr := res.(connectionsRes)

	return &magistrala.ConnectionsRes{Ok: cr.ok}, nil
}

func decodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*magistrala.ThingsAuthzRes)
	return authorizeRes{authorized: res.Authorized, id: res.Id}, nil
}

func encodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(things.AuthzReq)
	return &magistrala.ThingsAuthzReq{
		ChannelID:  req.ChannelID,
		ThingID:    req.ThingID,
		ThingKey:   req.ThingKey,
		Permission: req.Permission,
	}, nil
}

func decodeGetEntityBasicResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*magistrala.EntityBasic)

	return thingBasic{
		id:     res.GetId(),
		domain: res.GetDomainId(),
		status: uint8(res.GetStatus()),
	}, nil
}

func encodeGetEntityBasicRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(string)
	return &magistrala.EntityReq{
		Id: req,
	}, nil
}

func decodeGetEntitiesBasicResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*magistrala.EntitiesBasicRes)

	ths := []thingBasic{}

	for _, entity := range res.Entities {
		ths = append(ths, thingBasic{
			id:     entity.GetId(),
			domain: entity.GetDomainId(),
			status: uint8(entity.GetStatus()),
		})
	}
	return getEntitiesBasicRes{total: res.GetTotal(), limit: res.GetLimit(), offset: res.GetOffset(), things: ths}, nil
}

func encodeGetEntitiesBasicRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.([]string)
	return &magistrala.EntitiesReq{
		Ids: req,
	}, nil
}

func decodeConnectionsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*magistrala.ConnectionsRes)

	return connectionsRes{ok: res.GetOk()}, nil
}

func encodeConnectionsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.([]things.Connection)

	conns := []*magistrala.Connection{}

	for _, r := range req {
		conns = append(conns, &magistrala.Connection{
			ThingId:   r.ThingID,
			ChannelId: r.ChannelID,
			DomainId:  r.DomainID,
		})
	}
	return &magistrala.ConnectionsReq{
		Connections: conns,
	}, nil
}
func decodeError(err error) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unauthenticated:
			return errors.Wrap(svcerr.ErrAuthentication, errors.New(st.Message()))
		case codes.PermissionDenied:
			return errors.Wrap(svcerr.ErrAuthorization, errors.New(st.Message()))
		case codes.InvalidArgument:
			return errors.Wrap(errors.ErrMalformedEntity, errors.New(st.Message()))
		case codes.FailedPrecondition:
			return errors.Wrap(errors.ErrMalformedEntity, errors.New(st.Message()))
		case codes.NotFound:
			return errors.Wrap(svcerr.ErrNotFound, errors.New(st.Message()))
		case codes.AlreadyExists:
			return errors.Wrap(svcerr.ErrConflict, errors.New(st.Message()))
		case codes.OK:
			if msg := st.Message(); msg != "" {
				return errors.Wrap(errors.ErrUnidentified, errors.New(msg))
			}
			return nil
		default:
			return errors.Wrap(fmt.Errorf("unexpected gRPC status: %s (status code:%v)", st.Code().String(), st.Code()), errors.New(st.Message()))
		}
	}
	return err
}

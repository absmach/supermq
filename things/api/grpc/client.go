// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"fmt"
	"time"

	grpcCommonV1 "github.com/absmach/magistrala/internal/grpc/common/v1"
	grpcThingsV1 "github.com/absmach/magistrala/internal/grpc/things/v1"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/things"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const svcName = "things.v1.ThingsService"

var _ grpcThingsV1.ThingsServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	timeout                    time.Duration
	authorize                  endpoint.Endpoint
	retrieveEntity             endpoint.Endpoint
	retrieveEntities           endpoint.Endpoint
	addConnections             endpoint.Endpoint
	removeConnections          endpoint.Endpoint
	removeChannelConnections   endpoint.Endpoint
	unsetParentGroupFormThings endpoint.Endpoint
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn, timeout time.Duration) grpcThingsV1.ThingsServiceClient {
	return &grpcClient{
		authorize: kitgrpc.NewClient(
			conn,
			svcName,
			"Authorize",
			encodeAuthorizeRequest,
			decodeAuthorizeResponse,
			grpcThingsV1.AuthzRes{},
		).Endpoint(),

		retrieveEntity: kitgrpc.NewClient(
			conn,
			svcName,
			"RetrieveEntity",
			encodeRetrieveEntityRequest,
			decodeRetrieveEntityResponse,
			grpcCommonV1.RetrieveEntityRes{},
		).Endpoint(),

		retrieveEntities: kitgrpc.NewClient(
			conn,
			svcName,
			"RetrieveEntities",
			encodeRetrieveEntitiesRequest,
			decodeRetrieveEntitiesResponse,
			grpcCommonV1.RetrieveEntitiesRes{},
		).Endpoint(),

		addConnections: kitgrpc.NewClient(
			conn,
			svcName,
			"AddConnections",
			encodeAddConnectionsRequest,
			decodeAddConnectionsResponse,
			grpcCommonV1.AddConnectionsRes{},
		).Endpoint(),

		removeConnections: kitgrpc.NewClient(
			conn,
			svcName,
			"RemoveConnections",
			encodeRemoveConnectionsRequest,
			decodeRemoveConnectionsResponse,
			grpcCommonV1.RemoveConnectionsRes{},
		).Endpoint(),

		removeChannelConnections: kitgrpc.NewClient(
			conn,
			svcName,
			"RemoveChannelConnections",
			encodeRemoveChannelConnectionsRequest,
			decodeRemoveChannelConnectionsResponse,
			grpcThingsV1.RemoveChannelConnectionsRes{},
		).Endpoint(),

		unsetParentGroupFormThings: kitgrpc.NewClient(
			conn,
			svcName,
			"UnsetParentGroupFormThings",
			encodeUnsetParentGroupFormThingsRequest,
			decodeUnsetParentGroupFormThingsResponse,
			grpcThingsV1.UnsetParentGroupFormThingsRes{},
		).Endpoint(),

		timeout: timeout,
	}
}

func (client grpcClient) Authorize(ctx context.Context, req *grpcThingsV1.AuthzReq, _ ...grpc.CallOption) (r *grpcThingsV1.AuthzRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.authorize(ctx, things.AuthzReq{
		ThingID:    req.GetThingId(),
		ThingKey:   req.GetThingKey(),
		ChannelID:  req.GetChannelId(),
		Permission: req.GetPermission(),
	})
	if err != nil {
		return &grpcThingsV1.AuthzRes{}, decodeError(err)
	}

	ar := res.(authorizeRes)
	return &grpcThingsV1.AuthzRes{Authorized: ar.authorized, Id: ar.id}, nil
}

func encodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(things.AuthzReq)
	return &grpcThingsV1.AuthzReq{
		ChannelId:  req.ChannelID,
		ThingId:    req.ThingID,
		ThingKey:   req.ThingKey,
		Permission: req.Permission,
	}, nil
}

func decodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*grpcThingsV1.AuthzRes)
	return authorizeRes{authorized: res.Authorized, id: res.Id}, nil
}

func (client grpcClient) RetrieveEntity(ctx context.Context, req *grpcCommonV1.RetrieveEntityReq, _ ...grpc.CallOption) (r *grpcCommonV1.RetrieveEntityRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.retrieveEntity(ctx, req.GetId())
	if err != nil {
		return &grpcCommonV1.RetrieveEntityRes{}, decodeError(err)
	}

	ebr := res.(retrieveEntityRes)

	return &grpcCommonV1.RetrieveEntityRes{Entity: &grpcCommonV1.EntityBasic{Id: ebr.id, DomainId: ebr.domain, Status: uint32(ebr.status)}}, nil
}

func encodeRetrieveEntityRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(string)
	return &grpcCommonV1.RetrieveEntityReq{
		Id: req,
	}, nil
}

func decodeRetrieveEntityResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*grpcCommonV1.RetrieveEntityRes)

	return retrieveEntityRes{
		id:     res.Entity.GetId(),
		domain: res.Entity.GetDomainId(),
		status: uint8(res.Entity.GetStatus()),
	}, nil
}

func (client grpcClient) RetrieveEntities(ctx context.Context, req *grpcCommonV1.RetrieveEntitiesReq, _ ...grpc.CallOption) (r *grpcCommonV1.RetrieveEntitiesRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.retrieveEntities(ctx, req.GetIds())
	if err != nil {
		return &grpcCommonV1.RetrieveEntitiesRes{}, decodeError(err)
	}

	ep := res.(retrieveEntitiesRes)

	entities := []*grpcCommonV1.EntityBasic{}
	for _, thing := range ep.things {
		entities = append(entities, &grpcCommonV1.EntityBasic{
			Id:       thing.id,
			DomainId: thing.domain,
			Status:   uint32(thing.status),
		})
	}
	return &grpcCommonV1.RetrieveEntitiesRes{Total: ep.total, Limit: ep.limit, Offset: ep.offset, Entities: entities}, nil
}

func encodeRetrieveEntitiesRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.([]string)
	return &grpcCommonV1.RetrieveEntitiesReq{
		Ids: req,
	}, nil
}

func decodeRetrieveEntitiesResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*grpcCommonV1.RetrieveEntitiesRes)

	ths := []thingBasic{}

	for _, entity := range res.Entities {
		ths = append(ths, thingBasic{
			id:     entity.GetId(),
			domain: entity.GetDomainId(),
			status: uint8(entity.GetStatus()),
		})
	}
	return retrieveEntitiesRes{total: res.GetTotal(), limit: res.GetLimit(), offset: res.GetOffset(), things: ths}, nil
}

func (client grpcClient) AddConnections(ctx context.Context, req *grpcCommonV1.AddConnectionsReq, _ ...grpc.CallOption) (r *grpcCommonV1.AddConnectionsRes, err error) {
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
		return &grpcCommonV1.AddConnectionsRes{}, decodeError(err)
	}

	cr := res.(connectionsRes)

	return &grpcCommonV1.AddConnectionsRes{Ok: cr.ok}, nil
}

func encodeAddConnectionsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.([]things.Connection)

	conns := []*grpcCommonV1.Connection{}

	for _, r := range req {
		conns = append(conns, &grpcCommonV1.Connection{
			ThingId:   r.ThingID,
			ChannelId: r.ChannelID,
			DomainId:  r.DomainID,
		})
	}
	return &grpcCommonV1.AddConnectionsReq{
		Connections: conns,
	}, nil
}

func decodeAddConnectionsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*grpcCommonV1.AddConnectionsRes)

	return connectionsRes{ok: res.GetOk()}, nil
}

func (client grpcClient) RemoveConnections(ctx context.Context, req *grpcCommonV1.RemoveConnectionsReq, _ ...grpc.CallOption) (r *grpcCommonV1.RemoveConnectionsRes, err error) {
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
		return &grpcCommonV1.RemoveConnectionsRes{}, decodeError(err)
	}

	cr := res.(connectionsRes)

	return &grpcCommonV1.RemoveConnectionsRes{Ok: cr.ok}, nil
}

func encodeRemoveConnectionsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.([]things.Connection)

	conns := []*grpcCommonV1.Connection{}

	for _, r := range req {
		conns = append(conns, &grpcCommonV1.Connection{
			ThingId:   r.ThingID,
			ChannelId: r.ChannelID,
			DomainId:  r.DomainID,
		})
	}
	return &grpcCommonV1.RemoveConnectionsReq{
		Connections: conns,
	}, nil
}

func decodeRemoveConnectionsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*grpcCommonV1.RemoveConnectionsRes)

	return connectionsRes{ok: res.GetOk()}, nil
}
func (client grpcClient) RemoveChannelConnections(ctx context.Context, req *grpcThingsV1.RemoveChannelConnectionsReq, _ ...grpc.CallOption) (r *grpcThingsV1.RemoveChannelConnectionsRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	if _, err := client.removeChannelConnections(ctx, req); err != nil {
		return &grpcThingsV1.RemoveChannelConnectionsRes{}, decodeError(err)
	}

	return &grpcThingsV1.RemoveChannelConnectionsRes{}, nil
}

func encodeRemoveChannelConnectionsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq.(*grpcThingsV1.RemoveChannelConnectionsReq), nil
}

func decodeRemoveChannelConnectionsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	return grpcRes.(*grpcThingsV1.RemoveChannelConnectionsRes), nil
}

func (client grpcClient) UnsetParentGroupFormThings(ctx context.Context, req *grpcThingsV1.UnsetParentGroupFormThingsReq, _ ...grpc.CallOption) (r *grpcThingsV1.UnsetParentGroupFormThingsRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	if _, err := client.unsetParentGroupFormThings(ctx, req); err != nil {
		return &grpcThingsV1.UnsetParentGroupFormThingsRes{}, decodeError(err)
	}

	return &grpcThingsV1.UnsetParentGroupFormThingsRes{}, nil
}

func encodeUnsetParentGroupFormThingsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq.(*grpcThingsV1.UnsetParentGroupFormThingsReq), nil
}

func decodeUnsetParentGroupFormThingsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	return grpcRes.(*grpcThingsV1.UnsetParentGroupFormThingsRes), nil
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

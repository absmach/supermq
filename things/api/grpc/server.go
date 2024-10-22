// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/absmach/magistrala"
	mgauth "github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/things"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ magistrala.ThingsServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	magistrala.UnimplementedThingsServiceServer
	authorize         kitgrpc.Handler
	getEntityBasic    kitgrpc.Handler
	getEntitiesBasic  kitgrpc.Handler
	addConnections    kitgrpc.Handler
	removeConnections kitgrpc.Handler
}

// NewServer returns new AuthServiceServer instance.
func NewServer(svc things.Service) magistrala.ThingsServiceServer {
	return &grpcServer{
		authorize: kitgrpc.NewServer(
			(authorizeEndpoint(svc)),
			decodeAuthorizeRequest,
			encodeAuthorizeResponse,
		),
		getEntityBasic: kitgrpc.NewServer(
			(getEntityBasicEndpoint(svc)),
			decodeGetEntityBasicRequest,
			encodeGetEntityBasicResponse,
		),
		getEntitiesBasic: kitgrpc.NewServer(
			(getEntitiesBasicEndpoint(svc)),
			decodeGetEntitiesBasicRequest,
			encodeGetEntitiesBasicResponse,
		),
		addConnections: kitgrpc.NewServer(
			(addConnectionsEndpoint(svc)),
			decodeConnectionsRequest,
			encodeConnectionsResponse,
		),
		removeConnections: kitgrpc.NewServer(
			(removeConnectionsEndpoint(svc)),
			decodeConnectionsRequest,
			encodeConnectionsResponse,
		),
	}
}

func (s *grpcServer) Authorize(ctx context.Context, req *magistrala.ThingsAuthzReq) (*magistrala.ThingsAuthzRes, error) {
	_, res, err := s.authorize.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*magistrala.ThingsAuthzRes), nil
}

func decodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*magistrala.ThingsAuthzReq)
	return authorizeReq{
		ThingID:    req.GetThingID(),
		ThingKey:   req.GetThingKey(),
		ChannelID:  req.GetChannelID(),
		Permission: req.GetPermission(),
	}, nil
}

func encodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(authorizeRes)
	return &magistrala.ThingsAuthzRes{Authorized: res.authorized, Id: res.id}, nil
}

func (s *grpcServer) GetEntityBasic(ctx context.Context, req *magistrala.EntityReq) (*magistrala.EntityBasic, error) {
	_, res, err := s.getEntityBasic.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*magistrala.EntityBasic), nil
}

func decodeGetEntityBasicRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*magistrala.EntityReq)
	return getEntityBasicReq{
		Id: req.GetId(),
	}, nil
}

func encodeGetEntityBasicResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(thingBasic)

	return &magistrala.EntityBasic{
		Id:       res.id,
		DomainId: res.domain,
		Status:   uint32(res.status),
	}, nil
}

func (s *grpcServer) GetEntitiesBasic(ctx context.Context, req *magistrala.EntitiesReq) (*magistrala.EntitiesBasicRes, error) {
	_, res, err := s.getEntitiesBasic.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*magistrala.EntitiesBasicRes), nil
}

func decodeGetEntitiesBasicRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*magistrala.EntitiesReq)
	return getEntitiesBasicReq{
		Ids: req.GetIds(),
	}, nil
}

func encodeGetEntitiesBasicResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(getEntitiesBasicRes)

	entities := []*magistrala.EntityBasic{}
	for _, thing := range res.things {
		entities = append(entities, &magistrala.EntityBasic{
			Id:       thing.id,
			DomainId: thing.domain,
			Status:   uint32(thing.status),
		})
	}
	return &magistrala.EntitiesBasicRes{Total: res.total, Limit: res.limit, Offset: res.offset, Entities: entities}, nil
}

func (s *grpcServer) AddConnections(ctx context.Context, req *magistrala.ConnectionsReq) (*magistrala.ConnectionsRes, error) {
	_, res, err := s.addConnections.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*magistrala.ConnectionsRes), nil
}

func (s *grpcServer) RemoveConnections(ctx context.Context, req *magistrala.ConnectionsReq) (*magistrala.ConnectionsRes, error) {
	_, res, err := s.removeConnections.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*magistrala.ConnectionsRes), nil
}

func decodeConnectionsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*magistrala.ConnectionsReq)

	conns := []connection{}
	for _, c := range req.Connections {
		conns = append(conns, connection{
			thingID:   c.GetThingId(),
			channelID: c.GetChannelId(),
			domainID:  c.GetDomainId(),
		})
	}
	return connectionsReq{
		connections: conns,
	}, nil
}

func encodeConnectionsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(connectionsRes)

	return &magistrala.ConnectionsRes{Ok: res.ok}, nil
}

func encodeError(err error) error {
	switch {
	case errors.Contains(err, nil):
		return nil
	case errors.Contains(err, errors.ErrMalformedEntity),
		err == apiutil.ErrInvalidAuthKey,
		err == apiutil.ErrMissingID,
		err == apiutil.ErrMissingMemberType,
		err == apiutil.ErrMissingPolicySub,
		err == apiutil.ErrMissingPolicyObj,
		err == apiutil.ErrMalformedPolicyAct:
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Contains(err, svcerr.ErrAuthentication),
		errors.Contains(err, mgauth.ErrKeyExpired),
		err == apiutil.ErrMissingEmail,
		err == apiutil.ErrBearerToken:
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Contains(err, svcerr.ErrAuthorization):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

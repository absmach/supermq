// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	mgauth "github.com/absmach/magistrala/auth"
	channels "github.com/absmach/magistrala/channels/private"
	grpcChannelsV1 "github.com/absmach/magistrala/internal/grpc/channels/v1"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ grpcChannelsV1.ChannelsServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	grpcChannelsV1.UnimplementedChannelsServiceServer

	removeThingConnections       kitgrpc.Handler
	unsetParentGroupFormChannels kitgrpc.Handler
}

// NewServer returns new AuthServiceServer instance.
func NewServer(svc channels.Service) grpcChannelsV1.ChannelsServiceServer {
	return &grpcServer{
		removeThingConnections: kitgrpc.NewServer(
			removeThingConnectionsEndpoint(svc),
			decodeRemoveThingConnectionsRequest,
			encodeRemoveThingConnectionsResponse,
		),
		unsetParentGroupFormChannels: kitgrpc.NewServer(
			unsetParentGroupFormChannelsEndpoint(svc),
			decodeUnsetParentGroupFormChannelsRequest,
			encodeUnsetParentGroupFormChannelsResponse,
		),
	}
}

func (s *grpcServer) RemoveThingConnections(ctx context.Context, req *grpcChannelsV1.RemoveThingConnectionsReq) (*grpcChannelsV1.RemoveThingConnectionsRes, error) {
	_, res, err := s.removeThingConnections.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*grpcChannelsV1.RemoveThingConnectionsRes), nil
}

func decodeRemoveThingConnectionsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcChannelsV1.RemoveThingConnectionsReq)

	return removeThingConnectionsReq{
		thingID: req.GetThingId(),
	}, nil
}

func encodeRemoveThingConnectionsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	_ = grpcRes.(removeThingConnectionsRes)
	return &grpcChannelsV1.RemoveThingConnectionsRes{}, nil
}

func (s *grpcServer) UnsetParentGroupFormChannels(ctx context.Context, req *grpcChannelsV1.UnsetParentGroupFormChannelsReq) (*grpcChannelsV1.UnsetParentGroupFormChannelsRes, error) {
	_, res, err := s.unsetParentGroupFormChannels.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*grpcChannelsV1.UnsetParentGroupFormChannelsRes), nil
}

func decodeUnsetParentGroupFormChannelsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcChannelsV1.UnsetParentGroupFormChannelsReq)

	return unsetParentGroupFormChannelsReq{
		parentGroupID: req.GetParentGroupId(),
	}, nil
}

func encodeUnsetParentGroupFormChannelsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	_ = grpcRes.(unsetParentGroupFormChannelsRes)
	return &grpcChannelsV1.UnsetParentGroupFormChannelsRes{}, nil
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

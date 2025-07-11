// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	grpcapi "github.com/absmach/supermq/auth/api/grpc"
	domains "github.com/absmach/supermq/domains/private"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

var _ grpcDomainsV1.DomainsServiceServer = (*domainsGrpcServer)(nil)

type domainsGrpcServer struct {
	grpcDomainsV1.UnimplementedDomainsServiceServer
	deleteUserFromDomains kitgrpc.Handler
	retrieveEntity        kitgrpc.Handler
	retrieveIDByRoute     kitgrpc.Handler
}

func NewDomainsServer(svc domains.Service) grpcDomainsV1.DomainsServiceServer {
	return &domainsGrpcServer{
		deleteUserFromDomains: kitgrpc.NewServer(
			(deleteUserFromDomainsEndpoint(svc)),
			decodeDeleteUserRequest,
			encodeDeleteUserResponse,
		),
		retrieveEntity: kitgrpc.NewServer(
			retrieveEntityEndpoint(svc),
			decodeRetrieveEntityRequest,
			encodeRetrieveEntityResponse,
		),
		retrieveIDByRoute: kitgrpc.NewServer(
			retrieveIDByRouteEndpoint(svc),
			decodeRetrieveIDByRouteRequest,
			encodeRetrieveIDByRouteResponse,
		),
	}
}

func decodeDeleteUserRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcDomainsV1.DeleteUserReq)
	return deleteUserPoliciesReq{
		ID: req.GetId(),
	}, nil
}

func encodeDeleteUserResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(deleteUserRes)
	return &grpcDomainsV1.DeleteUserRes{Deleted: res.deleted}, nil
}

func (s *domainsGrpcServer) DeleteUserFromDomains(ctx context.Context, req *grpcDomainsV1.DeleteUserReq) (*grpcDomainsV1.DeleteUserRes, error) {
	_, res, err := s.deleteUserFromDomains.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcapi.EncodeError(err)
	}
	return res.(*grpcDomainsV1.DeleteUserRes), nil
}

func decodeRetrieveEntityRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcCommonV1.RetrieveEntityReq)

	return retrieveEntityReq{
		ID: req.GetId(),
	}, nil
}

func encodeRetrieveEntityResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(retrieveEntityRes)

	return &grpcCommonV1.RetrieveEntityRes{
		Entity: &grpcCommonV1.EntityBasic{
			Id:     res.id,
			Status: uint32(res.status),
		},
	}, nil
}

func (s *domainsGrpcServer) RetrieveEntity(ctx context.Context, req *grpcCommonV1.RetrieveEntityReq) (*grpcCommonV1.RetrieveEntityRes, error) {
	_, res, err := s.retrieveEntity.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcapi.EncodeError(err)
	}

	return res.(*grpcCommonV1.RetrieveEntityRes), nil
}

func decodeRetrieveIDByRouteRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpcCommonV1.RetrieveIDByRouteReq)

	return retrieveIDByRouteReq{
		Route: req.GetRoute(),
	}, nil
}

func encodeRetrieveIDByRouteResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(retrieveIDByRouteRes)

	return &grpcCommonV1.RetrieveEntityRes{
		Entity: &grpcCommonV1.EntityBasic{
			Id: res.id,
		},
	}, nil
}

func (s *domainsGrpcServer) RetrieveIDByRoute(ctx context.Context, req *grpcCommonV1.RetrieveIDByRouteReq) (*grpcCommonV1.RetrieveEntityRes, error) {
	_, res, err := s.retrieveIDByRoute.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcapi.EncodeError(err)
	}

	return res.(*grpcCommonV1.RetrieveEntityRes), nil
}

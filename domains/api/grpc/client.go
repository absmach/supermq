// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"time"

	grpcapi "github.com/absmach/supermq/auth/api/grpc"
	grpcDomainsV1 "github.com/absmach/supermq/internal/grpc/domains/v1"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

const domainsSvcName = "domains.v1.DomainsService"

var _ grpcDomainsV1.DomainsServiceClient = (*domainsGrpcClient)(nil)

type domainsGrpcClient struct {
	deleteUserFromDomains endpoint.Endpoint
	retrieveDomainStatus  endpoint.Endpoint
	timeout               time.Duration
}

// NewDomainsClient returns new domains gRPC client instance.
func NewDomainsClient(conn *grpc.ClientConn, timeout time.Duration) grpcDomainsV1.DomainsServiceClient {
	return &domainsGrpcClient{
		deleteUserFromDomains: kitgrpc.NewClient(
			conn,
			domainsSvcName,
			"DeleteUserFromDomains",
			encodeDeleteUserRequest,
			decodeDeleteUserResponse,
			grpcDomainsV1.DeleteUserRes{},
		).Endpoint(),
		retrieveDomainStatus: kitgrpc.NewClient(
			conn,
			domainsSvcName,
			"RetrieveDomainStatus",
			encodeRetrieveDomainStatusRequest,
			decodeRetrieveDomainStatusResponse,
			grpcDomainsV1.RetrieveDomainStatusRes{},
		).Endpoint(),
		timeout: timeout,
	}
}

func (client domainsGrpcClient) DeleteUserFromDomains(ctx context.Context, in *grpcDomainsV1.DeleteUserReq, opts ...grpc.CallOption) (*grpcDomainsV1.DeleteUserRes, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.deleteUserFromDomains(ctx, deleteUserPoliciesReq{
		ID: in.GetId(),
	})
	if err != nil {
		return &grpcDomainsV1.DeleteUserRes{}, grpcapi.DecodeError(err)
	}

	dpr := res.(deleteUserRes)
	return &grpcDomainsV1.DeleteUserRes{Deleted: dpr.deleted}, nil
}

func decodeDeleteUserResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*grpcDomainsV1.DeleteUserRes)
	return deleteUserRes{deleted: res.GetDeleted()}, nil
}

func encodeDeleteUserRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(deleteUserPoliciesReq)
	return &grpcDomainsV1.DeleteUserReq{
		Id: req.ID,
	}, nil
}

func (client domainsGrpcClient) RetrieveDomainStatus(ctx context.Context, in *grpcDomainsV1.RetrieveDomainStatusReq, opts ...grpc.CallOption) (*grpcDomainsV1.RetrieveDomainStatusRes, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.retrieveDomainStatus(ctx, retrieveDomainStatusReq{
		ID: in.GetId(),
	})
	if err != nil {
		return &grpcDomainsV1.RetrieveDomainStatusRes{}, grpcapi.DecodeError(err)
	}

	rdsr := res.(retrieveDomainStatusRes)
	return &grpcDomainsV1.RetrieveDomainStatusRes{Status: rdsr.status}, nil
}

func decodeRetrieveDomainStatusResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*grpcDomainsV1.RetrieveDomainStatusRes)
	return retrieveDomainStatusRes{status: res.GetStatus()}, nil
}

func encodeRetrieveDomainStatusRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(retrieveDomainStatusReq)
	return &grpcDomainsV1.RetrieveDomainStatusReq{
		Id: req.ID,
	}, nil
}

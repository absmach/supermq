// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"time"

	grpcAuthV1 "github.com/absmach/supermq/api/grpc/auth/v1"
	"github.com/absmach/supermq/auth"
	grpcapi "github.com/absmach/supermq/auth/api/grpc"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

const authSvcName = "auth.v1.AuthService"

type authGrpcClient struct {
	authenticate endpoint.Endpoint
	authorize    endpoint.Endpoint
	timeout      time.Duration
}

var _ grpcAuthV1.AuthServiceClient = (*authGrpcClient)(nil)

// NewAuthClient returns new auth gRPC client instance.
func NewAuthClient(conn *grpc.ClientConn, timeout time.Duration) grpcAuthV1.AuthServiceClient {
	return &authGrpcClient{
		authenticate: kitgrpc.NewClient(
			conn,
			authSvcName,
			"Authenticate",
			encodeIdentifyRequest,
			decodeIdentifyResponse,
			grpcAuthV1.AuthNRes{},
		).Endpoint(),
		authorize: kitgrpc.NewClient(
			conn,
			authSvcName,
			"Authorize",
			encodeAuthorizeRequest,
			decodeAuthorizeResponse,
			grpcAuthV1.AuthZRes{},
		).Endpoint(),
		timeout: timeout,
	}
}

func (client authGrpcClient) Authenticate(ctx context.Context, token *grpcAuthV1.AuthNReq, _ ...grpc.CallOption) (*grpcAuthV1.AuthNRes, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.authenticate(ctx, authenticateReq{token: token.GetToken()})
	if err != nil {
		return &grpcAuthV1.AuthNRes{}, grpcapi.DecodeError(err)
	}
	ir := res.(authenticateRes)
	return &grpcAuthV1.AuthNRes{Id: ir.id, UserId: ir.userID, UserRole: uint32(ir.userRole), Verified: ir.verified, TokenType: ir.tokenType}, nil
}

func encodeIdentifyRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(authenticateReq)
	return &grpcAuthV1.AuthNReq{Token: req.token}, nil
}

func decodeIdentifyResponse(_ context.Context, grpcRes any) (any, error) {
	res := grpcRes.(*grpcAuthV1.AuthNRes)
	return authenticateRes{id: res.GetId(), userID: res.GetUserId(), userRole: auth.Role(res.UserRole), verified: res.GetVerified(), tokenType: res.GetTokenType()}, nil
}

func (client authGrpcClient) Authorize(ctx context.Context, req *grpcAuthV1.AuthZReq, _ ...grpc.CallOption) (r *grpcAuthV1.AuthZRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	var authReqData authReq

	if policy := req.GetPolicy(); policy != nil {
		authReqData = authReq{
			TokenType:   policy.GetTokenType(),
			Domain:      policy.GetDomain(),
			SubjectType: policy.GetSubjectType(),
			Subject:     policy.GetSubject(),
			SubjectKind: policy.GetSubjectKind(),
			Relation:    policy.GetRelation(),
			Permission:  policy.GetPermission(),
			ObjectType:  policy.GetObjectType(),
			Object:      policy.GetObject(),
		}
	} else if pat := req.GetPat(); pat != nil {
		authReqData = authReq{
			TokenType:        pat.GetTokenType(),
			UserID:           pat.GetUserId(),
			PatID:            pat.GetPatId(),
			EntityType:       auth.EntityType(pat.GetEntityType()),
			OptionalDomainID: pat.GetOptionalDomainId(),
			Operation:        auth.Operation(pat.GetOperation()),
			EntityID:         pat.GetEntityId(),
		}
	}

	res, err := client.authorize(ctx, authReqData)
	if err != nil {
		return &grpcAuthV1.AuthZRes{}, grpcapi.DecodeError(err)
	}

	ar := res.(authorizeRes)
	return &grpcAuthV1.AuthZRes{Authorized: ar.authorized, Id: ar.id}, nil
}

func decodeAuthorizeResponse(_ context.Context, grpcRes any) (any, error) {
	res := grpcRes.(*grpcAuthV1.AuthZRes)
	return authorizeRes{authorized: res.Authorized, id: res.Id}, nil
}

func encodeAuthorizeRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(authReq)

	if req.PatID != "" {
		return &grpcAuthV1.AuthZReq{
			AuthType: &grpcAuthV1.AuthZReq_Pat{
				Pat: &grpcAuthV1.PATReq{
					TokenType:        req.TokenType,
					UserId:           req.UserID,
					PatId:            req.PatID,
					EntityType:       uint32(req.EntityType),
					OptionalDomainId: req.OptionalDomainID,
					Operation:        uint32(req.Operation),
					EntityId:         req.EntityID,
				},
			},
		}, nil
	}

	return &grpcAuthV1.AuthZReq{
		AuthType: &grpcAuthV1.AuthZReq_Policy{
			Policy: &grpcAuthV1.PolicyReq{
				TokenType:   req.TokenType,
				Domain:      req.Domain,
				SubjectType: req.SubjectType,
				Subject:     req.Subject,
				SubjectKind: req.SubjectKind,
				Relation:    req.Relation,
				Permission:  req.Permission,
				ObjectType:  req.ObjectType,
				Object:      req.Object,
			},
		},
	}, nil
}

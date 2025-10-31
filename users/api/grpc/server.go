// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	grpcUsersV1 "github.com/absmach/supermq/api/grpc/users/v1"
	grpcapi "github.com/absmach/supermq/auth/api/grpc"
	"github.com/absmach/supermq/users"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

var _ grpcUsersV1.UsersServiceServer = (*usersGrpcServer)(nil)

type usersGrpcServer struct {
	grpcUsersV1.UnimplementedUsersServiceServer
	sendEmail kitgrpc.Handler
}

// NewUsersServer creates a new users gRPC server.
func NewUsersServer(svc users.Service) grpcUsersV1.UsersServiceServer {
	return &usersGrpcServer{
		sendEmail: kitgrpc.NewServer(
			sendEmailEndpoint(svc),
			decodeSendEmailRequest,
			encodeSendEmailResponse,
		),
	}
}

func decodeSendEmailRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*grpcUsersV1.SendEmailReq)
	return sendEmailReq{
		to:      req.GetTo(),
		from:    req.GetFrom(),
		subject: req.GetSubject(),
		header:  req.GetHeader(),
		user:    req.GetUser(),
		content: req.GetContent(),
		footer:  req.GetFooter(),
	}, nil
}

func encodeSendEmailResponse(_ context.Context, grpcRes any) (any, error) {
	res := grpcRes.(sendEmailRes)
	return &grpcUsersV1.SendEmailRes{Sent: res.sent}, nil
}

func (s *usersGrpcServer) SendEmail(ctx context.Context, req *grpcUsersV1.SendEmailReq) (*grpcUsersV1.SendEmailRes, error) {
	_, res, err := s.sendEmail.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcapi.EncodeError(err)
	}
	return res.(*grpcUsersV1.SendEmailRes), nil
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	grpcUsersV1 "github.com/absmach/supermq/api/grpc/emails/v1"
	grpcapi "github.com/absmach/supermq/auth/api/grpc"
	"github.com/absmach/supermq/users"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

var _ grpcUsersV1.EmailServiceServer = (*usersGrpcServer)(nil)

type usersGrpcServer struct {
	grpcUsersV1.UnimplementedEmailServiceServer
	sendEmail kitgrpc.Handler
}

// NewUsersServer creates a new users gRPC server.
func NewUsersServer(svc users.Service) grpcUsersV1.EmailServiceServer {
	return &usersGrpcServer{
		sendEmail: kitgrpc.NewServer(
			sendEmailEndpoint(svc),
			decodeSendEmailRequest,
			encodeSendEmailResponse,
		),
	}
}

func decodeSendEmailRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(*grpcUsersV1.EmailReq)
	opts := req.GetOptions()

	tmpl := ""
	if req.Template != nil {
		tmpl = *req.Template
	}

	templateFile := ""
	if req.TemplateFile != nil {
		templateFile = *req.TemplateFile
	}

	return sendEmailReq{
		to:           req.GetTos(),
		toType:       req.GetToType(),
		from:         req.GetFrom(),
		fromType:     req.GetFromType(),
		subject:      req.GetSubject(),
		header:       opts["header"],
		user:         opts["user"],
		content:      req.GetContent(),
		footer:       opts["footer"],
		Template:     tmpl,
		templateFile: templateFile,
		Options:      opts,
	}, nil
}

func encodeSendEmailResponse(_ context.Context, grpcRes any) (any, error) {
	res := grpcRes.(sendEmailRes)
	errMsg := ""
	if !res.sent && res.err != nil {
		errMsg = res.err.Error()
	}
	return &grpcUsersV1.SendEmailRes{Error: errMsg}, nil
}

func (s *usersGrpcServer) SendEmail(ctx context.Context, req *grpcUsersV1.EmailReq) (*grpcUsersV1.SendEmailRes, error) {
	_, res, err := s.sendEmail.ServeGRPC(ctx, req)
	if err != nil {
		return nil, grpcapi.EncodeError(err)
	}
	return res.(*grpcUsersV1.SendEmailRes), nil
}

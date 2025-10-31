// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"time"

	grpcUsersV1 "github.com/absmach/supermq/api/grpc/users/v1"
	grpcapi "github.com/absmach/supermq/auth/api/grpc"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

const usersSvcName = "users.v1.UsersService"

var _ grpcUsersV1.UsersServiceClient = (*usersGrpcClient)(nil)

type usersGrpcClient struct {
	sendEmail endpoint.Endpoint
	timeout   time.Duration
}

// NewUsersClient returns new users gRPC client instance.
func NewUsersClient(conn *grpc.ClientConn, timeout time.Duration) grpcUsersV1.UsersServiceClient {
	return &usersGrpcClient{
		sendEmail: kitgrpc.NewClient(
			conn,
			usersSvcName,
			"SendEmail",
			encodeSendEmailClientRequest,
			decodeSendEmailClientResponse,
			grpcUsersV1.SendEmailRes{},
		).Endpoint(),
		timeout: timeout,
	}
}

func (client usersGrpcClient) SendEmailWithUserId(ctx context.Context, in *grpcUsersV1.SendEmailWithUserIdReq, opts ...grpc.CallOption) (*grpcUsersV1.SendEmailRes, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.sendEmail(ctx, sendEmailClientReq{
		to:      in.GetUsers(),
		from:    in.GetFrom(),
		subject: in.GetSubject(),
		header:  in.GetHeader(),
		user:    in.GetUser(),
		content: in.GetContent(),
		footer:  in.GetFooter(),
	})
	if err != nil {
		return &grpcUsersV1.SendEmailRes{}, grpcapi.DecodeError(err)
	}

	ser := res.(sendEmailClientRes)
	return &grpcUsersV1.SendEmailRes{Sent: ser.sent}, nil
}

func decodeSendEmailClientResponse(_ context.Context, grpcRes any) (any, error) {
	res := grpcRes.(*grpcUsersV1.SendEmailRes)
	return sendEmailClientRes{sent: res.GetSent()}, nil
}

func encodeSendEmailClientRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(sendEmailClientReq)
	return &grpcUsersV1.SendEmailWithUserIdReq{
		Users:   req.to,
		From:    req.from,
		Subject: req.subject,
		Header:  req.header,
		User:    req.user,
		Content: req.content,
		Footer:  req.footer,
	}, nil
}

type sendEmailClientReq struct {
	to      []string
	from    string
	subject string
	header  string
	user    string
	content string
	footer  string
}

type sendEmailClientRes struct {
	sent bool
}

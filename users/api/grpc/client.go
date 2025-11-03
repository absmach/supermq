// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"time"

	grpcUsersV1 "github.com/absmach/supermq/api/grpc/emails/v1"
	grpcapi "github.com/absmach/supermq/auth/api/grpc"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

const usersSvcName = "users.v1.UsersService"

var _ grpcUsersV1.EmailServiceClient = (*usersGrpcClient)(nil)

type usersGrpcClient struct {
	sendEmail endpoint.Endpoint
	timeout   time.Duration
}

// NewUsersClient returns new users gRPC client instance.
func NewUsersClient(conn *grpc.ClientConn, timeout time.Duration) grpcUsersV1.EmailServiceClient {
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

func (client usersGrpcClient) SendEmail(ctx context.Context, in *grpcUsersV1.EmailReq, opts ...grpc.CallOption) (*grpcUsersV1.SendEmailRes, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	options := in.GetOptions()
	res, err := client.sendEmail(ctx, sendEmailClientReq{
		to:      in.GetTos(),
		from:    in.GetFrom(),
		subject: in.GetSubject(),
		header:  options["header"],
		user:    options["user"],
		content: in.GetContent(),
		footer:  options["footer"],
	})
	if err != nil {
		return &grpcUsersV1.SendEmailRes{}, grpcapi.DecodeError(err)
	}

	ser := res.(sendEmailClientRes)
	errMsg := ""
	if !ser.sent {
		errMsg = "failed to send email"
	}
	return &grpcUsersV1.SendEmailRes{Error: errMsg}, nil
}

func decodeSendEmailClientResponse(_ context.Context, grpcRes any) (any, error) {
	res := grpcRes.(*grpcUsersV1.SendEmailRes)
	sent := res.GetError() == ""
	return sendEmailClientRes{sent: sent}, nil
}

func encodeSendEmailClientRequest(_ context.Context, grpcReq any) (any, error) {
	req := grpcReq.(sendEmailClientReq)
	return &grpcUsersV1.EmailReq{
		Tos:      req.to,
		ToType:   grpcUsersV1.ContactType_CONTACT_TYPE_ID,
		From:     req.from,
		FromType: grpcUsersV1.ContactType_CONTACT_TYPE_ID,
		Subject:  req.subject,
		Content:  req.content,
		Options: map[string]string{
			"header": req.header,
			"user":   req.user,
			"footer": req.footer,
		},
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

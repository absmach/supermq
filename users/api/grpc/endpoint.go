// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/users"
	"github.com/go-kit/kit/endpoint"
)

func sendEmailEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(sendEmailReq)
		if err := req.validate(); err != nil {
			return sendEmailRes{}, err
		}

		if err := svc.SendEmailWithUserId(ctx, req.to, req.from, req.subject, req.header, req.user, req.content, req.footer); err != nil {
			return sendEmailRes{}, err
		}

		return sendEmailRes{sent: true}, nil
	}
}

type sendEmailReq struct {
	to      []string
	from    string
	subject string
	header  string
	user    string
	content string
	footer  string
}

func (req sendEmailReq) validate() error {
	if len(req.to) == 0 {
		return errors.ErrMalformedEntity
	}
	if req.subject == "" {
		return errors.ErrMalformedEntity
	}
	return nil
}

type sendEmailRes struct {
	sent bool
}

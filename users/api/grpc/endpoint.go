// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"bytes"
	"context"
	"text/template"

	grpcUsersV1 "github.com/absmach/supermq/api/grpc/emails/v1"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/users"
	"github.com/go-kit/kit/endpoint"
)

func sendEmailEndpoint(svc users.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(sendEmailReq)
		if err := req.validate(); err != nil {
			return sendEmailRes{
				err: err,
			}, err
		}

		if err := svc.SendEmail(ctx, req.to, req.toType, req.from, req.fromType, req.subject, req.header, req.user, req.content, req.footer); err != nil {
			return sendEmailRes{
				err: err,
			}, err
		}

		return sendEmailRes{
			sent: true,
		}, nil
	}
}

type sendEmailReq struct {
	to           []string
	toType       grpcUsersV1.ContactType
	from         string
	fromType     grpcUsersV1.ContactType
	subject      string
	header       string
	user         string
	content      string
	footer       string
	Template     string
	templateFile string
	Options      map[string]string
}

func (req sendEmailReq) validate() error {
	if len(req.to) == 0 {
		return errors.ErrMalformedEntity
	}
	if req.subject == "" {
		return errors.ErrMalformedEntity
	}

	if req.Template != "" {
		t, err := template.New("body").Parse(req.Template)
		if err != nil {
			return err
		}
		buff := new(bytes.Buffer)
		if err := t.Execute(buff, req.Options); err != nil {
			return err
		}
	}

	return nil
}

type sendEmailRes struct {
	sent bool
	err  error
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/absmach/supermq/auditlogs"
	api "github.com/absmach/supermq/api/http"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/go-kit/kit/endpoint"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
)

func retrieveAuditLogEndpoint(svc auditlogs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retriveAuditLogReq)

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		log, err := svc.RetrieveByID(ctx, session, req.id)
		if err != nil {
			return nil, err
		}

		return retrieveAuditLogRes{log}, nil
	}
}

func retrieveAllAuditLogsEndpoint(svc auditlogs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrieveAllAuditLogsReq)

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		page, err := svc.RetrieveAll(ctx, session, req.Page)
		if err != nil {
			return nil, err
		}

		return retrieveAllAuditLogsRes{page}, nil
	}
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/absmach/supermq/internal/api"
	"github.com/absmach/supermq/journal"
	"github.com/absmach/supermq/pkg/apiutil"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/go-kit/kit/endpoint"
)

func retrieveJournalsEndpoint(svc journal.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrieveJournalsReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthorization
		}

		page, err := svc.RetrieveAll(ctx, session, req.page)
		if err != nil {
			return nil, err
		}

		return pageRes{
			JournalsPage: page,
		}, nil
	}
}

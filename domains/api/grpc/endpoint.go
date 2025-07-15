// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	domains "github.com/absmach/supermq/domains/private"
	"github.com/go-kit/kit/endpoint"
)

func deleteUserFromDomainsEndpoint(svc domains.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deleteUserPoliciesReq)
		if err := req.validate(); err != nil {
			return deleteUserRes{}, err
		}

		if err := svc.DeleteUserFromDomains(ctx, req.ID); err != nil {
			return deleteUserRes{}, err
		}

		return deleteUserRes{deleted: true}, nil
	}
}

func retrieveStatusEndpoint(svc domains.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrieveStatusReq)
		if err := req.validate(); err != nil {
			return retrieveStatusRes{}, err
		}

		status, err := svc.RetrieveStatus(ctx, req.ID)
		if err != nil {
			return retrieveStatusRes{}, err
		}

		return retrieveStatusRes{
			status: uint8(status),
		}, nil
	}
}

func retrieveByRouteEndpoint(svc domains.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrieveByRouteReq)
		if err := req.validate(); err != nil {
			return retrieveEntityRes{}, err
		}

		dom, err := svc.RetrieveByRoute(ctx, req.Route)
		if err != nil {
			return retrieveEntityRes{}, err
		}

		return retrieveEntityRes{
			id:     dom.ID,
			status: uint8(dom.Status),
		}, nil
	}
}

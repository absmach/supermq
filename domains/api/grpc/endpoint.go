// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/absmach/supermq/domains"
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

func retrieveDomainStatusEndpoint(svc domains.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrieveDomainStatusReq)
		if err := req.validate(); err != nil {
			return retrieveDomainStatusRes{}, err
		}

		status, err := svc.RetrieveStatus(ctx, req.ID)
		if err != nil {
			return retrieveDomainStatusRes{}, err
		}

		return retrieveDomainStatusRes{status: status.String()}, nil
	}
}
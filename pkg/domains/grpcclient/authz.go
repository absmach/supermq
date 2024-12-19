// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpcclient

import (
	"context"

	"github.com/absmach/supermq/domains"
	grpcCommonV1 "github.com/absmach/supermq/internal/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/internal/grpc/domains/v1"
	pkgDomains "github.com/absmach/supermq/pkg/domains"
	"github.com/absmach/supermq/pkg/grpcclient"
)

type authorization struct {
	domainsSvcClient grpcDomainsV1.DomainsServiceClient
}

var _ pkgDomains.Authorization = (*authorization)(nil)

func NewAuthorization(ctx context.Context, cfg grpcclient.Config) (pkgDomains.Authorization, grpcDomainsV1.DomainsServiceClient, grpcclient.Handler, error) {
	domainsClient, domainsHandler, err := grpcclient.SetupDomainsClient(ctx, cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	return authorization{domainsSvcClient: domainsClient}, domainsClient, domainsHandler, nil
}

func (a authorization) RetrieveEntity(ctx context.Context, id string) (domains.Domain, error) {
	req := grpcCommonV1.RetrieveEntityReq{
		Id: id,
	}
	res, err := a.domainsSvcClient.RetrieveEntity(ctx, &req)
	if err != nil {
		return domains.Domain{}, err
	}

	return domains.Domain{
		ID:     res.Entity.GetId(),
		Status: domains.Status(res.Entity.GetStatus()),
	}, nil
}

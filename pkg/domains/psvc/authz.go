// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package domainscache

import (
	"context"

	"github.com/absmach/supermq/domains"
	"github.com/absmach/supermq/domains/private"
	pkgDomains "github.com/absmach/supermq/pkg/domains"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
)

type authorization struct {
	psvc private.Service
}

var _ pkgDomains.Authorization = (*authorization)(nil)

func NewAuthorization(psvc private.Service) pkgDomains.Authorization {
	return authorization{
		psvc: psvc,
	}
}

func (a authorization) RetrieveStatus(ctx context.Context, id string) (domains.Status, error) {
	status, err := a.psvc.RetrieveStatus(ctx, id)
	if err != nil {
		return domains.AllStatus, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return status, nil
}

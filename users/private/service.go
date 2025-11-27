// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package private

import (
	"context"

	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/users"
)

type Service interface {
	RetrieveByIDs(ctx context.Context, ids []string, offset, limit uint64) (users.UsersPage, error)
}

var _ Service = (*service)(nil)

func New(repo users.Repository) Service {
	return service{
		repo: repo,
	}
}

type service struct {
	repo users.Repository
}

func (svc service) RetrieveByIDs(ctx context.Context, ids []string, offset, limit uint64) (users.UsersPage, error) {
	if len(ids) == 0 {
		return users.UsersPage{}, svcerr.ErrMalformedEntity
	}

	if limit == 0 {
		limit = uint64(len(ids))
	}

	pm := users.Page{
		IDs:    ids,
		Offset: offset,
		Limit:  limit,
	}

	page, err := svc.repo.RetrieveAllByIDs(ctx, pm)
	if err != nil {
		return users.UsersPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return page, nil
}

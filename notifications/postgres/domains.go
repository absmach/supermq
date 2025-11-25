// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"

	"github.com/absmach/supermq/notifications"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
)

var _ notifications.DomainRepository = (*domainRepository)(nil)

type domainRepository struct {
	db postgres.Database
}

// NewDomainRepository instantiates a PostgreSQL implementation of domain repository.
func NewDomainRepository(db postgres.Database) notifications.DomainRepository {
	return &domainRepository{
		db: db,
	}
}

func (repo *domainRepository) RetrieveByID(ctx context.Context, id string) (notifications.Domain, error) {
	query := `SELECT id, name FROM domains WHERE id = $1`

	var domain notifications.Domain
	if err := repo.db.QueryRowxContext(ctx, query, id).StructScan(&domain); err != nil {
		return notifications.Domain{}, errors.Wrap(repoerr.ErrNotFound, err)
	}

	return domain, nil
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"

	// required for SQL access.

	"github.com/mainflux/mainflux/internal/postgres"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	pgclients "github.com/mainflux/mainflux/pkg/clients/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
)

var _ mfclients.Repository = (*clientRepo)(nil)

type clientRepo struct {
	pgclients.ClientRepository
}

// NewRepository instantiates a PostgreSQL
// implementation of Clients repository.
func NewRepository(db postgres.Database) mfclients.Repository {
	return &clientRepo{
		ClientRepository: pgclients.ClientRepository{DB: db},
	}
}

// RetrieveByIdentity retrieves client by its unique credentials.
func (clientRepo) RetrieveByIdentity(ctx context.Context, identity string) (mfclients.Client, error) {
	return mfclients.Client{}, nil
}

func (repo clientRepo) Save(ctx context.Context, cs ...mfclients.Client) ([]mfclients.Client, error) {
	tx, err := repo.ClientRepository.DB.BeginTxx(ctx, nil)
	if err != nil {
		return []mfclients.Client{}, errors.Wrap(errors.ErrCreateEntity, err)
	}

	for _, cli := range cs {
		q := `INSERT INTO clients (id, name, tags, owner_id, identity, secret, metadata, created_at, updated_at, updated_by, status)
        VALUES (:id, :name, :tags, :owner_id, :identity, :secret, :metadata, :created_at, :updated_at, :updated_by, :status)
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`

		dbcli, err := pgclients.ToDBClient(cli)
		if err != nil {
			return []mfclients.Client{}, errors.Wrap(errors.ErrCreateEntity, err)
		}

		if _, err := tx.NamedExecContext(ctx, q, dbcli); err != nil {
			if err := tx.Rollback(); err != nil {
				return []mfclients.Client{}, postgres.HandleError(err, errors.ErrCreateEntity)
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return []mfclients.Client{}, errors.Wrap(errors.ErrCreateEntity, err)
	}

	return cs, nil
}

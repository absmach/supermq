// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"

	// required for SQL access.

	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/pkg/clients"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	pgclients "github.com/mainflux/mainflux/pkg/clients/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
)

var _ mfclients.Repository = (*clientRepo)(nil)

type clientRepo struct {
	pgclients.ClientRepository
}

type Repository interface {
	mfclients.Repository

	// Save persists the client account. A non-nil error is returned to indicate
	// operation failure.
	Save(ctx context.Context, client ...clients.Client) ([]clients.Client, error)

	// RetrieveBySecret retrieves a client based on the secret (key).
	RetrieveBySecret(ctx context.Context, key string) (clients.Client, error)
}

// NewRepository instantiates a PostgreSQL
// implementation of Clients repository.
func NewRepository(db postgres.Database) Repository {
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
	var clients []mfclients.Client

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

		row, err := repo.ClientRepository.DB.NamedQueryContext(ctx, q, dbcli)
		if err != nil {
			return []mfclients.Client{}, postgres.HandleError(err, errors.ErrCreateEntity)
		}

		defer row.Close()
		row.Next()
		dbcli = pgclients.DBClient{}
		if err := row.StructScan(&dbcli); err != nil {
			return []mfclients.Client{}, err
		}

		client, err := pgclients.ToClient(dbcli)
		if err != nil {
			return []mfclients.Client{}, err
		}
		clients = append(clients, client)
	}
	if err = tx.Commit(); err != nil {
		return []mfclients.Client{}, errors.Wrap(errors.ErrCreateEntity, err)
	}

	return clients, nil
}

func (repo clientRepo) RetrieveBySecret(ctx context.Context, key string) (clients.Client, error) {
	q := fmt.Sprintf(`SELECT id, name, tags, COALESCE(owner_id, '') AS owner_id, identity, secret, metadata, created_at, updated_at, updated_by, status
        FROM clients
        WHERE secret = $1 AND status = %d`, clients.EnabledStatus)

	dbc := pgclients.DBClient{
		Secret: key,
	}

	if err := repo.DB.QueryRowxContext(ctx, q, key).StructScan(&dbc); err != nil {
		if err == sql.ErrNoRows {
			return clients.Client{}, errors.Wrap(errors.ErrNotFound, err)

		}
		return clients.Client{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	return pgclients.ToClient(dbc)
}

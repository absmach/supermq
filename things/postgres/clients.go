// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"

	mgclients "github.com/absmach/magistrala/pkg/clients"
	pgclients "github.com/absmach/magistrala/pkg/clients/postgres"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	"github.com/absmach/magistrala/pkg/postgres"
	rolesPostgres "github.com/absmach/magistrala/pkg/roles/repo/postgres"
	"github.com/absmach/magistrala/things"
)

const thingsRolesTableNamePrefix = "things"

var _ things.Repository = (*thingsRepo)(nil)

type thingsRepo struct {
	clientsRepo
	rolesPostgres.Repository
}

// NewRepository instantiates a PostgreSQL
// implementation of Clients repository.
func NewRepository(db postgres.Database) things.Repository {
	repo := rolesPostgres.NewRepository(db, thingsRolesTableNamePrefix)

	return &thingsRepo{
		clientsRepo{
			pgclients.Repository{DB: db},
		},
		repo,
	}
}

type clientsRepo struct {
	pgclients.Repository
}

func (repo *clientsRepo) Save(ctx context.Context, cs ...mgclients.Client) ([]mgclients.Client, error) {
	tx, err := repo.DB.BeginTxx(ctx, nil)
	if err != nil {
		return []mgclients.Client{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}
	var clients []mgclients.Client

	for _, cli := range cs {
		q := `INSERT INTO clients (id, name, tags, domain_id, identity, secret, metadata, created_at, updated_at, updated_by, status)
        VALUES (:id, :name, :tags, :domain_id, :identity, :secret, :metadata, :created_at, :updated_at, :updated_by, :status)
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(domain_id, '') AS domain_id, status, created_at, updated_at, updated_by`

		dbcli, err := pgclients.ToDBClient(cli)
		if err != nil {
			return []mgclients.Client{}, errors.Wrap(repoerr.ErrCreateEntity, err)
		}

		row, err := repo.DB.NamedQueryContext(ctx, q, dbcli)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return []mgclients.Client{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
			}
			return []mgclients.Client{}, errors.Wrap(repoerr.ErrCreateEntity, err)
		}

		defer row.Close()

		if row.Next() {
			dbcli = pgclients.DBClient{}
			if err := row.StructScan(&dbcli); err != nil {
				return []mgclients.Client{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
			}

			client, err := pgclients.ToClient(dbcli)
			if err != nil {
				return []mgclients.Client{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
			}
			clients = append(clients, client)
		}
	}
	if err = tx.Commit(); err != nil {
		return []mgclients.Client{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	return clients, nil
}

func (repo *clientsRepo) RetrieveBySecret(ctx context.Context, key string) (mgclients.Client, error) {
	q := fmt.Sprintf(`SELECT id, name, tags, COALESCE(domain_id, '') AS domain_id, identity, secret, metadata, created_at, updated_at, updated_by, status
        FROM clients
        WHERE secret = :secret AND status = %d`, mgclients.EnabledStatus)

	dbc := pgclients.DBClient{
		Secret: key,
	}

	rows, err := repo.DB.NamedQueryContext(ctx, q, dbc)
	if err != nil {
		return mgclients.Client{}, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	dbc = pgclients.DBClient{}
	if rows.Next() {
		if err = rows.StructScan(&dbc); err != nil {
			return mgclients.Client{}, postgres.HandleError(repoerr.ErrViewEntity, err)
		}

		client, err := pgclients.ToClient(dbc)
		if err != nil {
			return mgclients.Client{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
		}

		return client, nil
	}

	return mgclients.Client{}, repoerr.ErrNotFound
}

func (repo *clientsRepo) RemoveThings(ctx context.Context, clientIDs ...[]string) error {
	q := "DELETE FROM clients AS c  WHERE c.id = ANY(:client_ids) ;"

	params := map[string]interface{}{
		"client_ids": clientIDs,
	}
	result, err := repo.DB.NamedExecContext(ctx, q, params)
	if err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return repoerr.ErrNotFound
	}

	return nil
}

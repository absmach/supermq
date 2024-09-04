// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"github.com/absmach/magistrala/pkg/errors"

	entityRolesRepo "github.com/absmach/magistrala/pkg/entityroles/postrgres"
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

var errRoleMigration = errors.New("role migration initialization failed")

// Migration of Auth service.
func Migration() (*migrate.MemoryMigrationSource, error) {
	rolesMigration, err := entityRolesRepo.Migration("domains", "id")
	if err != nil {
		return &migrate.MemoryMigrationSource{}, errors.Wrap(errRoleMigration, err)
	}

	domainMigrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "domain_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS domains (
                        id          VARCHAR(36) PRIMARY KEY,
                        name        VARCHAR(254),
                        tags        TEXT[],
                        metadata    JSONB,
					    alias       VARCHAR(254) NOT NULL UNIQUE,
                        created_at  TIMESTAMP,
                        updated_at  TIMESTAMP,
                        updated_by  VARCHAR(254),
                        created_by  VARCHAR(254),
                        status      SMALLINT NOT NULL DEFAULT 0 CHECK (status >= 0)
                    );`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS domains`,
				},
			},
		},
	}

	domainMigrations.Migrations = append(domainMigrations.Migrations, rolesMigration.Migrations...)

	return domainMigrations, nil
}

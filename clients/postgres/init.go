// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	gpostgres "github.com/absmach/supermq/groups/postgres"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	rolesPostgres "github.com/absmach/supermq/pkg/roles/repo/postgres"
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

func Migration() (*migrate.MemoryMigrationSource, error) {
	clientsRolesMigration, err := rolesPostgres.Migration(rolesTableNamePrefix, entityTableName, entityIDColumnName)
	if err != nil {
		return &migrate.MemoryMigrationSource{}, errors.Wrap(repoerr.ErrRoleMigration, err)
	}

	clientsMigration := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "clients_01",
				// VARCHAR(36) for columns with IDs as UUIDS have a maximum of 36 characters
				// STATUS 0 to imply enabled and 1 to imply disabled
				Up: []string{
					`CREATE TABLE IF NOT EXISTS clients (
						id			       VARCHAR(36) PRIMARY KEY,
						name		       VARCHAR(1024),
						domain_id	       VARCHAR(36) NOT NULL,
						parent_group_id    VARCHAR(36) DEFAULT NULL,
						identity	       VARCHAR(254),
						secret		       VARCHAR(4096) NOT NULL,
						tags		       TEXT[],
						metadata	       JSONB,
						created_at	       TIMESTAMP,
						updated_at	       TIMESTAMP,
						updated_by         VARCHAR(254),
						status		       SMALLINT NOT NULL DEFAULT 0 CHECK (status >= 0),
						UNIQUE		       (domain_id, secret),
						UNIQUE		       (domain_id, name),
						UNIQUE		       (domain_id, id)
					)`,
					`CREATE TABLE IF NOT EXISTS connections (
						channel_id    VARCHAR(36),
						domain_id 	  VARCHAR(36),
						client_id     VARCHAR(36),
						type          SMALLINT NOT NULL CHECK (type IN (1, 2)),
						FOREIGN KEY   (client_id, domain_id) REFERENCES clients (id, domain_id) ON DELETE CASCADE ON UPDATE CASCADE,
						PRIMARY KEY   (channel_id, domain_id, client_id, type)
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS clients`,
					`DROP TABLE IF EXISTS connections`,
				},
			},
			{
				Id: "clients_02",
				Up: []string{
					`ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_domain_id_name_key`,
				},
				Down: []string{
					`ALTER TABLE clients ADD CONSTRAINT clients_domain_id_name_key UNIQUE (domain_id, name)`,
				},
			},
			{
				Id: "clients_03",
				Up: []string{
					`ALTER TABLE clients ALTER COLUMN created_at TYPE TIMESTAMPTZ;`,
					`ALTER TABLE clients ALTER COLUMN updated_at TYPE TIMESTAMPTZ;`,
				},
				Down: []string{
					`ALTER TABLE clients ALTER COLUMN created_at TYPE TIMESTAMP;`,
					`ALTER TABLE clients ALTER COLUMN updated_at TYPE TIMESTAMP;`,
				},
			},
		},
	}

	clientsMigration.Migrations = append(clientsMigration.Migrations, clientsRolesMigration.Migrations...)

	groupsMigration, err := gpostgres.Migration()
	if err != nil {
		return &migrate.MemoryMigrationSource{}, err
	}

	clientsMigration.Migrations = append(clientsMigration.Migrations, groupsMigration.Migrations...)

	return clientsMigration, nil
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "activities_01",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS activities (
						id			UUID NOT NULL,
						operation	VARCHAR NOT NULL,
						occurred_at	TIMESTAMP NOT NULL,
						payload		JSONB NOT NULL,
						UNIQUE (id, operation, occurred_at),
						PRIMARY KEY (id, operation, occurred_at)
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS activities`,
				},
			},
		},
	}
}

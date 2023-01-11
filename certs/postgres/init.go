// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

// Migration of certs service.
func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "certs_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS certs (
						thing_id     TEXT NOT NULL,
						owner_id     TEXT NOT NULL,
						expire       TIMESTAMPTZ NOT NULL,
						serial       TEXT NOT NULL,
						PRIMARY KEY  (thing_id, owner_id, serial)
					);`,
				},
				Down: []string{
					"DROP TABLE IF EXISTS certs;",
				},
			},
		},
	}
}

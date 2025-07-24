// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import migrate "github.com/rubenv/sql-migrate"

// Migration of Certs service.
func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "certs_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS certs (
						client_id    		TEXT NOT NULL,
						expiry_time  		TIMESTAMPTZ NOT NULL,
						serial_number       TEXT NOT NULL,
						revoked             BOOLEAN DEFAULT FALSE,
						PRIMARY KEY  (client_id, serial_number)
					);`,
				},
				Down: []string{
					"DROP TABLE IF EXISTS certs;",
				},
			},
		},
	}
}

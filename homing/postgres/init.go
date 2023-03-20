// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import migrate "github.com/rubenv/sql-migrate"

// Migration of Users service
func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "telemetry_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS telemetry (
						id			UUID	PRIMARY	KEY,
						ip_address	TEXT	UNIQUE	NOT	NULL,
					 	longitude 	FLOAT	NOT	NULL,
						latitude	FLOAT	NOT NULL,
						mf_version	TEXT,
						services	TEXT,
					)`,
				},
				Down: []string{"DROP TABLE telemetry"},
			},
		},
	}
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pat

import (
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

// Migration of Auth service.
func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "pat_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS pats (
						id 					VARCHAR(36) PRIMARY KEY,
                        name        		VARCHAR(254),
						user_id 			VARCHAR(36),
						description			TEXT,
						domain 				VARCHAR(36),
						secret				TEXT,
						issued_at			TIMESTAMP,
						expires_at 			TIMESTAMP,
						updated_at 			TIMESTAMP,
						revoked 			BOOLEAN,
						revoked_at 			TIMESTAMP,
						entity_type 		TEXT,
						allowed_operation	[]TEXT,
						constraints 		TEXT,
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS pats`,
				},
			},
		},
	}
}

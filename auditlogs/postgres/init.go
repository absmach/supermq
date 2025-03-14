// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

// Migration of Audit Logs service.
func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "auditlogs_01",

				Up: []string{
					`CREATE TABLE IF NOT EXISTS audit_logs (
						id          VARCHAR(36) PRIMARY KEY,
						request_id  VARCHAR(36),
						domain_id   VARCHAR(36),
						occured_at  TIMESTAMP,
						actor_id    VARCHAR(36),
						current_state JSONB,
						previous_state JSONB,
						state_attributes JSONB,
						entity_id   VARCHAR(36),
						entity_type VARCHAR(36),
						metadata    JSONB
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS audit_logs`,
				},
			},
		},
	}
}

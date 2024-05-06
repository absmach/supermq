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
						id 			UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
						operation	VARCHAR NOT NULL CHECK (operation <> ''),
						occurred_at	TIMESTAMP NOT NULL,
						attributes	JSONB NOT NULL,
						metadata	JSONB,
						UNIQUE(operation, occurred_at, attributes)
					)`,
					`CREATE INDEX idx_activities_default_user_filter ON activities(operation, (attributes->>'id'), (attributes->>'user_id'), occurred_at DESC);`,
					`CREATE INDEX idx_activities_default_group_filter ON activities(operation, (attributes->>'id'), (attributes->>'group_id'), occurred_at DESC);`,
					`CREATE INDEX idx_activities_default_thing_filter ON activities(operation, (attributes->>'id'), (attributes->>'thing_id'), occurred_at DESC);`,
					`CREATE INDEX idx_activities_default_channel_filter ON activities(operation, (attributes->>'id'), (attributes->>'channel_id'), occurred_at DESC);`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS activities`,
				},
			},
		},
	}
}

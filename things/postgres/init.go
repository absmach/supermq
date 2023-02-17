package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "clients_01",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS clients (
						id          VARCHAR(254),
						name        VARCHAR(1024),
						owner       VARCHAR(254),
						identity    VARCHAR(254),
						secret      VARCHAR(4096) UNIQUE NOT NULL,
						tags        TEXT[],
						metadata    JSONB,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						status      SMALLINT NOT NULL CHECK (status >= 0) DEFAULT 1,
						PRIMARY KEY (id, owner)
					)`,
					`CREATE TABLE IF NOT EXISTS groups (
						id          VARCHAR(254),
						parent_id   VARCHAR(254),
						owner       VARCHAR(254) NOT NULL,
						name        VARCHAR(1024) NOT NULL,
						description VARCHAR(1024),
						metadata    JSONB,
						created_at  TIMESTAMP,
						updated_at  TIMESTAMP,
						status      SMALLINT NOT NULL CHECK (status >= 0) DEFAULT 1,
						PRIMARY KEY (id, owner)
					)`,
					`CREATE TABLE IF NOT EXISTS connections (
						owner         VARCHAR(254) NOT NULL,
						thing_id      VARCHAR(254) NOT NULL,
						thing_owner   VARCHAR(254) NOT NULL,
						channel_id    VARCHAR(254) NOT NULL,
						channel_owner VARCHAR(254) NOT NULL,
						actions       TEXT[],
						created_at    TIMESTAMP,
						updated_at    TIMESTAMP,
						FOREIGN KEY (thing_id, thing_owner) REFERENCES clients (id, owner) ON DELETE CASCADE ON UPDATE CASCADE,
						FOREIGN KEY (channel_id, channel_owner) REFERENCES groups (id, owner) ON DELETE CASCADE ON UPDATE CASCADE,
						PRIMARY KEY (channel_id, channel_owner, thing_id, thing_owner)
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS clients`,
					`DROP TABLE IF EXISTS groups`,
					`DROP TABLE IF EXISTS connections`,
				},
			},
		},
	}
}

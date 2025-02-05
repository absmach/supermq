// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

// Migration of Auth service.
func Migration() *migrate.MemoryMigrationSource {
	return &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "auth_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS keys (
                        id          VARCHAR(254) NOT NULL,
                        type        SMALLINT,
                        subject     VARCHAR(254) NOT NULL,
                        issuer_id   VARCHAR(254) NOT NULL,
                        issued_at   TIMESTAMP NOT NULL,
                        expires_at  TIMESTAMP,
                        PRIMARY KEY (id, issuer_id)
                    )`,

					`CREATE TABLE IF NOT EXISTS domains (
                        id          VARCHAR(36) PRIMARY KEY,
                        name        VARCHAR(254),
                        tags        TEXT[],
                        metadata    JSONB,
                        alias       VARCHAR(254) NULL UNIQUE,
                        created_at  TIMESTAMP,
                        updated_at  TIMESTAMP,
                        updated_by  VARCHAR(254),
                        created_by  VARCHAR(254),
                        status      SMALLINT NOT NULL DEFAULT 0 CHECK (status >= 0)
                    );`,
					`CREATE TABLE IF NOT EXISTS policies (
                        subject_type        VARCHAR(254) NOT NULL,
                        subject_id          VARCHAR(254) NOT NULL,
                        subject_relation    VARCHAR(254) NOT NULL,
                        relation            VARCHAR(254) NOT NULL,
                        object_type         VARCHAR(254) NOT NULL,
                        object_id           VARCHAR(254) NOT NULL,
                        CONSTRAINT unique_policy_constraint UNIQUE (subject_type, subject_id, subject_relation, relation, object_type, object_id)
                    );`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS keys`,
				},
			},
			{
				Id: "auth_2",
				Up: []string{
					`ALTER TABLE domains ALTER COLUMN alias SET NOT NULL`,
				},
			},
			{
				Id: "auth_3",
				Up: []string{
					`DROP TABLE IF EXISTS policies;
                     DROP TABLE IF EXISTS domains;
                    `,
				},
			},
			{
				Id: "auth_4",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS pats (
						id 					VARCHAR(36) PRIMARY KEY,
                        name        		VARCHAR(254),
						user_id	 			VARCHAR(36),
						description			TEXT,
						secret				TEXT,
						issued_at			TIMESTAMP,
						expires_at 			TIMESTAMP,
						updated_at 			TIMESTAMP,
						revoked 			BOOLEAN,
						revoked_at 			TIMESTAMP,
						entity_type 		TEXT,
						last_used_at		TIMESTAMP
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS pats`,
				},
			},
			{
				Id: "auth_5",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS pat_scopes (
						id              SERIAL PRIMARY KEY,
						pat_id          VARCHAR(36) REFERENCES pats(id) ON DELETE CASCADE,
						platform_type   VARCHAR(50) NOT NULL,
						domain_id       VARCHAR(36),
						domain_type     VARCHAR(50),
						operation_type  VARCHAR(50) NOT NULL,
						entity_ids      TEXT[] NOT NULL,
						UNIQUE (pat_id, platform_type, domain_id, domain_type, operation_type)
					)`,
				},
				Down: []string{
					`DROP TABLE IF EXISTS pat_scopes`,
				},
			},
		},
	}
}

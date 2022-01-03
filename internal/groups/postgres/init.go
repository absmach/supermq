// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

// Config defines the options that are used when connecting to a PostgreSQL instance
type Config struct {
	Host        string
	Port        string
	User        string
	Pass        string
	Name        string
	SSLMode     string
	SSLCert     string
	SSLKey      string
	SSLRootCert string
}

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s sslcert=%s sslkey=%s sslrootcert=%s", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass, cfg.SSLMode, cfg.SSLCert, cfg.SSLKey, cfg.SSLRootCert)

	db, err := sqlx.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrateDB(db *sqlx.DB) error {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "groups_1",
				Up: []string{
					`CREATE EXTENSION IF NOT EXISTS LTREE`,
					`CREATE TABLE IF NOT EXISTS groups ( 
						id          VARCHAR(254) UNIQUE NOT NULL,
						parent_id   VARCHAR(254), 
						owner_id    VARCHAR(254),
						name        VARCHAR(254) NOT NULL,
						description VARCHAR(1024),
						metadata    JSONB,
						path        LTREE,
						created_at  TIMESTAMPTZ,
						updated_at  TIMESTAMPTZ,
						UNIQUE (owner_id, name, parent_id),
						FOREIGN KEY (parent_id) REFERENCES groups (id) ON DELETE CASCADE
				   )`,
					`CREATE INDEX path_gist_idx ON groups USING GIST (path);`,
					`CREATE OR REPLACE FUNCTION inherit_group()
					 RETURNS trigger 
					 LANGUAGE PLPGSQL
					 AS
					 $$
					 BEGIN
					 IF NEW.parent_id IS NULL OR NEW.parent_id = '' THEN
						RETURN NEW;
					 END IF;
					 IF NOT EXISTS (SELECT id FROM groups WHERE id = NEW.parent_id) THEN
						RAISE EXCEPTION 'wrong parent id';
					 END IF;
					 SELECT text2ltree(ltree2text(path) || '.' || NEW.id) INTO NEW.path FROM groups WHERE id = NEW.parent_id;
					 RETURN NEW;
					 END;
					 $$`,
					`CREATE TRIGGER inherit_group_tr
					 BEFORE INSERT
					 ON groups
					 FOR EACH ROW
					 EXECUTE PROCEDURE inherit_group();`,
				},
				Down: []string{
					`DROP EXTENSION IF EXISTS LTREE`,
					`DROP TABLE IF EXISTS groups`,
					`DROP FUNCTION IF EXISTS inherit_group`,
					`DROP TRIGGER IF EXISTS inherit_group_tr ON groups`,
				},
			},
		},
	}

	_, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	return err
}

//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // required for SQL access
	migrate "github.com/rubenv/sql-migrate"
)

// Connect creates a connection to the PostgreSQL instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate
// failure.
func Connect(host, port, name, user, pass, sslMode string) (*sql.DB, error) {
	url := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", host, port, user, name, pass, sslMode)

	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := migrateDB(db); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDB(db *sql.DB) error {
	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id: "things_1",
				Up: []string{
					`CREATE TABLE IF NOT EXISTS things (
						id          	 BIGSERIAL,
						mainflux_key     CHAR(36) UNIQUE,
						owner            VARCHAR(254) NOT NULL,
						mainflux_thing   TEXT UNIQUE,
                        external_id      TEXT UNIQUE NOT NULL,
						mainflux_channel TEXT UNIQUE,
                        status           BIGINT NOT NULL,
						PRIMARY KEY (id, owner)
					)`,
				},
				Down: []string{
					"DROP TABLE things",
				},
			},
		},
	}

	_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	return err
}

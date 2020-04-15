// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go"
	"github.com/golang-migrate/migrate/v4"
	migrateClickhouse "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file" // required for read migrastions from files
	"github.com/jmoiron/sqlx"
)

// Config defines the options that are used when connecting to a Clickhouse instance
type Config struct {
	Host string
	Port string
	User string
	Pass string
	Name string
}

// Connect creates a connection to the Clickhouse instance and applies any
// unapplied database migrations. A non-nil error is returned to indicate
// failure.
func Connect(cfg Config) (*sqlx.DB, error) {
	url := fmt.Sprintf("tcp://%s:%s?username=%s&database=%s&password=%s&debug=false", cfg.Host, cfg.Port, cfg.User, cfg.Name, cfg.Pass)

	db, err := sqlx.Open("clickhouse", url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			return nil, fmt.Errorf("[%d] %s \n%s", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}

	if err := migrateDB(db.DB); err != nil {
		return nil, err
	}

	return db, nil
}

func migrateDB(db *sql.DB) error {
	driver, err := migrateClickhouse.WithInstance(db, &migrateClickhouse.Config{})
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"clickhouse",
		driver)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("UP: %v", err)
	}
	return err
}

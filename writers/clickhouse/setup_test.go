// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package clickhouse_test contains tests for Clickhouse repository
// implementations.
package clickhouse_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/writers/clickhouse"
	dockertest "gopkg.in/ory/dockertest.v3"
)

var db *sqlx.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("yandex/clickhouse-server", "20.1", []string{})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	port := container.GetPort("9000/tcp")

	if err := pool.Retry(func() error {
		url := fmt.Sprintf("tcp://localhost:%s?username=default&password=&debug=true", port)

		db, err = sqlx.Open("clickhouse", url)
		if err != nil {
			return err
		}
		if err := db.Ping(); err != nil {
			return err
		}
		if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS test"); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	dbConfig := clickhouse.Config{
		Host: "localhost",
		Port: port,
		User: "default",
		Pass: "",
		Name: "test",
	}

	db, err = clickhouse.Connect(dbConfig)
	if err != nil {
		log.Fatalf("Could not setup test DB connection: %s", err)
	}

	code := m.Run()

	// defers will not be run when using os.Exit
	db.Close()
	if err := pool.Purge(container); err != nil {
		log.Fatalf("Could not purge container: %s", err)
	}

	os.Exit(code)
}

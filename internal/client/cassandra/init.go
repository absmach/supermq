// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"github.com/gocql/gocql"
)

// Config contains Cassandra DB specific parameters.
type Config struct {
	Hosts    []string `env:"DB_CLUSTER"     envDefault:"127.0.0.1" envSeparator:","`
	Keyspace string   `env:"DB_KEYSPACE"    envDefault:"mainflux"`
	User     string   `env:"DB_USER"        envDefault:"mainflux"`
	Pass     string   `env:"DB_PASS"        envDefault:"mainflux"`
	Port     int      `env:"DB_PORT"        envDefault:"9042"`
}

// Connect establishes connection to the Cassandra cluster.
func Connect(cfg Config) (*gocql.Session, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.Keyspace = cfg.Keyspace
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cfg.User,
		Password: cfg.Pass,
	}
	cluster.Port = cfg.Port

	return cluster.CreateSession()
}

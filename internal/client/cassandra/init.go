// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"github.com/gocql/gocql"
)

// DBConfig contains Cassandra DB specific parameters.
type DBConfig struct {
	Hosts    []string `env:"DB_CLUSTER"     default:"127.0.0.1" envSeparator:","`
	Keyspace string   `env:"DB_KEYSPACE"    default:"mainflux"`
	User     string   `env:"DB_USER"        default:"mainflux"`
	Pass     string   `env:"DB_PASS"        default:"mainflux"`
	Port     int      `env:"DB_PORT"        default:"9042"`
}

// Connect establishes connection to the Cassandra cluster.
func Connect(cfg DBConfig) (*gocql.Session, error) {
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

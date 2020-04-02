// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mainflux

import (
	"os"

	"github.com/nats-io/nats.go"
)

const (
	// DefNatsURL NATS broker URL
	DefNatsURL = nats.DefaultURL
	// DefLogLevelError default loggger error flag
	DefLogLevelError = "error"
	// DefDBHost default DB Host
	DefDBHost = "localhost"
	// DefDBName default DB Name
	DefDBName = "mainflux"
	// DefUsersDBName default Users DB Name
	DefUsersDBName = "users"
	// DefAuthnDBName default Authn DB Name
	DefAuthnDBName = "authn"
	// DefDBUser default DB User
	DefDBUser = "mainflux"
	// DefDBPass default DB Password
	DefDBPass = "mainflux"
	// DefRedisURL Redis service URL
	DefRedisURL = "localhost:6379"
	// DefAuthnURL AuthN service gRPC URL
	DefAuthnURL = "localhost:8181"
	// DefThingsHTTPPort Things service HTTP Port
	DefThingsHTTPPort = "8180"
	// DefThingsAuthHTTPPort Things service Auth HTTP Port
	DefThingsAuthHTTPPort = "8989"
	// DefThingsAuthGRPCPort Things service Auth gRPC Port
	DefThingsAuthGRPCPort = "8183"
	// DefThingsAuthURL Things service Auth gRPC URL
	DefThingsAuthURL = "localhost:8183"
	// DefHTTPPort HTTP adapter Port
	DefHTTPPort = "8185"
	// DefCoapPort CoAP adapter Port
	DefCoapPort = "5683"
	// DefTwinsHTTPort Twins HTTP Port
	DefTwinsHTTPort = "9021"
	// DefWSPort WS Port
	DefWSPort = "8186"
	// DefInfluxWriterPort infuxdb-writer HTTP Port
	DefInfluxWriterPort = "8900"
	// DefInfluxWriterDBPort influxdb-writer DB Port
	DefInfluxWriterDBPort = "8086"
	// DefMongoWriterPort mmongo-writer HTTP Port
	DefMongoWriterPort = "8901"
	// DefMongoDBPort mongo-writer DB Port
	DefMongoDBPort = "27017"
	// DefCassandraWriterPort cassandra-writer HTTP Port
	DefCassandraWriterPort = "8902"
	// DefCassandraWriterDBPort cassandra-writer DB Port
	DefCassandraWriterDBPort = "9042"
	// DefPostgresWriterPort postgres-writer HTTP Port
	DefPostgresWriterPort = "9204"
	// DefPostgresDBPort Postgrres DB Port
	DefPostgresDBPort = "5432"
)

// Env reads specified environment variable. If no value has been found,
// fallback is returned.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

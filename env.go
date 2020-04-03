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
	// DefDBUser default DB User
	DefDBUser = "mainflux"
	// DefDBPass default DB Password
	DefDBPass = "mainflux"
	// DefUsersDB default users DB Name
	DefUsersDB = "users"
	// DefAuthnDB default authn DB Name
	DefAuthnDB = "authn"
	// DefThingsDB default things DB Name
	DefThingsDB = "things"
	// DefBootstrapDB default bootstrap DB Name
	DefBootstrapDB = "bootstrap"
	// DefWritersDBName default messages DB Name
	DefWritersDBName = "writer"
	// DefTwinsDB default twins DB Name
	DefTwinsDB = "mainflux-twins"

	// DefRedisURL Redis service URL
	DefRedisURL = "localhost:6379"
	// DefAuthnURL AuthN service gRPC URL
	DefAuthnURL = "localhost:8181"
	// DefAuthnGRPCPort AuthN service gRPC Port
	DefAuthnGRPCPort = "8181"
	// DefAuthnHTTPPort AuthN service HTTP Port
	DefAuthnHTTPPort = "8189"
	// DefUsersHTTPPort Users service HTTP Port
	DefUsersHTTPPort = "8180"
	// DefThingsHTTPPort Things service HTTP Port
	DefThingsHTTPPort = "8182"

	// DefThingsAuthHTTPPort Things service Auth HTTP Port
	DefThingsAuthHTTPPort = "8989"
	// DefThingsAuthGRPCPort Things service Auth gRPC Port
	DefThingsAuthGRPCPort = "8183"
	// DefBootstrapPort service HTTP Port
	DefBootstrapPort = "8202"
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
	// DefLoraHTTPPort Lora adapter HTTP Port
	DefLoraHTTPPort = "8187"

	// DefInfluxReaderPort infuxdb-reader HTTP Port
	DefInfluxReaderPort = "8905"
	// DefInfluxWriterPort infuxdb-writer HTTP Port
	DefInfluxWriterPort = "8900"
	// DefMongoReaderPort mongo-reader HTTP Port
	DefMongoReaderPort = "8904"
	// DefMongoWriterPort mongo-writer HTTP Port
	DefMongoWriterPort = "8901"
	// DefCassandraReaderPort cassandra-reader HTTP Port
	DefCassandraReaderPort = "8903"
	// DefCassandraWriterPort cassandra-writer HTTP Port
	DefCassandraWriterPort = "8902"
	// DefPostgresWriterPort postgres-writer HTTP Port
	DefPostgresWriterPort = "9204"

	// DefInfluxDBPort Influx DB Port
	DefInfluxDBPort = "8086"
	// DefMongoDBPort Mongo DB Port
	DefMongoDBPort = "27017"
	// DefCassandraDBPort Casssandra DB Port
	DefCassandraDBPort = "9042"
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

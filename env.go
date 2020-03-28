// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mainflux

import (
	"os"
)

const (
	// DefDBHost default DB Host
	DefDBHost = "localhost"
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
)

// Env reads specified environment variable. If no value has been found,
// fallback is returned.
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

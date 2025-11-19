// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/redis/go-redis/v9"
	"github.com/redis/go-redis/v9/maintnotifications"
)

var (
	storeClient *redis.Client
	storeURL    string
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("docker.io/redis", "8.2.2-alpine3.22", nil)
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	storeURL = fmt.Sprintf("redis://localhost:%s/0", container.GetPort("6379/tcp"))
	opts, err := redis.ParseURL(storeURL)
	if err != nil {
		log.Fatalf("Could not parse redis URL: %s", err)
	}
	opts.MaintNotificationsConfig = &maintnotifications.Config{
		Mode: maintnotifications.ModeDisabled,
	}

	if err := pool.Retry(func() error {
		storeClient = redis.NewClient(opts)

		return storeClient.Ping(context.Background()).Err()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(container); err != nil {
		log.Fatalf("Could not purge container: %s", err)
	}

	os.Exit(code)
}

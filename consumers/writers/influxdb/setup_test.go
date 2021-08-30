// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	influxdb "github.com/influxdata/influxdb-client-go/v2"
	dockertest "github.com/ory/dockertest/v3"
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		testLog.Error(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	cfg := []string{
		"DOCKER_INFLUXDB_INIT_MODE=setup",
		"DOCKER_INFLUXDB_INIT_USERNAME=test",
		"DOCKER_INFLUXDB_INIT_PASSWORD=mainflux",
		"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=test",
		"DOCKER_INFLUXDB_INIT_ORG=test",
		"DOCKER_INFLUXDB_INIT_BUCKET=test",
	}
	container, err := pool.Run("influxdb", "2.0", cfg)
	if err != nil {
		testLog.Error(fmt.Sprintf("Could not start container: %s", err))
	}

	hostPort = container.GetHostPort("8086/tcp")
	Addr := fmt.Sprintf("http://%s", hostPort)

	if err := pool.Retry(func() error {
		client = influxdb.NewClient(Addr, "test")
		ctx := context.Background()
		ready, err := client.Ready(ctx)
		if err != nil || !ready {
			testLog.Error(fmt.Sprintf("Not ready: %s", err))
			return err
		}
		testLog.Info(fmt.Sprintf("Successfully started docker: %s", client.ServerURL()))
		return nil
	}); err != nil {
		testLog.Error(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	code := m.Run()

	if err := pool.Purge(container); err != nil {
		testLog.Error(fmt.Sprintf("Could not purge container: %s", err))
	}

	os.Exit(code)
}

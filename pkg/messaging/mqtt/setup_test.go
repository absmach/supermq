// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/mqtt"
	dockertest "github.com/ory/dockertest/v3"
)

const (
	protocol = "mqtt"
	id       = "mqtt-publisher"
	timeout  = time.Second
	qos      = 1
)

var (
	//mqttClient mqtt.Client
	publisher  messaging.Publisher
	subscriber messaging.Subscriber
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("vernemq/vernemq", "latest-alpine", []string{"DOCKER_VERNEMQ_ACCEPT_EULA=yes", "DOCKER_VERNEMQ_ALLOW_ANONYMOUS=on"})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	address := fmt.Sprintf("%s:%s", "localhost", container.GetPort("1883/tcp"))
	if err := pool.Retry(func() error {
		publisher, err = mqtt.NewPublisher(address, timeout)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	logger, err := logger.New(os.Stdout, "error")
	if err != nil {
		log.Fatalf(err.Error())
	}
	if err := pool.Retry(func() error {
		subscriber, err = mqtt.NewSubscriber(address, timeout, logger)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	if err := pool.Purge(container); err != nil {
		log.Fatalf("Could not purge container: %s", err)
	}

	os.Exit(code)
}

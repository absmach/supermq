// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !rabbitmq
// +build !rabbitmq

package nats_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/mainflux/mainflux/pkg/events/nats"
	"github.com/ory/dockertest/v3"
)

var (
	redisURL string
	stream   = "tests.events"
	consumer = "tests-consumer"
	ctx      = context.Background()
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "nats",
		Tag:        "2.9.21-alpine",
		Cmd:        []string{"-DVV", "-js"},
	})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	handleInterrupt(pool, container)

	address := fmt.Sprintf("%s:%s", "localhost", container.GetPort("4222/tcp"))

	if err := pool.Retry(func() error {
		_, err = nats.NewPublisher(ctx, address, stream)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if err := pool.Retry(func() error {
		_, err = nats.NewSubscriber(ctx, address, stream, consumer, logger)
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

func handleInterrupt(pool *dockertest.Pool, container *dockertest.Resource) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		if err := pool.Purge(container); err != nil {
			log.Fatalf("Could not purge container: %s", err)
		}
		os.Exit(0)
	}()
}

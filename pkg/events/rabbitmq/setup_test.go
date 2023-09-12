// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build rabbitmq && test
// +build rabbitmq,test

package rabbitmq_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/mainflux/mainflux/pkg/events/rabbitmq"
	"github.com/ory/dockertest/v3"
)

var (
	rabbitmqURL string
	stream      = "tests.events"
	consumer    = "tests-consumer"
	ctx         = context.Background()
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("rabbitmq", "3.9.20", []string{})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}

	handleInterrupt(pool, container)

	rabbitmqURL = fmt.Sprintf("amqp://%s:%s", "localhost", container.GetPort("5672/tcp"))

	fmt.Printf("address: %s\n", rabbitmqURL)
	if err := pool.Retry(func() error {
		_, err = rabbitmq.NewPublisher(ctx, rabbitmqURL, stream)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if err := pool.Retry(func() error {
		_, err = rabbitmq.NewSubscriber(rabbitmqURL, stream, consumer, logger)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	fmt.Println("RabbitMQ started")

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

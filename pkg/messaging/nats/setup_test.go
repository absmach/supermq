// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats_test

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	mainflux_log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	broker "github.com/nats-io/nats.go"
	dockertest "github.com/ory/dockertest/v3"
)

const (
	port          = "4222/tcp"
	brokerName    = "nats"
	brokerVersion = "1.3.0"
)

var (
	pubsub  messaging.PubSub
	address string
	logger  mainflux_log.Logger
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run(brokerName, brokerVersion, []string{})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}
	handleInterrupt(pool, container)

	address = fmt.Sprintf("%s:%s", "localhost", container.GetPort(port))

	logger, err = mainflux_log.New(os.Stdout, mainflux_log.Debug.String())
	if err != nil {
		log.Fatalf(err.Error())
	}
	if err := pool.Retry(func() error {
		pubsub, err = nats.NewPubSub(address, "", logger)
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

func newConn() (*broker.Conn, error) {
	conn, err := broker.Connect(address)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func handleInterrupt(pool *dockertest.Pool, container *dockertest.Resource) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := pool.Purge(container); err != nil {
			log.Fatalf("Could not purge container: %s", err)
		}
		os.Exit(0)
	}()
}

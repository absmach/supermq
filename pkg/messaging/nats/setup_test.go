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
	"time"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	broker "github.com/nats-io/nats.go"
	dockertest "github.com/ory/dockertest/v3"
)

const (
	port          = "4222/tcp"
	brokerName    = "nats"
	brokerVersion = "1.3.0"
	queue         = "mainflux-nats"
)

var (
	pubsub      messaging.PubSub
	queuePubsub messaging.PubSub
	address     string
	logger      mflog.Logger
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

	logger, err = mflog.New(os.Stdout, mflog.Debug.String())
	if err != nil {
		log.Fatalf(err.Error())
	}
	if err := pool.Retry(func() error {
		pubsub, err = nats.NewPubSub(address, "", logger)
		return err
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	if err := pool.Retry(func() error {
		queuePubsub, err = nats.NewPubSub(address, queue, logger)
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
	// opts := broker.Options{
	// 	Url:              address,
	// 	AllowReconnect:   true,
	// 	MaxReconnect:     10,
	// 	ReconnectWait:    200 * time.Millisecond,
	// 	Timeout:          1 * time.Second,
	// 	ReconnectBufSize: 5 * 1024 * 1024,
	// 	PingInterval:     1 * time.Second,
	// 	MaxPingsOut:      5,
	// }
	opts := broker.Options{
		Url:            address,
		AllowReconnect: false,
		// MaxReconnect:     10,
		// ReconnectWait:    200 * time.Millisecond,
		Timeout: 1 * time.Second,
		// ReconnectBufSize: 5 * 1024 * 1024,
		// PingInterval:     1 * time.Second,
		// MaxPingsOut:      5,
	}

	conn, err := opts.Connect()
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

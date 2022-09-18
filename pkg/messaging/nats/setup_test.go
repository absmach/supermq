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

	"github.com/gogo/protobuf/proto"
	logg "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	broker "github.com/nats-io/nats.go"
	dockertest "github.com/ory/dockertest/v3"
)

var (
	pubsub  messaging.PubSub
	address string
	logger  logg.Logger
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	container, err := pool.Run("nats", "1.3.0", []string{})
	if err != nil {
		log.Fatalf("Could not start container: %s", err)
	}
	handleInterrupt(pool, container)

	address = fmt.Sprintf("%s:%s", "localhost", container.GetPort("4222/tcp"))

	logger, err = logg.New(os.Stdout, "error")
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

func natsHandler(h messaging.MessageHandler) broker.MsgHandler {
	fmt.Println()
	fmt.Println("Reached setup_test nats handler")
	fmt.Println()

	return func(m *broker.Msg) {
		var msg messaging.Message
		if err := proto.Unmarshal(m.Data, &msg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			fmt.Println()
			fmt.Println("error while unmarshal in natsHandler ->", err)
			fmt.Println()
			return
		}
		fmt.Println()
		fmt.Println("reached h.handle")
		fmt.Println()
		if err := h.Handle(msg); err != nil {
			fmt.Println()
			fmt.Println("h.handle -> error ->", err)
			fmt.Println()
			logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
		}
		fmt.Println()
		fmt.Println("back from h.handle")
		fmt.Println()
	}
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

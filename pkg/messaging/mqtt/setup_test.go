// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	logg "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	mqtt_pubsub "github.com/mainflux/mainflux/pkg/messaging/mqtt"
)

var (
	publisher messaging.Publisher
	pubsub    messaging.PubSub
	client    mqtt.Client

	logger logg.Logger

	address = "tcp://mqtt.eclipseprojects.io:1883"
)

const (
	username = "mainflux-mqtt"
	qos      = 2
)

func TestMain(m *testing.M) {

	// pool, err := dockertest.NewPool("")
	// if err != nil {
	// 	log.Fatalf("Could not connect to docker: %s", err)
	// }

	// // container, err := pool.Run("eclipse-mosquitto", "1.6.13", nil)
	// container, err := pool.Run("mainflux/vernemq", "0.13.0", []string{"DOCKER_VERNEMQ_ALLOW_ANONYMOUS=on", "ERL_MAX_ETS_TABLES=256000", "ERL_CRASH_DUMP=/erl_crash.dump", "ERL_FULLSWEEP_AFTER=0", "ERL_MAX_PORTS=256000"})
	// if err != nil {
	// 	log.Fatalf("Could not start container: %s", err)
	// }
	// handleInterrupt(m, pool, container)

	// // address = fmt.Sprintf("%s:%s", "localhost", container.GetPort("1883/tcp"))
	// // pool.MaxWait = 120 * time.Second
	// if err := pool.Retry(func() error {
	// 	publisher, err = mqtt_pubsub.NewPublisher(address, 30*time.Second)
	// 	return err
	// }); err != nil {
	// 	log.Fatalf("Could not connect to docker: %s", err)
	// }

	// logger, err = logg.New(os.Stdout, "error")
	// if err != nil {
	// 	log.Fatalf(err.Error())
	// }
	// if err := pool.Retry(func() error {
	// 	pubsub, err = mqtt_pubsub.NewPubSub(address, "", 30*time.Second, logger)
	// 	return err
	// }); err != nil {
	// 	log.Fatalf("Could not connect to docker: %s", err)
	// }

	// code := m.Run()
	// if err := pool.Purge(container); err != nil {
	// 	log.Fatalf("Could not purge container: %s", err)
	// }

	// os.Exit(code)

	logger, err := logg.New(os.Stdout, "error")
	if err != nil {
		log.Fatalf(err.Error())
	}
	publisher, err = mqtt_pubsub.NewPublisher(address, 3*time.Second)
	if err != nil {
		log.Fatalf(err.Error())
	}
	pubsub, err = mqtt_pubsub.NewPubSub(address, "", 3*time.Second, logger)
	if err != nil {
		log.Fatalf(err.Error())
	}

	client, err = newClient(address, "", 30*time.Second)
	if err != nil {
		log.Fatalf(err.Error())
	}

	defer func() {
		err = publisher.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
		err = pubsub.Close()
		if err != nil {
			log.Fatalf(err.Error())
		}
	}()

	code := m.Run()
	os.Exit(code)

}

// func handleInterrupt(m *testing.M, pool *dockertest.Pool, container *dockertest.Resource) {
// 	c := make(chan os.Signal, 2)
// 	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
// 	go func() {
// 		<-c
// 		if err := pool.Purge(container); err != nil {
// 			log.Fatalf("Could not purge container: %s", err)
// 		}
// 		os.Exit(0)
// 	}()
// }

func mqttHandler(h messaging.MessageHandler) mqtt.MessageHandler {
	return func(c mqtt.Client, m mqtt.Message) {
		var msg messaging.Message
		msg.Payload = m.Payload()
		// if err := proto.Unmarshal(m.Payload(), &msg); err != nil {
		// 	logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
		// 	return
		// }
		if err := h.Handle(msg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
		}
	}
}

func newClient(address, id string, timeout time.Duration) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		SetUsername(username).
		AddBroker(address).
		SetClientID(id).
		SetDefaultPublishHandler(mqttHandler(handler{false}))

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Error() != nil {
		return nil, token.Error()
	}

	ok := token.WaitTimeout(timeout)
	if !ok {
		return nil, mqtt_pubsub.ErrConnect
	}

	if token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

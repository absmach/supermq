// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package nats contains NATS message publisher implementation.
package nats

import (
	"context"
	"fmt"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/brokers"
	"github.com/mainflux/mainflux/logger"
	"github.com/nats-io/nats.go"
)

const prefix = "channel"

var _ brokers.MessagePublisher = (*natsPub)(nil)

type natsPub struct {
	conn *nats.Conn
}

// NewPublisher NATS message publisher.
func NewPublisher(url string, log logger.Logger) brokers.MessagePublisher {
	nc, err := nats.Connect(url)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}

	return &natsPub{conn: nc}
}

func (np *natsPub) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", prefix, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	return np.conn.Publish(subject, data)
}

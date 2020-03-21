// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package nats contains NATS message publisher implementation.
package nats

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/errors"
	"github.com/nats-io/nats.go"
)

// Publisher specifies a message publishing API.
type Publisher interface {
	// Publish publishes message to the msessage broker.
	Publish(context.Context, string, mainflux.Message) error

	// Conn returns NATS connection.
	Conn() *nats.Conn
}

var errNatsConn = errors.New("Failed to connect to NATS")

const prefix = "channel"

var _ Publisher = (*pub)(nil)

type pub struct {
	conn *nats.Conn
}

// NewPublisher NATS message publisher.
func NewPublisher(url string) (Publisher, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, errors.Wrap(errNatsConn, err)
	}

	return &pub{conn: nc}, nil
}

func (p pub) Publish(_ context.Context, _ string, msg mainflux.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", prefix, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	return p.conn.Publish(subject, data)
}

func (p pub) Conn() *nats.Conn {
	return p.conn
}

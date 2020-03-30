// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package nats contains NATS message publisher implementation.
package nats

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux/broker"
	"github.com/mainflux/mainflux/errors"
	"github.com/nats-io/nats.go"
)

const prefix = "channel"

var errNatsConn = errors.New("Failed to connect to NATS")

var _ broker.Nats = (*pubsub)(nil)

type pubsub struct {
	conn *nats.Conn
}

// New returns NATS message publisher.
func New(url string) (broker.Nats, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, errors.Wrap(errNatsConn, err)
	}

	return &pubsub{
		conn: nc,
	}, nil
}

func (ps pubsub) Publish(_ context.Context, _ string, msg broker.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", prefix, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	return ps.conn.Publish(subject, data)
}

func fmtSubject(chanID, subtopic string) string {
	subject := fmt.Sprintf("%s.%s", prefix, chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func (ps pubsub) Subscribe(chanID, subtopic string, f func(msg *nats.Msg)) (*nats.Subscription, error) {
	subject := fmtSubject(chanID, subtopic)
	sub, err := ps.conn.Subscribe(subject, f)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (ps pubsub) Close() {
	ps.conn.Close()
}

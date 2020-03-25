// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package nats contains NATS message publisher implementation.
package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// Subscriber specifies a message subscribing API.
type Subscriber interface {
	// Subscribe subscribes to the message broker for a given channel ID and subtopic.
	Subscribe(string, string, func(msg *nats.Msg)) (*nats.Subscription, error)

	// Close closes NATS connection.
	Close()
}

var _ Subscriber = (*sub)(nil)

type sub struct {
	conn *nats.Conn
}

// NewSubscriber instantiates NATS message publisher.
func NewSubscriber(url string) (Subscriber, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &sub{conn: nc}, nil
}

func fmtSubject(chanID, subtopic string) string {
	subject := fmt.Sprintf("%s.%s", prefix, chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func (s sub) Subscribe(chanID, subtopic string, f func(msg *nats.Msg)) (*nats.Subscription, error) {
	subject := fmtSubject(chanID, subtopic)
	sub, err := s.conn.Subscribe(subject, f)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s sub) Close() {
	s.conn.Close()
}

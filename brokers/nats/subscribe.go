// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package nats contains NATS message publisher implementation.
package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// NatsSubscriber specifies a message subscribing API.
type NatsSubscriber interface {
	// Subscribe subscribes to the message broker for a given channel ID and subtopic.
	Subscribe(string, string, func(msg *nats.Msg)) (*nats.Subscription, error)

	SubConn() *nats.Conn
}

var _ NatsSubscriber = (*natsSub)(nil)

type natsSub struct {
	conn *nats.Conn
}

// NewSubscriber instantiates NATS message publisher.
func NewSubscriber(url string) (NatsSubscriber, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &natsSub{conn: nc}, nil
}

func fmtSubject(chanID, subtopic string) string {
	subject := fmt.Sprintf("%s.%s", prefix, chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func (ns *natsSub) Subscribe(chanID, subtopic string, f func(msg *nats.Msg)) (*nats.Subscription, error) {
	subject := fmtSubject(chanID, subtopic)
	sub, err := ns.conn.Subscribe(subject, f)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (ns *natsSub) SubConn() *nats.Conn {
	return ns.conn
}

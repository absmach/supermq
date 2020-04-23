// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	broker "github.com/nats-io/nats.go"
)

const chansPrefix = "channels"

// SubjectAllChannels represents subject to subscribe for all the channels.
const SubjectAllChannels = "channels.>"

var errNotSubscribed = errors.New("not subscribed")

var _ mainflux.PubSub = (*nats)(nil)

type nats struct {
	conn         *broker.Conn
	subscription *broker.Subscription
	logger       log.Logger
	mu           sync.Mutex
	subject      string
	queue        string
}

// New returns NATS message broker.
func New(conn *broker.Conn, subject, queue string, logger log.Logger) mainflux.PubSub {
	return &nats{
		conn:    conn,
		queue:   queue,
		logger:  logger,
		subject: subject,
	}
}

func (n *nats) Publish(msg mainflux.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", chansPrefix, msg.Channel)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	if err := n.conn.Publish(subject, data); err != nil {
		return err
	}

	return nil
}

func (n *nats) Subscribe(subHandler mainflux.SubscribeHandler) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	sub, err := n.subscribe(subHandler)
	if err != nil {
		return err
	}
	n.subscription = sub
	return nil
}

func (n *nats) Unsubscribe() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.subscription == nil {
		return errNotSubscribed
	}
	if err := n.subscription.Unsubscribe(); err != nil {
		return err
	}
	n.subscription = nil
	return nil
}

func (n *nats) subscribe(subHandler mainflux.SubscribeHandler) (*broker.Subscription, error) {
	if n.queue != "" {
		return n.conn.QueueSubscribe(n.subject, n.queue, n.natsHandler(subHandler))
	}
	ps := SubjectAllChannels
	if n.subject != "" {
		ps = fmt.Sprintf("%s.%s", chansPrefix, n.subject)
	}
	return n.conn.Subscribe(ps, n.natsHandler(subHandler))
}

func (n *nats) natsHandler(h mainflux.SubscribeHandler) broker.MsgHandler {
	return func(m *broker.Msg) {
		var msg mainflux.Message
		if err := proto.Unmarshal(m.Data, &msg); err != nil {
			n.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h(msg); err != nil {
			n.logger.Warn(fmt.Sprintf("Failed handle Mainflux message: %s", err))
		}
	}
}

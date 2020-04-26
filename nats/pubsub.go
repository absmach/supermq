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

var (
	errAlreadySubscribed = errors.New("already subscribed to topic")
	errNotSubscribed     = errors.New("not subscribed")
)

var _ mainflux.PubSub = (*nats)(nil)

type nats struct {
	conn          *broker.Conn
	subscriptions map[string]*broker.Subscription
	logger        log.Logger
	mu            sync.Mutex
	queue         string
}

// New returns NATS message broker.
func New(conn *broker.Conn, queue string, logger log.Logger) mainflux.PubSub {
	return &nats{
		conn:   conn,
		queue:  queue,
		logger: logger,
	}
}

func (n *nats) Publish(topic string, msg mainflux.Message) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", chansPrefix, topic)
	if msg.Subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, msg.Subtopic)
	}
	if err := n.conn.Publish(subject, data); err != nil {
		return err
	}

	return nil
}

func (n *nats) Subscribe(topic string, handler mainflux.MessageHandler) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.subscriptions[topic]; ok {
		return errAlreadySubscribed
	}
	s := SubjectAllChannels
	if topic != "" {
		s = fmt.Sprintf("%s.%s", chansPrefix, topic)
	}
	if n.queue != "" {
		sub, err := n.conn.QueueSubscribe(s, n.queue, n.natsHandler(handler))
		if err != nil {
			return err
		}
		n.subscriptions[s] = sub
		return nil
	}
	sub, err := n.conn.Subscribe(s, n.natsHandler(handler))
	if err != nil {
		return err
	}
	n.subscriptions[s] = sub
	return nil
}

func (n *nats) Unsubscribe(topic string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	sub, ok := n.subscriptions[topic]
	if !ok {
		return errNotSubscribed
	}

	if err := sub.Unsubscribe(); err != nil {
		return err
	}

	delete(n.subscriptions, topic)
	return nil
}

// func (n *nats) subscribe(topic string, handler mainflux.MessageHandler) (*broker.Subscription, error) {
// 	if n.queue != "" {
// 		return n.conn.QueueSubscribe(topic, n.queue, n.natsHandler(handler))
// 	}
// 	ps := SubjectAllChannels
// 	if topic != "" {
// 		ps = fmt.Sprintf("%s.%s", chansPrefix, topic)
// 	}
// 	return n.conn.Subscribe(ps, n.natsHandler(handler))
// }

func (n *nats) natsHandler(h mainflux.MessageHandler) broker.MsgHandler {
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

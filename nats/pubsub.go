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
	errEmptyTopic        = errors.New("empty topic")
)

var _ mainflux.PubSub = (*nats)(nil)

type nats struct {
	conn          *broker.Conn
	logger        log.Logger
	mu            sync.Mutex
	queue         string
	subscriptions map[string]*broker.Subscription
}

// New returns NATS message broker.
// Paramter queue specifies the queue for the Subscribe method. If queue is specified (is not an empty string),
// Subscribe method will execute NATS QueueSubscibe which is conceptually different from ordinary subscribe.
// For more information, please take a look here: https://docs.nats.io/developing-with-nats/receiving/queues.
// If the queue is empty, Subscribe will be used.
func New(conn *broker.Conn, queue string, logger log.Logger) mainflux.PubSub {
	return &nats{
		conn:          conn,
		queue:         queue,
		logger:        logger,
		subscriptions: make(map[string]*broker.Subscription),
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
	if topic == "" {
		return errEmptyTopic
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if _, ok := n.subscriptions[topic]; ok {
		return errAlreadySubscribed
	}
	topic = fmt.Sprintf("%s.%s", chansPrefix, topic)
	if n.queue != "" {
		sub, err := n.conn.QueueSubscribe(topic, n.queue, n.natsHandler(handler))
		if err != nil {
			return err
		}
		n.subscriptions[topic] = sub
		return nil
	}
	sub, err := n.conn.Subscribe(topic, n.natsHandler(handler))
	if err != nil {
		return err
	}
	n.subscriptions[topic] = sub
	return nil
}

func (n *nats) Unsubscribe(topic string) error {
	if topic == "" {
		return errEmptyTopic
	}
	n.mu.Lock()
	defer n.mu.Unlock()

	topic = fmt.Sprintf("%s.%s", chansPrefix, topic)

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

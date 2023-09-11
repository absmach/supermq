// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"sync"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

const (
	chansPrefix = "channels"
	// SubjectAllChannels represents subject to subscribe for all the channels.
	SubjectAllChannels = "channels.#"
	exchangeName       = "messages"
)

var (
	// ErrNotSubscribed indicates that the topic is not subscribed to.
	ErrNotSubscribed = errors.New("not subscribed")

	// ErrEmptyTopic indicates the absence of topic.
	ErrEmptyTopic = errors.New("empty topic")

	// ErrEmptyID indicates the absence of ID.
	ErrEmptyID = errors.New("empty ID")
)
var _ messaging.PubSub = (*Rpubsub)(nil)

type subscription struct {
	cancel func() error
}
type Rpubsub struct {
	RPublisher
	logger        mflog.Logger
	subscriptions map[string]map[string]subscription
	mu            sync.Mutex
}

// NewPubSub returns RabbitMQ message publisher/subscriber.
func NewPubSub(url string, logger mflog.Logger, opts ...messaging.Option) (messaging.PubSub, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.ExchangeDeclare(exchangeName, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		return nil, err
	}

	ret := &Rpubsub{
		RPublisher: RPublisher{
			Conn:    conn,
			Channel: ch,
		},
		logger:        logger,
		subscriptions: make(map[string]map[string]subscription),
	}

	for _, opt := range opts {
		if err := opt(ret); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func (ps *Rpubsub) Subscribe(ctx context.Context, cfg messaging.SubscriberConfig) error {
	if cfg.ID == "" {
		return ErrEmptyID
	}
	if cfg.Topic == "" {
		return ErrEmptyTopic
	}
	ps.mu.Lock()

	cfg.Topic = formatTopic(cfg.Topic)
	// Check topic
	s, ok := ps.subscriptions[cfg.Topic]
	if ok {
		// Check client ID
		if _, ok := s[cfg.ID]; ok {
			// Unlocking, so that Unsubscribe() can access ps.subscriptions
			ps.mu.Unlock()
			if err := ps.Unsubscribe(ctx, cfg.ID, cfg.Topic); err != nil {
				return err
			}

			ps.mu.Lock()
			// value of s can be changed while ps.mu is unlocked
			s = ps.subscriptions[cfg.Topic]
		}
	}
	defer ps.mu.Unlock()
	if s == nil {
		s = make(map[string]subscription)
		ps.subscriptions[cfg.Topic] = s
	}

	clientID := fmt.Sprintf("%s-%s", cfg.Topic, cfg.ID)

	queue, err := ps.Channel.QueueDeclare(clientID, true, false, false, false, nil)
	if err != nil {
		return err
	}

	if err := ps.Channel.QueueBind(queue.Name, cfg.Topic, exchangeName, false, nil); err != nil {
		return err
	}

	msgs, err := ps.Channel.Consume(queue.Name, clientID, true, false, false, false, nil)
	if err != nil {
		return err
	}
	go ps.handle(msgs, cfg.Handler)
	s[cfg.ID] = subscription{
		cancel: func() error {
			if err := ps.Channel.Cancel(clientID, false); err != nil {
				return err
			}
			return cfg.Handler.Cancel()
		},
	}

	return nil
}

func (ps *pubsub) Unsubscribe(ctx context.Context, id, topic string) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	topic = formatTopic(topic)
	// Check topic
	s, ok := ps.subscriptions[topic]
	if !ok {
		return ErrNotSubscribed
	}
	// Check topic ID
	current, ok := s[id]
	if !ok {
		return ErrNotSubscribed
	}
	if current.cancel != nil {
		if err := current.cancel(); err != nil {
			return err
		}
	}
	if err := ps.ch.QueueUnbind(topic, topic, exchangeName, nil); err != nil {
		return err
	}

	delete(s, id)
	if len(s) == 0 {
		delete(ps.subscriptions, topic)
	}
	return nil
}

func (ps *Rpubsub) handle(deliveries <-chan amqp.Delivery, h messaging.MessageHandler) {
	for d := range deliveries {
		var msg messaging.Message
		if err := proto.Unmarshal(d.Body, &msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h.Handle(&msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
			return
		}
	}
}

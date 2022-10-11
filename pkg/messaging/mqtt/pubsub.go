// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"errors"
	"fmt"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gogo/protobuf/proto"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	username = "mainflux-mqtt"
	qos      = 2
)

var (
	ErrConnect                = errors.New("failed to connect to MQTT broker")
	ErrSubscribeTimeout       = errors.New("failed to subscribe due to timeout reached")
	ErrUnsubscribeTimeout     = errors.New("failed to unsubscribe due to timeout reached")
	ErrUnsubscribeDeleteTopic = errors.New("failed to unsubscribe due to deletion of topic")
	ErrNotSubscribed          = errors.New("not subscribed")
	ErrEmptyTopic             = errors.New("empty topic")
	ErrEmptyID                = errors.New("empty ID")
	ErrFailedHandleMessage    = errors.New("failed to handle mainflux message")
)

var _ messaging.PubSub = (*pubsub)(nil)

type subscription struct {
	client mqtt.Client
	topics []string
	cancel func() error
}

type pubsub struct {
	publisher
	logger        log.Logger
	mu            sync.RWMutex
	address       string
	timeout       time.Duration
	subscriptions map[string]subscription
}

// NewPubSub returns MQTT message publisher/subscriber.
func NewPubSub(url, queue string, timeout time.Duration, logger log.Logger) (messaging.PubSub, error) {
	client, err := newClient(url, "mqtt-publisher", timeout)
	if err != nil {
		return nil, err
	}
	ret := &pubsub{
		publisher: publisher{
			client:  client,
			timeout: timeout,
		},
		address:       url,
		timeout:       timeout,
		logger:        logger,
		subscriptions: make(map[string]subscription),
	}
	return ret, nil
}

func (ps *pubsub) Subscribe(id, topic string, handler messaging.MessageHandler) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	s, ok := ps.subscriptions[id]
	// If the client exists, check if it's subscribed to the topic and unsubscribe if needed.
	switch ok {
	case true:
		if ok := s.contains(topic); ok {
			if err := s.unsubscribe(topic, ps.timeout); err != nil {
				return err
			}
		}
	default:
		client, err := newClient(ps.address, id, ps.timeout)
		if err != nil {
			return err
		}
		s = subscription{
			client: client,
			topics: []string{},
			cancel: handler.Cancel,
		}
	}
	s.topics = append(s.topics, topic)
	ps.subscriptions[id] = s

	token := s.client.Subscribe(topic, qos, ps.mqttHandler(handler))
	if token.Error() != nil {
		return token.Error()
	}
	if ok := token.WaitTimeout(ps.timeout); !ok {
		return ErrSubscribeTimeout
	}
	return token.Error()
}

func (ps *pubsub) Unsubscribe(id, topic string) error {
	if id == "" {
		return ErrEmptyID
	}
	if topic == "" {
		return ErrEmptyTopic
	}
	ps.mu.Lock()
	defer ps.mu.Unlock()

	s, ok := ps.subscriptions[id]
	if !ok || !s.contains(topic) {
		return ErrNotSubscribed
	}

	if err := s.unsubscribe(topic, ps.timeout); err != nil {
		return err
	}
	ps.subscriptions[id] = s

	if len(s.topics) == 0 {
		delete(ps.subscriptions, id)
	}
	return nil
}

func (s *subscription) unsubscribe(topic string, timeout time.Duration) error {
	if s.cancel != nil {
		if err := s.cancel(); err != nil {
			return err
		}
	}

	token := s.client.Unsubscribe(topic)
	if token.Error() != nil {
		return token.Error()
	}

	if ok := token.WaitTimeout(timeout); !ok {
		return ErrUnsubscribeTimeout
	}
	if ok := s.delete(topic); !ok {
		return ErrUnsubscribeDeleteTopic
	}
	return token.Error()
}

func newClient(address, id string, timeout time.Duration) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions().
		SetUsername(username).
		AddBroker(address).
		SetClientID(id).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			time.Sleep(500 * time.Millisecond)
		}).
		SetReconnectingHandler(func(c mqtt.Client, options *mqtt.ClientOptions) {
			time.Sleep(500 * time.Millisecond)
		})
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Error() != nil {
		return nil, token.Error()
	}

	ok := token.WaitTimeout(timeout)
	if !ok {
		return nil, ErrConnect
	}

	if token.Error() != nil {
		return nil, token.Error()
	}

	return client, nil
}

func (ps *pubsub) mqttHandler(h messaging.MessageHandler) mqtt.MessageHandler {
	return func(c mqtt.Client, m mqtt.Message) {
		var msg messaging.Message
		if err := proto.Unmarshal(m.Payload(), &msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		if err := h.Handle(msg); err != nil {
			ps.logger.Warn(fmt.Sprintf("Failed to handle Mainflux message: %s", err))
		}
	}
}

// contains checks if a topic is present
func (sub subscription) contains(topic string) bool {
	return sub.indexOf(topic) != -1
}

// Finds the index of an item in the topics
func (sub subscription) indexOf(element string) int {
	for k, v := range sub.topics {
		if element == v {
			return k
		}
	}
	return -1
}

// Deletes a topic from the slice
func (sub *subscription) delete(topic string) bool {
	index := sub.indexOf(topic)
	if index == -1 {
		return false
	}
	topics := make([]string, len(sub.topics)-1)
	copy(topics[:index], sub.topics[:index])
	copy(topics[index:], sub.topics[index+1:])
	sub.topics = topics
	return true
}

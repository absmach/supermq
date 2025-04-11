// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/messaging"
	broker "github.com/absmach/supermq/pkg/messaging/rabbitmq"
)

var _ events.Subscriber = (*subEventStore)(nil)

var (
	exchangeName = "events"
	eventsPrefix = "events"

	// ErrEmptyStream is returned when stream name is empty.
	ErrEmptyStream = errors.New("stream name cannot be empty")

	// ErrEmptyConsumer is returned when consumer name is empty.
	ErrEmptyConsumer = errors.New("consumer name cannot be empty")
)

type subEventStore struct {
	pubsub messaging.PubSub
}

func NewSubscriber(url string, logger *slog.Logger) (events.Subscriber, error) {
	pubsub, err := broker.NewPubSub(url, logger, broker.Prefix(eventsPrefix), broker.Exchange(exchangeName))
	if err != nil {
		return nil, err
	}

	return &subEventStore{
		pubsub: pubsub,
	}, nil
}

func (es *subEventStore) Subscribe(ctx context.Context, cfg events.SubscriberConfig) error {
	if cfg.Stream == "" {
		return ErrEmptyStream
	}
	if cfg.Consumer == "" {
		return ErrEmptyConsumer
	}

	subCfg := messaging.SubscriberConfig{
		ID:    cfg.Consumer,
		Topic: cfg.Stream,
		Handler: &eventHandler{
			handler: cfg.Handler,
			ctx:     ctx,
		},
		DeliveryPolicy: messaging.DeliverNewPolicy,
	}

	return es.pubsub.Subscribe(ctx, subCfg)
}

func (es *subEventStore) Close() error {
	return es.pubsub.Close()
}

type event struct {
	Data map[string]interface{}
}

func (re event) Encode() (map[string]interface{}, error) {
	return re.Data, nil
}

type eventHandler struct {
	handler events.EventHandler
	ctx     context.Context
}

func (eh *eventHandler) Handle(msg *messaging.Message) error {
	event := event{
		Data: make(map[string]interface{}),
	}

	if err := json.Unmarshal(msg.GetPayload(), &event.Data); err != nil {
		return err
	}

	if err := eh.handler.Handle(eh.ctx, event); err != nil {
		return fmt.Errorf("failed to handle rabbitmq event: %s", err)
	}

	return nil
}

func (eh *eventHandler) Cancel() error {
	return nil
}

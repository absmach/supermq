// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build rabbitmq && test
// +build rabbitmq,test

package rabbitmq

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/mainflux/mainflux/pkg/events"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	broker "github.com/mainflux/mainflux/pkg/messaging/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type pubEventStore struct {
	conn              *amqp.Connection
	publisher         messaging.Publisher
	unpublishedEvents chan *messaging.Message
	stream            string
	mu                sync.Mutex
}

func NewPublisher(ctx context.Context, url, stream string) (events.Publisher, error) {
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

	publisher, err := broker.NewPublisher(url, brokers.WithPrefix(&eventsPrefix), brokers.WithChannel(ch))
	if err != nil {
		return nil, err
	}

	es := &pubEventStore{
		conn:              conn,
		publisher:         publisher,
		unpublishedEvents: make(chan *messaging.Message, events.MaxUnpublishedEvents),
		stream:            stream,
	}

	go es.StartPublishingRoutine(ctx)

	return es, nil
}

func (es *pubEventStore) Publish(ctx context.Context, event events.Event) error {
	values, err := event.Encode()
	if err != nil {
		return err
	}
	values["occurred_at"] = time.Now().UnixNano()

	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	var record = &messaging.Message{
		Payload: data,
	}

	if ok := es.checkConnection(ctx); !ok {
		es.mu.Lock()
		defer es.mu.Unlock()

		select {
		case es.unpublishedEvents <- record:
		default:
			// If the channel is full (rarely happens), drop the events.
			return nil
		}

		return nil
	}

	return es.publisher.Publish(ctx, es.stream, record)
}

func (es *pubEventStore) StartPublishingRoutine(ctx context.Context) {
	defer close(es.unpublishedEvents)

	ticker := time.NewTicker(events.UnpublishedEventsCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if ok := es.checkConnection(ctx); ok {
				es.mu.Lock()
				for i := len(es.unpublishedEvents) - 1; i >= 0; i-- {
					record := <-es.unpublishedEvents
					if err := es.publisher.Publish(ctx, es.stream, record); err != nil {
						es.unpublishedEvents <- record

						break
					}
				}
				es.mu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (es *pubEventStore) Close() error {
	es.conn.Close()

	return es.publisher.Close()
}

func (es *pubEventStore) checkConnection(ctx context.Context) bool {
	// A timeout is used to avoid blocking the main thread
	ctx, cancel := context.WithTimeout(ctx, events.ConnCheckInterval)
	defer cancel()

	return es.conn.IsClosed()
}

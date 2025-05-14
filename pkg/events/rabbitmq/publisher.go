// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/messaging"
	broker "github.com/absmach/supermq/pkg/messaging/rabbitmq"
)

type pubEventStore struct {
	publisher messaging.Publisher
}

func NewPublisher(ctx context.Context, url string) (events.Publisher, error) {
	publisher, err := broker.NewPublisher(url, broker.Prefix(eventsPrefix), broker.Exchange(exchangeName))
	if err != nil {
		return nil, err
	}

	es := &pubEventStore{
		publisher: publisher,
	}

	return es, nil
}

func (es *pubEventStore) Publish(ctx context.Context, stream string, event events.Event) error {
	values, err := event.Encode()
	if err != nil {
		return err
	}
	values["occurred_at"] = time.Now().UnixNano()

	data, err := json.Marshal(values)
	if err != nil {
		return err
	}

	record := &messaging.Message{
		Payload: data,
	}

	return es.publisher.Publish(ctx, stream, record)
}

func (es *pubEventStore) Close() error {
	return es.publisher.Close()
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/absmach/supermq/auditlogs"
	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/events/store"
)

var ErrMissingOccurredAt = errors.New("missing occurred_at")

type handleFunc func(ctx context.Context, event events.Event) error

func (h handleFunc) Handle(ctx context.Context, event events.Event) error {
	return h(ctx, event)
}

func (h handleFunc) Cancel() error {
	return nil
}

// Start method starts consuming messages received from Event store.
func Start(ctx context.Context, consumer string, sub events.Subscriber, service auditlogs.Service) error {
	subCfg := events.SubscriberConfig{
		Consumer: consumer,
		Stream:   store.StreamAllEvents,
		Handler:  Handle(service),
	}

	return sub.Subscribe(ctx, subCfg)
}

func Handle(service auditlogs.Service) handleFunc {
	return func(ctx context.Context, event events.Event) error {
		data, err := event.Encode()
		if err != nil {
			return err
		}

		requestID, ok := data["request_id"].(string)
		if !ok {
			return errors.New("missing request_id")
		}
		delete(data, "request_id")

		domainID, ok := data["domain"].(string)
		if !ok {
			return errors.New("missing domain")
		}
		delete(data, "domain")

		operation, ok := data["operation"].(string)
		if !ok {
			return errors.New("missing operation")
		}
		delete(data, "operation")

		if operation == "" {
			return errors.New("missing operation")
		}

		entityType, state := toState(operation)
		if entityType == auditlogs.MessageEntity {
			return nil
		}

		entityID, ok := data["id"].(string)
		if !ok {
			return errors.New("missing entity_id")
		}

		occurredAt, ok := data["occurred_at"].(float64)
		if !ok {
			return ErrMissingOccurredAt
		}
		delete(data, "occurred_at")

		if occurredAt == 0 {
			return ErrMissingOccurredAt
		}

		metadata, ok := data["metadata"].(map[string]interface{})
		if !ok {
			metadata = make(map[string]interface{})
		}
		delete(data, "metadata")

		if len(data) == 0 {
			return errors.New("missing attributes")
		}

		al := auditlogs.AuditLog{
			RequestID:       requestID,
			DomainID:        domainID,
			OccurredAt:      time.Unix(0, int64(occurredAt)),
			StateAttributes: data,
			CurrentState:    state,
			EntityID:        entityID,
			EntityType:      entityType,
			Metadata:        metadata,
		}

		return service.Save(ctx, al)
	}
}

func toState(operation string) (auditlogs.EntityType, auditlogs.EntityState) {
	var entityType auditlogs.EntityType
	var entityState auditlogs.EntityState

	op := strings.Split(operation, ".")

	switch op[0] {
	case "user":
		entityType = auditlogs.UserEntity
	case "group":
		entityType = auditlogs.GroupEntity
	case "client":
		entityType = auditlogs.ClientEntity
	case "channel":
		entityType = auditlogs.ChannelEntity
	default:
		entityType = auditlogs.MessageEntity
	}

	switch op[1] {
	case "create":
		entityState = auditlogs.CreatedState
	case "remove", "delete":
		entityState = auditlogs.DeletedState
	default:
		entityState = auditlogs.UpdatedState
	}

	return entityType, entityState
}

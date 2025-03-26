// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auditlogs

import (
	"context"
	"time"

	"github.com/absmach/supermq/pkg/authn"
)

type AuditLog struct {
	ID              string      `json:"id"`
	RequestID       string      `json:"request_id"`
	DomainID        string      `json:"domain_id"`
	OccurredAt      time.Time   `json:"occured_at"`
	ActorID         string      `json:"actor_id"`
	CurrentState    EntityState `json:"current_state"`
	PreviousState   EntityState `json:"previous_state"`
	StateAttributes Metadata    `json:"state_attributes"`
	EntityID        string      `json:"entity_id"`
	EntityType      EntityType  `json:"entity_type"`
	Metadata        Metadata    `json:"metadata"`
}

type AuditLogPage struct {
	Page
	Logs []AuditLog `json:"logs"`
}

type Page struct {
	Total      uint64 `json:"total"`
	Offset     uint64 `json:"offset"`
	Limit      uint64 `json:"limit"`
	Order      string `json:"order,omitempty"`
	Dir        string `json:"dir,omitempty"`
	ID         string `json:"id,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	OccuredAt  string `json:"occured_at,omitempty"`
	ActorID    string `json:"actor_id,omitempty"`
	EntityType string `json:"entity_type,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
}

type Metadata map[string]any


//go:generate mockery --name Repository  --output=./mocks --filename repository.go --quiet --note "Copyright (c) Abstract Machines"
type Repository interface {
	// Save saves audit log.
	Save(ctx context.Context, log AuditLog) error

	// RetrieveByID retrieves audit log by its unique ID.
	RetrieveByID(ctx context.Context, id string) (AuditLog, error)

	// RetrieveAll retrieves all audit logs.
	RetrieveAll(ctx context.Context, pm Page) (AuditLogPage, error)
}

//go:generate mockery --name Service  --output=./mocks --filename service.go --quiet --note "Copyright (c) Abstract Machines"
type Service interface {
	// Save saves audit log.
	Save(ctx context.Context, log AuditLog) error

	// RetrieveByID retrieves audit log by its unique ID.
	RetrieveByID(ctx context.Context, session authn.Session, id string) (AuditLog, error)

	// RetrieveAll retrieves all audit logs.
	RetrieveAll(ctx context.Context, session authn.Session, pm Page) (AuditLogPage, error)
}

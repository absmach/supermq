// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package activitylog

import (
	"context"
	"encoding/json"
	"time"
)

// Activity represents an event activity that occurred in the system.
type Activity struct {
	ID         string                 `json:"id,omitempty" db:"id,omitempty"`
	Operation  string                 `json:"operation,omitempty" db:"operation,omitempty"`
	OccurredAt time.Time              `json:"occurred_at,omitempty" db:"occurred_at,omitempty"`
	Payload    map[string]interface{} `json:"payload,omitempty" db:"payload,omitempty"`
}

// ActivitiesPage represents a page of activities.
type ActivitiesPage struct {
	Total      uint64     `json:"total"`
	Offset     uint64     `json:"offset"`
	Limit      uint64     `json:"limit"`
	Activities []Activity `json:"activities"`
}

// Page is used to filter activities.
type Page struct {
	Offset      uint64    `json:"offset" db:"offset"`
	Limit       uint64    `json:"limit" db:"limit"`
	ID          string    `json:"id,omitempty" db:"id,omitempty"`
	EntityType  string    `json:"entity_type,omitempty"`
	Operation   string    `json:"operation,omitempty" db:"operation,omitempty"`
	From        time.Time `json:"from,omitempty" db:"from,omitempty"`
	To          time.Time `json:"to,omitempty" db:"to,omitempty"`
	WithPayload bool      `json:"with_payload,omitempty"`
	Direction   string    `json:"direction,omitempty"`
}

func (page ActivitiesPage) MarshalJSON() ([]byte, error) {
	type Alias ActivitiesPage
	a := struct {
		Alias
	}{
		Alias: Alias(page),
	}

	if a.Activities == nil {
		a.Activities = make([]Activity, 0)
	}

	return json.Marshal(a)
}

// Service provides access to the activity log service.
//
//go:generate mockery --name Service --output=./mocks --filename service.go --quiet --note "Copyright (c) Abstract Machines"
type Service interface {
	// Save saves the activity to the database.
	Save(ctx context.Context, activity Activity) error

	// ReadAll retrieves all activities from the database with the given page.
	ReadAll(ctx context.Context, token string, page Page) (ActivitiesPage, error)
}

// Repository provides access to the activity log database.
//
//go:generate mockery --name Repository --output=./mocks --filename repository.go --quiet --note "Copyright (c) Abstract Machines"
type Repository interface {
	// Save persists the activity to a database.
	Save(ctx context.Context, activity Activity) error

	// RetrieveAll retrieves all activities from the database with the given page.
	RetrieveAll(ctx context.Context, page Page) (ActivitiesPage, error)
}

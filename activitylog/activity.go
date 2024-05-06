// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package activitylog

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/absmach/magistrala/internal/apiutil"
)

type EntityType uint8

const (
	// Empty represents an empty entity type. This is the default value.
	EmptyEntity EntityType = iota
	UserEntity
	GroupEntity
	ThingEntity
	ChannelEntity
)

// String representation of the possible entity type values.
const (
	userEntityType    = "user"
	groupEntityType   = "group"
	thingEntityType   = "thing"
	channelEntityType = "channel"
)

var ErrMissingOccurredAt = errors.New("missing occurred_at")

// String converts entity type to string literal.
func (e EntityType) String() string {
	switch e {
	case UserEntity:
		return userEntityType
	case GroupEntity:
		return groupEntityType
	case ThingEntity:
		return thingEntityType
	case ChannelEntity:
		return channelEntityType
	default:
		return ""
	}
}

// AuthString returns the entity type as a string for authorization.
func (e EntityType) AuthString() string {
	switch e {
	case UserEntity:
		return userEntityType
	case GroupEntity, ChannelEntity:
		return groupEntityType
	case ThingEntity:
		return thingEntityType
	default:
		return ""
	}
}

// ToEntityType converts string value to a valid entity type.
func ToEntityType(entityType string) (EntityType, error) {
	switch entityType {
	case "":
		return EmptyEntity, nil
	case userEntityType:
		return UserEntity, nil
	case groupEntityType:
		return GroupEntity, nil
	case thingEntityType:
		return ThingEntity, nil
	case channelEntityType:
		return ChannelEntity, nil
	default:
		return EmptyEntity, apiutil.ErrInvalidEntityType
	}
}

// Query returns the SQL condition for the entity type.
func (e EntityType) Query() string {
	switch e {
	case UserEntity:
		return "((operation LIKE 'user.%' AND attributes->>'id' = :entity_id) OR (attributes->>'user_id' = :entity_id))"
	case GroupEntity:
		return "((operation LIKE 'group.%' AND attributes->>'id' = :entity_id) OR (attributes->>'group_id' = :entity_id))"
	case ThingEntity:
		return "((operation LIKE 'thing.%' AND attributes->>'id' = :entity_id) OR (attributes->>'thing_id' = :entity_id))"
	case ChannelEntity:
		return "((operation LIKE 'channel.%' AND attributes->>'id' = :entity_id) OR (attributes->>'group_id' = :entity_id))"
	default:
		return ""
	}
}

// Activity represents an event activity that occurred in the system.
type Activity struct {
	Operation  string                 `json:"operation,omitempty" db:"operation,omitempty"`
	OccurredAt time.Time              `json:"occurred_at,omitempty" db:"occurred_at,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty" db:"attributes,omitempty"` // This is extra information about the activity for example thing_id, user_id, group_id etc.
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata,omitempty"`     // This is decoded metadata from the activity.
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
	Offset         uint64     `json:"offset" db:"offset"`
	Limit          uint64     `json:"limit" db:"limit"`
	Operation      string     `json:"operation,omitempty" db:"operation,omitempty"`
	From           time.Time  `json:"from,omitempty" db:"from,omitempty"`
	To             time.Time  `json:"to,omitempty" db:"to,omitempty"`
	WithAttributes bool       `json:"with_attributes,omitempty"`
	WithMetadata   bool       `json:"with_metadata,omitempty"`
	EntityID       string     `json:"entity_id,omitempty" db:"entity_id,omitempty"`
	EntityType     EntityType `json:"entity_type,omitempty" db:"entity_type,omitempty"`
	Direction      string     `json:"direction,omitempty"`
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

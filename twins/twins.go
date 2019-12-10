//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twins

import (
	"context"
	"time"
)

// Metadata stores arbitrary twin data
type Metadata map[string]interface{}

// Attribute stores indivdual attribute data
type Attribute struct {
	ChannelID   string
	Subtopic    string
	PeristState bool
}

// Definition stores entity's attributes
type Definition struct {
	ID         int
	Created    time.Time
	Attributes map[string]Attribute
}

// Twin represents a Mainflux thing digital twin. Each twin is owned by one thing, and
// is assigned with the unique identifier and (temporary) access key.
type Twin struct {
	Owner       string
	ID          string
	Key         string
	Name        string
	ThingID     string
	Created     time.Time
	Updated     time.Time
	Revision    int
	Definitions []Definition
	Metadata    Metadata
}

// PageMetadata contains page metadata that helps navigation.
type PageMetadata struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Name   string
}

// TwinsPage contains page related metadata as well as a list of twins that
// belong to this page.
type TwinsPage struct {
	PageMetadata
	Twins []Twin
}

// TwinRepository specifies a twin persistence API.
type TwinRepository interface {
	// Save persists the twin
	Save(context.Context, Twin) (string, error)

	// Update performs an update to the existing twin. A non-nil error is
	// returned to indicate operation failure.
	Update(context.Context, Twin) error

	// UpdateKey performs an update key to the existing twin. A non-nil error is
	// returned to indicate operation failure.
	UpdateKey(ctx context.Context, owner, id, key string) error

	// RetrieveByID retrieves the twin having the provided identifier.
	RetrieveByID(ctx context.Context, owner, id string) (Twin, error)

	// RetrieveByKey retrieves the twin having the provided key.
	RetrieveByKey(context.Context, string) (string, error)

	// RetrieveAll retrieves the subset of things owned by the specified user.
	RetrieveAll(context.Context, string, uint64, uint64, string, Metadata) (TwinsPage, error)

	// RetrieveByThing retrieves the subset of twins that represent
	// specified thing.
	RetrieveByThing(context.Context, string, uint64, uint64) (TwinsPage, error)

	// Remove removes the twin having the provided identifier.
	Remove(ctx context.Context, owner, id string) error
}

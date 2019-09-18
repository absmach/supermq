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

type Metadata map[string]interface{}

// Twin represents a Mainflux thing digital twin. Each twin is owned by one thing, and
// is assigned with the unique identifier and (temporary) access key.
type Twin struct {
	ID         string
	Key        string
	Owner      string
	Name       string
	thingID    string
	channelID  string
	created    time.Time
	updated    time.Time
	attributes map[string]interface{}
	state      map[string]interface{}
	topic      string
	revision   int
	Metadata   Metadata
}

// SetMetadata contains page metadata that helps navigation.
type SetMetadata struct {
	Total uint64
	Limit uint64
	Name  string
}

// TwinsSet contains page related metadata as well as a list of twins that
// belong to this page.
type TwinsSet struct {
	SetMetadata
	Twins []Twin
}

// TwinRepository specifies a twin persistence API.
type TwinRepository interface {
	// Save persists the twin. Successful operation is indicated by non-nil
	// error response.
	Save(context.Context, Twin) error

	// Update performs an update to the existing twin. A non-nil error is
	// returned to indicate operation failure.
	Update(context.Context, Twin) error

	// UpdateKey performs an update key to the existing twin. A non-nil error is
	// returned to indicate operation failure.
	UpdateKey(ctx context.Context, id, key string) error

	// RetrieveByID retrieves the twin having the provided identifier.
	RetrieveByID(context.Context, string) (Twin, error)

	// RetrieveByKey retrieves the twin having the provided key.
	RetrieveByKey(context.Context, string) (Twin, error)

	// Remove removes the twin having the provided identifier.
	Remove(context.Context, string) error
}

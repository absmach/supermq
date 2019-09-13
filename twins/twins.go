//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twins

import "context"

// Twin represents a Mainflux thing digital twin. Each twin is owned by one thing, and
// it is assigned with the unique identifier and (temporary) access key.
type Twin struct {
	ID       string
	Owner    string
	Name     string
	Key      string
	Metadata map[string]interface{}
}

// TwinRepository specifies a twin persistence API.
type TwinRepository interface {
	// Save persists the twin. Successful operation is indicated by non-nil
	// error response.
	Save(context.Context, Twin) (string, error)

	// Update performs an update to the existing twin. A non-nil error is
	// returned to indicate operation failure.
	Update(context.Context, Twin) error

	// RetrieveByID retrieves the twin having the provided identifier, that is owned
	// by the specified thing.
	RetrieveByID(context.Context, string, string) (Twin, error)

	// RetrieveByKey retrieves the twin having the provided key, that is owned
	// by the specified thing.
	RetrieveByKey(context.Context, string) (string, error)

	// Remove removes the twin having the provided identifier, that is owned
	// by the specified thing.
	Remove(context.Context, string, string) error
}

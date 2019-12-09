package twins

import (
	"context"
	"time"
)

// State stores actual snapshot of entity's values
type State struct {
	TwinID     string
	ID         int64
	Definition int
	Created    time.Time
	Payload    []map[string]interface{}
}

// StateRepository specifies a state persistence API.
type StateRepository interface {
	// Save persists the state
	Save(context.Context, State) error

	// Count returns the number of states related to state
	Count(context.Context, Twin) (int64, error)
}

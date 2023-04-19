package groups

import (
	"time"

	mfclients "github.com/mainflux/mainflux/internal/mainflux"
)

const (
	// MaxLevel represents the maximum group hierarchy level.
	MaxLevel = uint64(5)
	// MinLevel represents the minimum group hierarchy level.
	MinLevel = uint64(0)
)

// Metadata represents arbitrary JSON.
type Metadata map[string]interface{}

// Group represents the group of Clients.
// Indicates a level in tree hierarchy. Root node is level 1.
// Path in a tree consisting of group IDs
// Paths are unique per owner.
type Group struct {
	ID          string           `json:"id"`
	Owner       string           `json:"owner_id"`
	Parent      string           `json:"parent_id,omitempty"`
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	Metadata    Metadata         `json:"metadata,omitempty"`
	Level       int              `json:"level,omitempty"`
	Path        string           `json:"path,omitempty"`
	Children    []*Group         `json:"children,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Status      mfclients.Status `json:"status"`
}

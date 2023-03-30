package groups

import (
	mfclients "github.com/mainflux/mainflux/internal/mainflux/clients"
	mfgroups "github.com/mainflux/mainflux/internal/mainflux/groups"
)

// Page contains page metadata that helps navigation.
type Page struct {
	Total        uint64
	Offset       uint64            `json:"offset,omitempty"`
	Limit        uint64            `json:"limit,omitempty"`
	Name         string            `json:"name,omitempty"`
	Metadata     mfgroups.Metadata `json:"metadata,omitempty"`
	Disconnected bool              // Used for connected or disconnected lists
	OwnerID      string
	Status       mfclients.Status
	Subject      string
	Action       string
}

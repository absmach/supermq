package groups

import (
	mfclients "github.com/mainflux/mainflux/internal/mainflux/clients"
	mfgroups "github.com/mainflux/mainflux/internal/mainflux/groups"
)

// Page contains page metadata that helps navigation.
type Page struct {
	Total    uint64
	Offset   uint64
	Limit    uint64
	Name     string
	OwnerID  string
	Tag      string
	Metadata mfgroups.Metadata
	SharedBy string
	Status   mfclients.Status
	Subject  string
	Action   string
}

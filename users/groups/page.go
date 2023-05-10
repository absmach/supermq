package groups

import (
	"github.com/mainflux/mainflux/pkg/clients"
)

// Page contains page metadata that helps navigation.
type Page struct {
	Total    uint64
	Offset   uint64
	Limit    uint64
	Name     string
	OwnerID  string
	Tag      string
	Metadata clients.Metadata
	SharedBy string
	Status   clients.Status
	Subject  string
	Action   string
}

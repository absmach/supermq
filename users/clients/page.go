package clients

import mfclients "github.com/mainflux/mainflux/pkg/clients"

// Page contains page metadata that helps navigation.
type Page struct {
	Total    uint64
	Offset   uint64
	Limit    uint64
	Name     string
	Identity string
	Owner    string
	Tag      string
	Metadata mfclients.Metadata
	SharedBy string
	Status   mfclients.Status
	Action   string
	Subject  string
}

package clients

import "github.com/mainflux/mainflux/internal/mainflux"

// Page contains page metadata that helps navigation.
type Page struct {
	Total    uint64
	Offset   uint64
	Limit    uint64
	Name     string
	Identity string
	Owner    string
	Tag      string
	Metadata mainflux.Metadata
	SharedBy string
	Status   mainflux.Status
	Action   string
	Subject  string
}

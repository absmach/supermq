package clients

import mfclients "github.com/mainflux/mainflux/internal/mainflux"

// Metadata represents arbitrary JSON.
type Metadata map[string]interface{}

// Page contains page metadata that helps navigation.
type Page struct {
	Total        uint64
	Offset       uint64   `json:"offset,omitempty"`
	Limit        uint64   `json:"limit,omitempty"`
	Name         string   `json:"name,omitempty"`
	Order        string   `json:"order,omitempty"`
	Dir          string   `json:"dir,omitempty"`
	Metadata     Metadata `json:"metadata,omitempty"`
	Disconnected bool     // Used for connected or disconnected lists
	Owner        string
	Tag          string
	SharedBy     string
	Status       mfclients.Status
	Action       string
	Subject      string
	IDs          []string
}

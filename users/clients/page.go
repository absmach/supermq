package clients

// Metadata represents arbitrary JSON.
type Metadata map[string]interface{}

// Page contains page metadata that helps navigation.
type Page struct {
	Total    uint64
	Offset   uint64
	Limit    uint64
	Name     string
	Identity string
	Owner    string
	Tag      string
	Metadata Metadata
	SharedBy string
	Status   Status
	Action   string
	Subject  string
}

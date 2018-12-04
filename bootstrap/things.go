package bootstrap

// Status represents status of the Thing:
// | Status   | What does it mean                                                             |
// |----------+-------------------------------------------------------------------------------|
// | Created  | Thing has been created and saved, but not bootstrapped                        |
// | Inactive | Thing is create and bootstrapped, but isn't able to communicate over Mainflux |
// | Active   | Thing is able to communicate using Mainflux                                   |
type Status int

const (
	// Created Thing is created, but not configured.
	Created Status = iota
	// Inactive Thing is created and configured, but not able to exchange messages using Mainflux.
	Inactive
	// Active Thing is created, configured, and whitelisted.
	Active
)

// Thing represents Mainflux thing.
type Thing struct {
	ID             string
	Owner          string
	MFThing        string
	MFKey          string
	MFChannels     []string
	ExternalID     string
	ExternalConfig string
	Status         Status
}

// ThingRepository specifies a Thing persistence API.
type ThingRepository interface {
	// Save persists the Thing. Successful operation is indicated by non-nil
	// error response.
	Save(Thing) (string, error)

	// RetrieveByID retrieves the Thing having the provided identifier, that is owned
	// by the specified user.
	RetrieveByID(string, string) (Thing, error)

	// RetrieveAll retrieves the subset of things owned by the specified user.
	RetrieveAll(string, uint64, uint64) []Thing

	// RetrieveByExternalID returns Thing for given external ID.
	RetrieveByExternalID(string) (Thing, error)

	// Update performs and update to an existing Thing. A non-nil error is returned
	// to indicate operation failure.
	Update(Thing) error

	// Remove removes the Thing having the provided identifier, that is owned
	// by the specified user.
	Remove(string, string) error

	// ChangeStatus changes of the Thing, that is owned by the specific user.
	ChangeStatus(string, string, Status) error
}

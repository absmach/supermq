package bootstrap

// State represents state of the Thing:
// | State   | What does it mean                                                             |
// |----------+-------------------------------------------------------------------------------|
// | NewThing | Thing sent a bootstrap request without being being preprovisioned             |
// | Created  | Thing has been created and saved, but not bootstrapped                        |
// | Inactive | Thing is create and bootstrapped, but isn't able to communicate over Mainflux |
// | Active   | Thing is able to communicate using Mainflux                                   |
type State int

const (
	// NewThing is the Thing that sent bootstrap request before corresponding thing has been
	// created on the Bootstrap service side. This means that the Thing is created during
	// bootstrapping process and needs to be approved by the operator to switch to Created state.
	NewThing State = iota
	// Created Thing is created, but not configured.
	Created
	// Inactive Thing is created and configured, but not able to exchange messages using Mainflux.
	Inactive
	// Active Thing is created, configured, and whitelisted.
	Active
)

// Thing represents Mainflux thing.
type Thing struct {
	ID          string
	Owner       string
	MFThing     string
	MFKey       string
	MFChannels  []string
	ExternalID  string
	ExternalKey string
	Config      string
	State       State
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
	RetrieveByExternalID(string, string) (Thing, error)

	// Update performs and update to an existing Thing. A non-nil error is returned
	// to indicate operation failure.
	Update(Thing) error

	// Remove removes the Thing having the provided identifier, that is owned
	// by the specified user.
	Remove(string, string) error

	// ChangeState changes of the Thing, that is owned by the specific user.
	ChangeState(string, string, State) error
}

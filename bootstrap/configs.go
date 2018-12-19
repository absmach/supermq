package bootstrap

import (
	"fmt"
	"strconv"
)

// State represents corresponding Mainflux Thing state. The possible Config States
// as well as description of what that State represents are given in the table:
// | State    | What it means 		                                                           |
// |----------+--------------------------------------------------------------------------------|
// | NewThing | Thing sent a bootstrap request without being preprovisioned 	               |
// | Created  | Thing has been created and saved, but not bootstrapped                         |
// | Inactive | Thing is created and bootstrapped, but isn't able to communicate over Mainflux |
// | Active   | Thing is able to communicate using Mainflux                                    |
type State int

// String returns string representation of State.
func (s State) String() string {
	return strconv.Itoa(int(s))
}

// Value represents human-readable value of the State.
func (s State) Value() string {
	switch s {
	case 0:
		return "New"
	case 1:
		return "Created"
	case 2:
		return "Inactive"
	case 3:
		return "Active"
	default:
		return fmt.Sprintf("Unknown: %d", s)
	}
}

const (
	// NewThing is the Thing that sent bootstrap request before corresponding thing has been
	// created on the Bootstrap service side. This means that the Config is created during
	// bootstrapping process and needs to be approved by the operator to switch to Created state.
	NewThing State = iota
	// Created Thing is created, but not configured.
	Created
	// Inactive Thing is created and configured, but not able to exchange messages using Mainflux.
	Inactive
	// Active Thing is created, configured, and whitelisted.
	Active
)

// Config represents Configuration entity. It wraps information about external entity
// as well as info about corresponding Mainflux entities.
type Config struct {
	ID          string
	Owner       string
	MFThing     string
	MFKey       string
	MFChannels  []string
	ExternalID  string
	ExternalKey string
	Content     string
	State       State
}

// Filter is used for the search filters.
type Filter map[string]string

// ConfigRepository specifies a Config persistence API.
type ConfigRepository interface {
	// Save persists the Config. Successful operation is indicated by non-nil
	// error response.
	Save(Config) (string, error)

	// RetrieveByID retrieves the Config having the provided identifier, that is owned
	// by the specified user.
	RetrieveByID(string, string) (Config, error)

	// RetrieveAll retrieves the subset of Configs with given parameters.
	RetrieveAll(Filter, uint64, uint64) []Config

	// RetrieveByExternalID returns Config for given external ID.
	RetrieveByExternalID(string, string) (Config, error)

	// Update performs and update to an existing Config. A non-nil error is returned
	// to indicate operation failure.
	Update(Config) error

	// Assign updates thing with state NewThing, changing state to Created and chaning
	// Config ownership to the corresponding owner.
	Assign(Config) error

	// Remove removes the Config having the provided identifier, that is owned
	// by the specified user.
	Remove(string, string) error

	// ChangeState changes of the Config, that is owned by the specific user.
	ChangeState(string, string, State) error
}

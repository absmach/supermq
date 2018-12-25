//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package bootstrap

import (
	"strconv"
)

// State represents corresponding Mainflux Thing state. The possible Config States
// as well as description of what that State represents are given in the table:
// | State    | What it means 		                                                           |
// |----------+--------------------------------------------------------------------------------|
// | Created  | Thing has been created and saved, but not bootstrapped                         |
// | Inactive | Thing is created and bootstrapped, but isn't able to communicate over Mainflux |
// | Active   | Thing is able to communicate using Mainflux                                    |
type State int

// String returns string representation of State.
func (s State) String() string {
	return strconv.Itoa(int(s))
}

const (
	// Created Thing is created, but not configured.
	Created State = iota
	// Inactive Thing is created and configured, but not able to exchange messages using Mainflux.
	Inactive
	// Active Thing is created, configured, and whitelisted.
	Active
)

// Config represents Configuration entity. It wraps information about external entity
// as well as info about corresponding Mainflux entities.
type Config struct {
	MFThing     string
	Owner       string
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

	// RetrieveAll retrieves the subset of Configs that are owned by the specific user,
	// with given filter parameters.
	RetrieveAll(string, Filter, uint64, uint64) []Config

	// RetrieveByExternalID returns Config for given external ID.
	RetrieveByExternalID(string, string) (Config, error)

	// Update performs and update to an existing Config. A non-nil error is returned
	// to indicate operation failure.
	Update(Config) error

	// Remove removes the Config having the provided identifier, that is owned
	// by the specified user.
	Remove(string, string) error

	// ChangeState changes of the Config, that is owned by the specific user.
	ChangeState(string, string, State) error

	// SaveUnknown saves Thing which unsuccessfully bootstrapped.
	SaveUnknown(string, string) error

	// RetrieveUnknown returns list of unsuccessfully bootstrapped Things.
	RetrieveUnknown(uint64, uint64) []Config

	// RemoveUnknown removes unsuccessfully bootstrapped Thing. This is done once the
	// corresponding Config is added to the list of existing configs (Save method).
	RemoveUnknown(string, string) error
}

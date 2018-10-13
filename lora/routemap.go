//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package lora

// RouteMapRepository store route map between Lora App Server and Mainflux
type RouteMapRepository interface {
	// Save stores/routes pair lora application topic & mainflux channel.
	Save(string, uint64) error

	// Channel returns mainflux channel for given lora application.
	Channel(string) (uint64, error)

	// Removes mapping from cache.
	Remove(uint64) error
}

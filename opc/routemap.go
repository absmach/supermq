//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package opc

// RouteMapRepository store route map between OPC-UA App Server and Mainflux
type RouteMapRepository interface {
	// Save stores/routes pair opc application topic & mainflux channel.
	Save(string, string) error

	// Channel returns mainflux channel for given opc application.
	Get(string) (string, error)

	// Removes mapping from cache.
	Remove(string) error
}

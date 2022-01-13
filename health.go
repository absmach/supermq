// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mainflux

import (
	"encoding/json"
	"net/http"
)

const (
	version         = "0.12.1"
	contentType     = "Content-Type"
	contentTypeJSON = "application/json"
	svcStatus       = "pass"
	description     = " service"
)

// HealthInfo contains version endpoint response.
type HealthInfo struct {
	// Status contains service status.
	Status string `json:"status"`

	// Version contains current service version.
	Version string `json:"version"`

	// Description contains service description.
	Description string `json:"description"`
}

// Health exposes an HTTP handler for retrieving service health.
func Health(service string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(contentType, contentTypeJSON)
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		res := HealthInfo{
			Status:      svcStatus,
			Description: service + description,
			Version:     version,
		}

		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

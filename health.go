// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mainflux

import (
	"encoding/json"
	"net/http"
)

const (
	contentType     = "Content-Type"
	contentTypeJSON = "application/json"
	svcStatus       = "pass"
	description     = " service"
)

var (
	// Version represents the service version.
	Version = "0.12.1"
	// BuildTime represents the service build time.
	BuildTime = "1970-01-01_00:00:00"
)

// HealthInfo contains version endpoint response.
type HealthInfo struct {
	// Status contains service status.
	Status string `json:"status"`

	// Version contains current service version.
	Version string `json:"version"`

	// Description contains service description.
	Description string `json:"description"`

	// BuildTime contains service build time.
	BuildTime string `json:"build_time"`
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
			Version:     Version,
			BuildTime:   BuildTime,
		}

		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(res); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// +build !test

package postgres

import (
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler() http.Handler {
	r := bone.New()
	r.GetFunc("/version", mainflux.Version("postgres-writer"))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

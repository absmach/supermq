// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0
package http

import (
	"log/slog"
	"net/http"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/auth/api/http/keys"
	"github.com/absmach/supermq/auth/api/http/pats"
	mgauthn "github.com/absmach/supermq/pkg/authn"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc auth.Service, authn mgauthn.Authentication, logger *slog.Logger, instanceID string) http.Handler {
	mux := chi.NewRouter()

	mux = keys.MakeHandler(svc, mux, logger)
	mux = pats.MakeHandler(svc, mux, authn, logger)

	mux.Get("/health", supermq.Health("auth", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

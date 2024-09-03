// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0
package http

import (
	"log/slog"
	"net/http"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/auth/api/http/keys"
	domainsHTTP "github.com/absmach/magistrala/internal/domains/api/http"
	"github.com/absmach/magistrala/pkg/domains"
	"github.com/absmach/magistrala/pkg/roles"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc auth.Service, dsvc domains.Service, dRoles roles.Roles, logger *slog.Logger, instanceID string) http.Handler {
	mux := chi.NewRouter()

	mux = keys.MakeHandler(svc, mux, logger)
	mux = domainsHTTP.MakeHandler(dsvc, dRoles, mux, logger)

	mux.Get("/health", magistrala.Health("auth", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

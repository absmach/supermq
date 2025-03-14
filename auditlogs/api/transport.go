// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"log/slog"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auditlogs"
	api "github.com/absmach/supermq/api/http"
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	kithttp "github.com/go-kit/kit/transport/http"
)

// MakeHandler returns a HTTP handler for Channels API endpoints.
func MakeHandler(svc auditlogs.Service, authn smqauthn.Authentication, mux *chi.Mux, logger *slog.Logger, instanceID string, idp supermq.IDProvider) *chi.Mux {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	mux.Route("/auditlogs", func(r chi.Router) {
		r.Use(api.AuthenticateMiddleware(authn, false))
		r.Use(api.RequestIDMiddleware(idp))

		r.Get("/{id}", otelhttp.NewHandler(kithttp.NewServer(
			retrieveAuditLogEndpoint(svc),
			decodeRetrieveAuditLogReq,
			api.EncodeResponse,
			opts...,
		), "retrieve_audit_log").ServeHTTP)

		r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
			retrieveAllAuditLogsEndpoint(svc),
			decodeRetrieveAllAuditLogsReq,
			api.EncodeResponse,
			opts...,
		), "retrieve_all_audit_logs").ServeHTTP)
	})

	mux.Get("health", supermq.Health("auditlogs", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"log/slog"
	"net/http"

	"github.com/absmach/supermq"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MakeHandler returns a HTTP handler for the notifications service.
func MakeHandler(serviceName, instanceID string) http.Handler {
	mux := chi.NewRouter()

	mux.Get("/health", supermq.Health(serviceName, instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

// LoggingErrorEncoder is an HTTP error encoder that logs the error.
func LoggingErrorEncoder(logger *slog.Logger, w http.ResponseWriter, err error) {
	logger.Error("error encoding response", slog.Any("error", err))
}

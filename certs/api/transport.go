// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/absmach/supermq"
	api "github.com/absmach/supermq/api/http"
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/certs"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	contentType = "application/json"
	offsetKey   = "offset"
	limitKey    = "limit"
	revokedKey  = "revoked"
	defOffset   = 0
	defLimit    = 10
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc certs.Service, authn smqauthn.AuthNMiddleware, logger *slog.Logger, instanceID string, idp supermq.IDProvider) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(authn.Middleware())
		r.Use(api.RequestIDMiddleware(idp))

		r.Route("/{domainID}", func(r chi.Router) {
			r.Route("/certs", func(r chi.Router) {
				r.Post("/", otelhttp.NewHandler(kithttp.NewServer(
					issueCert(svc),
					decodeCerts,
					api.EncodeResponse,
					opts...,
				), "issue").ServeHTTP)
				r.Get("/{certID}", otelhttp.NewHandler(kithttp.NewServer(
					viewCert(svc),
					decodeViewCert,
					api.EncodeResponse,
					opts...,
				), "view").ServeHTTP)
				r.Post("/{clientID}/revoke-all", otelhttp.NewHandler(kithttp.NewServer(
					revokeAllCerts(svc),
					decodeRevokeAllCerts,
					api.EncodeResponse,
					opts...,
				), "revoke").ServeHTTP)
				r.Post("/{certID}/revoke", otelhttp.NewHandler(kithttp.NewServer(
					revokeBySerial(svc),
					decodeRevokeBySerial,
					api.EncodeResponse,
					opts...,
				), "revoke_by_serial").ServeHTTP)
			})
			r.Get("/serials/{clientID}", otelhttp.NewHandler(kithttp.NewServer(
				listSerials(svc),
				decodeListCerts,
				api.EncodeResponse,
				opts...,
			), "list_serials").ServeHTTP)
		})
	})
	r.Handle("/metrics", promhttp.Handler())
	r.Get("/health", supermq.Health("certs", instanceID))

	return r
}

func decodeListCerts(_ context.Context, r *http.Request) (any, error) {
	l, err := apiutil.ReadNumQuery[uint64](r, limitKey, defLimit)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	o, err := apiutil.ReadNumQuery[uint64](r, offsetKey, defOffset)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	revoked, err := apiutil.ReadStringQuery(r, revokedKey, "all")
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	req := listReq{
		clientID: chi.URLParam(r, "clientID"),
		pm: certs.PageMetadata{
			Offset:  o,
			Limit:   l,
			Revoked: revoked,
		},
	}
	return req, nil
}

func decodeViewCert(_ context.Context, r *http.Request) (any, error) {
	req := viewReq{
		serialID: chi.URLParam(r, "certID"),
	}

	return req, nil
}

func decodeCerts(_ context.Context, r *http.Request) (any, error) {
	if r.Header.Get("Content-Type") != contentType {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := addCertsReq{
		token:    apiutil.ExtractBearerToken(r),
		domainID: chi.URLParam(r, "domainID"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	return req, nil
}

func decodeRevokeAllCerts(_ context.Context, r *http.Request) (any, error) {
	req := revokeAllReq{
		token:    apiutil.ExtractBearerToken(r),
		clientID: chi.URLParam(r, "clientID"),
		domainID: chi.URLParam(r, "domainID"),
	}

	return req, nil
}

func decodeRevokeBySerial(_ context.Context, r *http.Request) (any, error) {
	req := revokeBySerialReq{
		serialID: chi.URLParam(r, "certID"),
	}

	return req, nil
}

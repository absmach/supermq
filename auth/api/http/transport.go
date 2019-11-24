// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const contentType = "application/json"

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc auth.Service, tracer opentracing.Tracer) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()

	mux.Post("/keys", kithttp.NewServer(
		kitot.TraceServer(tracer, "issue")(issueEndpoint(svc)),
		decodeIssue,
		encodeResponse,
		opts...,
	))

	mux.Delete("/keys/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "revoke")(revokeEndpoint(svc)),
		decodeKeyReq,
		encodeResponse,
		opts...,
	))

	mux.Get("/keys/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "retrieve")(retrieveEndpoint(svc)),
		decodeKeyReq,
		encodeResponse,
		opts...,
	))

	mux.Post("/keys/identify", kithttp.NewServer(
		kitot.TraceServer(tracer, "identify")(identifyEndpoint(svc)),
		decodeKeyReq,
		encodeResponse,
		opts...,
	))

	mux.GetFunc("/version", mainflux.Version("auth"))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeIssue(_ context.Context, r *http.Request) (interface{}, error) {
	req := issueKeyReq{
		issuer: r.Header.Get("Authorization"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

func decodeKeyReq(_ context.Context, r *http.Request) (interface{}, error) {
	req := keyReq{
		issuer: r.Header.Get("Authorization"),
		id:     bone.GetValue(r, "id"),
	}
	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)

	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}

		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", contentType)

	switch err {
	case auth.ErrMalformedEntity:
		w.WriteHeader(http.StatusBadRequest)
	case auth.ErrUnauthorizedAccess:
		w.WriteHeader(http.StatusForbidden)
	case auth.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case auth.ErrConflict:
		w.WriteHeader(http.StatusConflict)
	case io.EOF, io.ErrUnexpectedEOF:
		w.WriteHeader(http.StatusBadRequest)
	default:
		switch err.(type) {
		case *json.SyntaxError:
			w.WriteHeader(http.StatusBadRequest)
		case *json.UnmarshalTypeError:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

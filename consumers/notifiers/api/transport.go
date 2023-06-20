// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	notifiers "github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	contentType = "application/json"
	offsetKey   = "offset"
	limitKey    = "limit"
	topicKey    = "topic"
	contactKey  = "contact"
	defOffset   = 0
	defLimit    = 20
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc notifiers.Service, logger logger.Logger, instanceID string) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, encodeError)),
	}

	mux := bone.New()

	mux.Post("/subscriptions", otelhttp.NewHandler(kithttp.NewServer(
		createSubscriptionEndpoint(svc),
		decodeCreate,
		encodeResponse,
		opts...,
	), "create"))

	mux.Get("/subscriptions/:subID", otelhttp.NewHandler(kithttp.NewServer(
		viewSubscriptionEndpint(svc),
		decodeSubscription,
		encodeResponse,
		opts...,
	), "view"))

	mux.Get("/subscriptions", otelhttp.NewHandler(kithttp.NewServer(
		listSubscriptionsEndpoint(svc),
		decodeList,
		encodeResponse,
		opts...,
	), "list"))

	mux.Delete("/subscriptions/:subID", otelhttp.NewHandler(kithttp.NewServer(
		deleteSubscriptionEndpint(svc),
		decodeSubscription,
		encodeResponse,
		opts...,
	), "delete"))

	mux.GetFunc("/health", mainflux.Health("notifier", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeCreate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errors.Wrap(errors.ErrUnsupportedContentType, apiutil.ErrValidation)
	}

	req := createSubReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrMalformedEntity, errors.Wrap(err, apiutil.ErrValidation))
	}

	return req, nil
}

func decodeSubscription(_ context.Context, r *http.Request) (interface{}, error) {
	req := subReq{
		id:    bone.GetValue(r, "subID"),
		token: apiutil.ExtractBearerToken(r),
	}

	return req, nil
}

func decodeList(_ context.Context, r *http.Request) (interface{}, error) {
	req := listSubsReq{token: apiutil.ExtractBearerToken(r)}
	vals := bone.GetQuery(r, topicKey)
	if len(vals) > 0 {
		req.topic = vals[0]
	}

	vals = bone.GetQuery(r, contactKey)
	if len(vals) > 0 {
		req.contact = vals[0]
	}

	offset, err := apiutil.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return listSubsReq{}, errors.Wrap(err, apiutil.ErrValidation)
	}
	req.offset = uint(offset)

	limit, err := apiutil.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return listSubsReq{}, errors.Wrap(err, apiutil.ErrValidation)
	}
	req.limit = uint(limit)

	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch {
	case errors.Contains(err, apiutil.ErrMalformedEntity),
		errors.Contains(err, apiutil.ErrInvalidContact),
		errors.Contains(err, apiutil.ErrInvalidTopic),
		errors.Contains(err, apiutil.ErrMissingID),
		errors.Contains(err, apiutil.ErrInvalidQueryParams):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, errors.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Contains(err, errors.ErrAuthentication),
		errors.Contains(err, apiutil.ErrBearerToken):
		w.WriteHeader(http.StatusUnauthorized)
	case errors.Contains(err, errors.ErrConflict):
		w.WriteHeader(http.StatusConflict)
	case errors.Contains(err, errors.ErrUnsupportedContentType):
		w.WriteHeader(http.StatusUnsupportedMediaType)

	case errors.Contains(err, errors.ErrCreateEntity),
		errors.Contains(err, errors.ErrViewEntity),
		errors.Contains(err, errors.ErrRemoveEntity):
		w.WriteHeader(http.StatusInternalServerError)

	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	if errorVal, ok := err.(errors.Error); ok {
		w.Header().Set("Content-Type", contentType)

		errMsg := errorVal.Msg()
		if errorVal.Err() != nil {
			errMsg = fmt.Sprintf("%s : %s", errorVal.Msg(), errorVal.Err().Msg())
		}

		if err := json.NewEncoder(w).Encode(apiutil.ErrorRes{Err: errMsg}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

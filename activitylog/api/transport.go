// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	operationKey  = "operation"
	fromKey       = "from"
	toKey         = "to"
	attributesKey = "with_attributes"
	metadataKey   = "with_metadata"
	entityIDKey   = "id"
	entityTypeKey = "entity_type"
)

// MakeHandler returns a HTTP API handler with health check and metrics.
func MakeHandler(svc activitylog.Service, logger *slog.Logger, svcName, instanceID string) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	mux := chi.NewRouter()

	mux.Get("/activities", otelhttp.NewHandler(kithttp.NewServer(
		listActivitiesEndpoint(svc),
		decodeListActivitiesReq,
		api.EncodeResponse,
		opts...,
	), "list_activities").ServeHTTP)

	mux.Get("/health", magistrala.Health(svcName, instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeListActivitiesReq(_ context.Context, r *http.Request) (interface{}, error) {
	offset, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	limit, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	operation, err := apiutil.ReadStringQuery(r, operationKey, "")
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	from, err := apiutil.ReadNumQuery[int64](r, fromKey, 0)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	if from > math.MaxInt32 {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidTimeFormat)
	}
	var fromTime time.Time
	if from != 0 {
		fromTime = time.Unix(from, 0)
	}
	to, err := apiutil.ReadNumQuery[int64](r, toKey, 0)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	if to > math.MaxInt32 {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidTimeFormat)
	}
	var toTime time.Time
	if to != 0 {
		toTime = time.Unix(to, 0)
	}
	attributes, err := apiutil.ReadBoolQuery(r, attributesKey, false)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	metadata, err := apiutil.ReadBoolQuery(r, metadataKey, false)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	dir, err := apiutil.ReadStringQuery(r, api.DirKey, api.DescDir)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	entityID, err := apiutil.ReadStringQuery(r, entityIDKey, "")
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	entity, err := apiutil.ReadStringQuery(r, entityTypeKey, "")
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	entityType, err := activitylog.ToEntityType(entity)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	req := listActivitiesReq{
		token: apiutil.ExtractBearerToken(r),
		page: activitylog.Page{
			Offset:         offset,
			Limit:          limit,
			Operation:      operation,
			From:           fromTime,
			To:             toTime,
			WithAttributes: attributes,
			WithMetadata:   metadata,
			EntityID:       entityID,
			EntityType:     entityType,
			Direction:      dir,
		},
	}

	return req, nil
}

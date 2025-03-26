// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"net/http"

	api "github.com/absmach/supermq/api/http"
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/auditlogs"
	"github.com/go-chi/chi/v5"
)

const (
	requestIDKey  = "request_id"
	actorIDKey    = "actor_id"
	entityTypeKey = "entity_type"
	entityIDKey   = "entity_id"
)

func decodeRetrieveAuditLogReq(_ context.Context, r *http.Request) (interface{}, error) {
	req := retriveAuditLogReq{
		id: chi.URLParam(r, "auditLogID"),
	}

	return req, nil
}

func decodeRetrieveAllAuditLogsReq(_ context.Context, r *http.Request) (interface{}, error) {
	order, err := apiutil.ReadStringQuery(r, api.OrderKey, "")
	if err != nil {
		return retrieveAllAuditLogsReq{}, err
	}

	dir, err := apiutil.ReadStringQuery(r, api.DirKey, "")
	if err != nil {
		return retrieveAllAuditLogsReq{}, err
	}

	requestID, err := apiutil.ReadStringQuery(r, requestIDKey, "")
	if err != nil {
		return retrieveAllAuditLogsReq{}, err
	}

	actorID, err := apiutil.ReadStringQuery(r, actorIDKey, "")
	if err != nil {
		return retrieveAllAuditLogsReq{}, err
	}

	entityType, err := apiutil.ReadStringQuery(r, entityTypeKey, "")
	if err != nil {
		return retrieveAllAuditLogsReq{}, err
	}

	entityID, err := apiutil.ReadStringQuery(r, entityIDKey, "")
	if err != nil {
		return retrieveAllAuditLogsReq{}, err
	}

	req := retrieveAllAuditLogsReq{
		Page: auditlogs.Page{
			Order:      order,
			Dir:        dir,
			RequestID:  requestID,
			ActorID:    actorID,
			EntityType: entityType,
			EntityID:   entityID,
		},
	}

	return req, nil

}

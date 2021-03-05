// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	intapihttp "github.com/mainflux/mainflux/internal/api/http"
	internalerr "github.com/mainflux/mainflux/internal/errors"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	contentType = "application/json"

	offsetKey   = "offset"
	limitKey    = "limit"
	nameKey     = "name"
	emailKey    = "email"
	metadataKey = "metadata"

	defOffset = 0
	defLimit  = 10
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc users.Service, tracer opentracing.Tracer) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()

	mux.Post("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "register")(registrationEndpoint(svc)),
		decodeCredentials,
		encodeResponse,
		opts...,
	))

	mux.Get("/users/profile", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_profile")(viewProfileEndpoint(svc)),
		decodeViewProfile,
		encodeResponse,
		opts...,
	))

	mux.Get("/users/:userID", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_user")(viewUserEndpoint(svc)),
		decodeViewUser,
		encodeResponse,
		opts...,
	))

	mux.Get("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_users")(listUsersEndpoint(svc)),
		decodeListUsers,
		encodeResponse,
		opts...,
	))

	mux.Put("/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_user")(updateUserEndpoint(svc)),
		decodeUpdateUser,
		encodeResponse,
		opts...,
	))

	mux.Get("/users/:userID/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_memberships")(listMembershipsEndpoint(svc)),
		decodeListUserGroupsRequest,
		encodeResponse,
		opts...,
	))

	mux.Post("/password/reset-request", kithttp.NewServer(
		kitot.TraceServer(tracer, "res-req")(passwordResetRequestEndpoint(svc)),
		decodePasswordResetRequest,
		encodeResponse,
		opts...,
	))

	mux.Put("/password/reset", kithttp.NewServer(
		kitot.TraceServer(tracer, "reset")(passwordResetEndpoint(svc)),
		decodePasswordReset,
		encodeResponse,
		opts...,
	))

	mux.Patch("/password", kithttp.NewServer(
		kitot.TraceServer(tracer, "reset")(passwordChangeEndpoint(svc)),
		decodePasswordChange,
		encodeResponse,
		opts...,
	))

	mux.Post("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "add_group")(createGroupEndpoint(svc)),
		decodeGroupCreate,
		encodeResponse,
		opts...,
	))

	mux.Get("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_groups")(listGroupsEndpoint(svc)),
		decodeListUserGroupsRequest,
		encodeResponse,
		opts...,
	))

	mux.Delete("/groups/:groupID", kithttp.NewServer(
		kitot.TraceServer(tracer, "delete_group")(deleteGroupEndpoint(svc)),
		decodeGroupRequest,
		encodeResponse,
		opts...,
	))

	mux.Put("/groups/:groupID/users/:userID", kithttp.NewServer(
		kitot.TraceServer(tracer, "assign_user_to_group")(assignUserToGroup(svc)),
		decodeUserGroupRequest,
		encodeResponse,
		opts...,
	))

	mux.Delete("/groups/:groupID/users/:userID", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_user_from_group")(removeUserFromGroup(svc)),
		decodeUserGroupRequest,
		encodeResponse,
		opts...,
	))

	mux.Get("/groups/:groupID/users", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_members")(listMembersEndpoint(svc)),
		decodeListUserGroupsRequest,
		encodeResponse,
		opts...,
	))

	mux.Patch("/groups/:groupID", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_group")(updateGroupEndpoint(svc)),
		decodeGroupUpdate,
		encodeResponse,
		opts...,
	))

	mux.Get("/groups/:groupID/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_children_groups")(listGroupsEndpoint(svc)),
		decodeListUserGroupsRequest,
		encodeResponse,
		opts...,
	))

	mux.Get("/groups/:groupID", kithttp.NewServer(
		kitot.TraceServer(tracer, "group")(viewGroupEndpoint(svc)),
		decodeGroupRequest,
		encodeResponse,
		opts...,
	))

	mux.Post("/tokens", kithttp.NewServer(
		kitot.TraceServer(tracer, "login")(loginEndpoint(svc)),
		decodeCredentials,
		encodeResponse,
		opts...,
	))

	mux.GetFunc("/version", mainflux.Version("users"))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeViewUser(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewUserReq{
		token:  r.Header.Get("Authorization"),
		userID: bone.GetValue(r, "userID"),
	}
	return req, nil
}

func decodeViewProfile(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewUserReq{
		token: r.Header.Get("Authorization"),
	}
	return req, nil
}

func decodeListUsers(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := intapihttp.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := intapihttp.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	e, err := intapihttp.ReadStringQuery(r, emailKey)
	if err != nil {
		return nil, err
	}

	m, err := intapihttp.ReadMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	req := listUsersReq{
		token:    r.Header.Get("Authorization"),
		offset:   o,
		limit:    l,
		email:    e,
		metadata: m,
	}
	return req, nil
}

func decodeUpdateUser(_ context.Context, r *http.Request) (interface{}, error) {
	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(users.ErrMalformedEntity, err)
	}

	req.token = r.Header.Get("Authorization")
	return req, nil
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, internalerr.ErrUnsupportedContentType
	}

	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return nil, errors.Wrap(internalerr.ErrMalformedEntity, err)
	}

	return userReq{user}, nil
}

func decodePasswordResetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, internalerr.ErrUnsupportedContentType
	}

	var req passwResetReq

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(internalerr.ErrMalformedEntity, err)
	}

	req.Host = r.Header.Get("Referer")
	return req, nil
}

func decodePasswordReset(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, internalerr.ErrUnsupportedContentType
	}

	var req resetTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(internalerr.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodePasswordChange(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, internalerr.ErrUnsupportedContentType
	}

	var req passwChangeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(internalerr.ErrMalformedEntity, err)
	}

	req.Token = r.Header.Get("Authorization")

	return req, nil
}

// Group related methods
func decodeGroupCreate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, internalerr.ErrUnsupportedContentType
	}

	var req createGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(internalerr.ErrMalformedEntity, err)
	}

	req.token = r.Header.Get("Authorization")

	return req, nil
}

func decodeGroupUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, internalerr.ErrUnsupportedContentType
	}

	req := updateGroupReq{
		token: r.Header.Get("Authorization"),
		id:    bone.GetValue(r, "groupID"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(internalerr.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := groupReq{
		token:   r.Header.Get("Authorization"),
		groupID: bone.GetValue(r, "groupID"),
		name:    bone.GetValue(r, "name"),
	}

	return req, nil
}

func decodeListUserGroupsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := intapihttp.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := intapihttp.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	n, err := intapihttp.ReadStringQuery(r, nameKey)
	if err != nil {
		return nil, err
	}

	m, err := intapihttp.ReadMetadataQuery(r, metadataKey)
	if err != nil {
		return nil, err
	}

	groupID := bone.GetValue(r, "groupID")
	userID := bone.GetValue(r, "userID")

	req := listUserGroupReq{
		token:    r.Header.Get("Authorization"),
		groupID:  groupID,
		userID:   userID,
		offset:   o,
		limit:    l,
		name:     n,
		metadata: m,
	}
	return req, nil
}

func decodeUserGroupRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := userGroupReq{
		token:   r.Header.Get("Authorization"),
		groupID: bone.GetValue(r, "groupID"),
		userID:  bone.GetValue(r, "userID"),
	}
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
	switch errorVal := err.(type) {
	case errors.Error:
		w.Header().Set("Content-Type", contentType)
		switch {
		case errors.Contains(errorVal, internalerr.ErrInvalidQueryParams):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrMalformedEntity):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrUnauthorizedAccess):
			w.WriteHeader(http.StatusForbidden)
		case errors.Contains(errorVal, users.ErrConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.Contains(errorVal, users.ErrGroupConflict):
			w.WriteHeader(http.StatusConflict)
		case errors.Contains(errorVal, internalerr.ErrUnsupportedContentType):
			w.WriteHeader(http.StatusUnsupportedMediaType)
		case errors.Contains(errorVal, internalerr.ErrMalformedEntity):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.ErrUnexpectedEOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, io.EOF):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrUserNotFound):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Contains(errorVal, users.ErrRecoveryToken):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		if errorVal.Msg() != "" {
			if err := json.NewEncoder(w).Encode(errorRes{Err: errorVal.Msg()}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

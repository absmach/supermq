// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/go-chi/chi/v5"
)

func DecodeCreateRole(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	req := createRoleReq{
		token: apiutil.ExtractBearerToken(r),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	return req, nil
}

func DecodeListRoles(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	req := listRolesReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		limit:    l,
		offset:   o,
	}
	return req, nil
}

func DecodeViewRole(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewRoleReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	return req, nil
}

func DecodeUpdateRole(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	req := updateRoleReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	return req, nil
}

func DecodeDeleteRole(_ context.Context, r *http.Request) (interface{}, error) {
	req := deleteRoleReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	return req, nil
}

func DecodeAddRoleOperations(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	req := addRoleOperationsReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	return req, nil
}

func DecodeListRoleOperations(_ context.Context, r *http.Request) (interface{}, error) {
	req := listRoleOperationsReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	return req, nil
}

func DecodeDeleteRoleOperations(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	req := deleteRoleOperationsReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	return req, nil
}

func DecodeDeleteAllRoleOperations(_ context.Context, r *http.Request) (interface{}, error) {
	req := deleteAllRoleOperationsReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	return req, nil
}

func DecodeAddRoleMembers(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	req := addRoleMembersReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	return req, nil
}

func DecodeListRoleMembers(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	req := listRoleMembersReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
		limit:    l,
		offset:   o,
	}
	return req, nil
}

func DecodeDeleteRoleMembers(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}
	req := deleteRoleMembersReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}
	return req, nil
}

func DecodeDeleteAllRoleMembers(_ context.Context, r *http.Request) (interface{}, error) {
	req := deleteAllRoleMembersReq{
		token:    apiutil.ExtractBearerToken(r),
		entityID: chi.URLParam(r, "entityID"),
		roleName: chi.URLParam(r, "roleName"),
	}
	return req, nil
}

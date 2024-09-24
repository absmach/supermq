package http

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

import (
	"log/slog"
	"net/http"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	entityRoleHttp "github.com/absmach/magistrala/pkg/entityroles/api/http"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// MakeHandler returns a HTTP handler for Groups API endpoints.
func groupsHandler(svc groups.Service, mux *chi.Mux, logger *slog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	mux.Route("/groups", func(r chi.Router) {
		r.Post("/", otelhttp.NewHandler(kithttp.NewServer(
			CreateGroupEndpoint(svc, auth.NewGroupKind),
			DecodeGroupCreate,
			api.EncodeResponse,
			opts...,
		), "create_group").ServeHTTP)

		r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
			ListGroupsEndpoint(svc),
			DecodeListGroupsRequest,
			api.EncodeResponse,
			opts...,
		), "list_groups").ServeHTTP)

		r.Route("/{groupID}", func(r chi.Router) {
			r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
				ViewGroupEndpoint(svc),
				DecodeGroupRequest,
				api.EncodeResponse,
				opts...,
			), "view_group").ServeHTTP)

			r.Put("/", otelhttp.NewHandler(kithttp.NewServer(
				UpdateGroupEndpoint(svc),
				DecodeGroupUpdate,
				api.EncodeResponse,
				opts...,
			), "update_group").ServeHTTP)

			r.Delete("/", otelhttp.NewHandler(kithttp.NewServer(
				DeleteGroupEndpoint(svc),
				DecodeGroupRequest,
				api.EncodeResponse,
				opts...,
			), "delete_group").ServeHTTP)

			r.Post("/enable", otelhttp.NewHandler(kithttp.NewServer(
				EnableGroupEndpoint(svc),
				DecodeChangeGroupStatusRequest,
				api.EncodeResponse,
				opts...,
			), "enable_group").ServeHTTP)

			r.Post("/disable", otelhttp.NewHandler(kithttp.NewServer(
				DisableGroupEndpoint(svc),
				DecodeChangeGroupStatusRequest,
				api.EncodeResponse,
				opts...,
			), "disable_group").ServeHTTP)

			r.Get("/parents", otelhttp.NewHandler(kithttp.NewServer(
				listParentGroupsEndpoint(svc),
				decodeListParentsRequest,
				api.EncodeResponse,
				opts...,
			), "list_parent_groups").ServeHTTP)

			r.Route("/parent", func(r chi.Router) {
				r.Post("/", otelhttp.NewHandler(kithttp.NewServer(
					addParentGroupEndpoint(svc),
					decodeAddParentGroupRequest,
					api.EncodeResponse,
					opts...,
				), "add_parent_group").ServeHTTP)

				r.Delete("/", otelhttp.NewHandler(kithttp.NewServer(
					removeParentGroupEndpoint(svc),
					decodeRemoveParentGroupRequest,
					api.EncodeResponse,
					opts...,
				), "remove_parent_group").ServeHTTP)

				r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
					viewParentGroupEndpoint(svc),
					decodeViewParentGroupRequest,
					api.EncodeResponse,
					opts...,
				), "view_parent_group").ServeHTTP)
			})

			r.Route("/children", func(r chi.Router) {
				r.Post("/", otelhttp.NewHandler(kithttp.NewServer(
					addChildrenGroupsEndpoint(svc),
					decodeAddChildrenGroupsRequest,
					api.EncodeResponse,
					opts...,
				), "add_children_groups").ServeHTTP)

				r.Delete("/", otelhttp.NewHandler(kithttp.NewServer(
					removeChildrenGroupsEndpoint(svc),
					decodeRemoveChildrenGroupsRequest,
					api.EncodeResponse,
					opts...,
				), "remove_children_groups").ServeHTTP)

				r.Delete("/all", otelhttp.NewHandler(kithttp.NewServer(
					removeAllChildrenGroupsEndpoint(svc),
					decodeRemoveAllChildrenGroupsRequest,
					api.EncodeResponse,
					opts...,
				), "remove_all_children_groups").ServeHTTP)

				r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
					listChildrenGroupsEndpoint(svc),
					decodeListChildrenGroupsRequest,
					api.EncodeResponse,
					opts...,
				), "list_children_groups").ServeHTTP)
			})
		})

	})
	mux = entityRoleHttp.RolesHandler(svc, "/groups", mux, logger)

	return mux
}

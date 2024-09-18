// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/absmach/magistrala/internal/api"
	gapi "github.com/absmach/magistrala/internal/groups/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/auth"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/absmach/magistrala/pkg/policies"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func groupsHandler(svc groups.Service, authClient auth.AuthClient, r *chi.Mux, logger *slog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	checkSuperAdminMiddleware := api.CheckSuperAdminMiddleware(authClient)
	r.Group(func(r chi.Router) {
		r.Use(api.IdentifyMiddleware(authClient))

		r.Route("/channels", func(r chi.Router) {
			authzMiddleware := api.AuthorizeMiddleware(authClient, gapi.CreateGroupAuthReq)
			r.Post("/", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(gapi.CreateGroupEndpoint(svc, authClient, policies.NewChannelKind)),
				gapi.DecodeGroupCreate,
				api.EncodeResponse,
				opts...,
			), "create_channel").ServeHTTP)

			authzMiddleware = api.AuthorizeMiddleware(authClient, gapi.ViewGroupAuthReq)
			r.Get("/{groupID}", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(gapi.ViewGroupEndpoint(svc, authClient)),
				gapi.DecodeGroupRequest,
				api.EncodeResponse,
				opts...,
			), "view_channel").ServeHTTP)

			authzMiddleware = api.AuthorizeMiddleware(authClient, gapi.DeleteGroupAuthReq)
			r.Delete("/{groupID}", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(gapi.DeleteGroupEndpoint(svc, authClient)),
				gapi.DecodeGroupRequest,
				api.EncodeResponse,
				opts...,
			), "delete_channel").ServeHTTP)

			r.Get("/{groupID}/permissions", otelhttp.NewHandler(kithttp.NewServer(
				gapi.ViewGroupPermsEndpoint(svc, authClient),
				gapi.DecodeGroupPermsRequest,
				api.EncodeResponse,
				opts...,
			), "view_channel_permissions").ServeHTTP)

			authzMiddleware = api.AuthorizeMiddleware(authClient, gapi.UpdateGroupAuthReq)
			r.Put("/{groupID}", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(gapi.UpdateGroupEndpoint(svc, authClient)),
				gapi.DecodeGroupUpdate,
				api.EncodeResponse,
				opts...,
			), "update_channel").ServeHTTP)

			authzMiddleware = api.AuthorizeMiddleware(authClient, gapi.ListGroupsByUserAuthReq)
			r.Get("/", otelhttp.NewHandler(kithttp.NewServer(
				checkSuperAdminMiddleware(authzMiddleware(gapi.ListGroupsEndpoint(svc, authClient, "channels", "users"))),
				gapi.DecodeListGroupsRequest,
				api.EncodeResponse,
				opts...,
			), "list_channels").ServeHTTP)

			authzMiddleware = api.AuthorizeMiddleware(authClient, gapi.ChangeGroupStatusAuthReq)
			r.Post("/{groupID}/enable", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(gapi.EnableGroupEndpoint(svc, authClient)),
				gapi.DecodeChangeGroupStatus,
				api.EncodeResponse,
				opts...,
			), "enable_channel").ServeHTTP)

			r.Post("/{groupID}/disable", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(gapi.DisableGroupEndpoint(svc, authClient)),
				gapi.DecodeChangeGroupStatus,
				api.EncodeResponse,
				opts...,
			), "disable_channel").ServeHTTP)

			// Request to add users to a channel
			// This endpoint can be used alternative to /channels/{groupID}/members
			authzMiddleware = api.AuthorizeMiddleware(authClient, assignUsersAuthReq)
			r.Post("/{groupID}/users/assign", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(assignUsersEndpoint(svc)),
				decodeAssignUsersRequest,
				api.EncodeResponse,
				opts...,
			), "assign_users").ServeHTTP)

			// Request to remove users from a channel
			// This endpoint can be used alternative to /channels/{groupID}/members
			r.Post("/{groupID}/users/unassign", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(unassignUsersEndpoint(svc)),
				decodeUnassignUsersRequest,
				api.EncodeResponse,
				opts...,
			), "unassign_users").ServeHTTP)

			// Request to add user_groups to a channel
			// This endpoint can be used alternative to /channels/{groupID}/members
			authzMiddleware = api.AuthorizeMiddleware(authClient, assignUserGroupsAuthReq)
			r.Post("/{groupID}/groups/assign", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(assignUserGroupsEndpoint(svc)),
				decodeAssignUserGroupsRequest,
				api.EncodeResponse,
				opts...,
			), "assign_groups").ServeHTTP)

			// Request to remove user_groups from a channel
			// This endpoint can be used alternative to /channels/{groupID}/members
			r.Post("/{groupID}/groups/unassign", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(unassignUserGroupsEndpoint(svc)),
				decodeUnassignUserGroupsRequest,
				api.EncodeResponse,
				opts...,
			), "unassign_groups").ServeHTTP)

			authzMiddleware = api.AuthorizeMiddleware(authClient, connectAuthReq)
			r.Post("/{groupID}/things/{thingID}/connect", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(connectChannelThingEndpoint(svc)),
				decodeConnectChannelThingRequest,
				api.EncodeResponse,
				opts...,
			), "connect_channel_thing").ServeHTTP)

			r.Post("/{groupID}/things/{thingID}/disconnect", otelhttp.NewHandler(kithttp.NewServer(
				authzMiddleware(disconnectChannelThingEndpoint(svc)),
				decodeDisconnectChannelThingRequest,
				api.EncodeResponse,
				opts...,
			), "disconnect_channel_thing").ServeHTTP)
		})

		// Ideal location: things service,  things endpoint
		// Reason for placing here :
		// SpiceDB provides list of channel ids to which thing id attached
		// and channel service can access spiceDB and get this channel ids list with given thing id.
		// Request to get list of channels to which thingID ({memberID}) belongs
		authzMiddleware := api.AuthorizeMiddleware(authClient, gapi.ListGroupsByThingAuthReq)
		r.Get("/things/{memberID}/channels", otelhttp.NewHandler(kithttp.NewServer(
			authzMiddleware(gapi.ListGroupsEndpoint(svc, authClient, "channels", "things")),
			gapi.DecodeListGroupsRequest,
			api.EncodeResponse,
			opts...,
		), "list_channel_by_thing_id").ServeHTTP)

		// Ideal location: users service, users endpoint
		// Reason for placing here :
		// SpiceDB provides list of channel ids attached to given user id
		// and channel service can access spiceDB and get this user ids list with given thing id.
		// Request to get list of channels to which userID ({memberID}) have permission.
		authzMiddleware = api.AuthorizeMiddleware(authClient, gapi.ListGroupsByUserAuthReq)
		r.Get("/users/{memberID}/channels", otelhttp.NewHandler(kithttp.NewServer(
			checkSuperAdminMiddleware(authzMiddleware(gapi.ListGroupsEndpoint(svc, authClient, "channels", "users"))),
			gapi.DecodeListGroupsRequest,
			api.EncodeResponse,
			opts...,
		), "list_channel_by_user_id").ServeHTTP)

		// Ideal location: users service, groups endpoint
		// SpiceDB provides list of channel ids attached to given user_group id
		// and channel service can access spiceDB and get this user ids list with given user_group id.
		// Request to get list of channels to which user_group_id ({memberID}) attached.
		authzMiddleware = api.AuthorizeMiddleware(authClient, gapi.ListGroupsByGroupAuthReq)
		r.Get("/groups/{memberID}/channels", otelhttp.NewHandler(kithttp.NewServer(
			authzMiddleware(gapi.ListGroupsEndpoint(svc, authClient, "channels", "groups")),
			gapi.DecodeListGroupsRequest,
			api.EncodeResponse,
			opts...,
		), "list_channel_by_user_group_id").ServeHTTP)

		// Connect channel and thing
		authzMiddleware = api.AuthorizeMiddleware(authClient, connectAuthReq)
		r.Post("/connect", otelhttp.NewHandler(kithttp.NewServer(
			authzMiddleware(connectEndpoint(svc)),
			decodeConnectRequest,
			api.EncodeResponse,
			opts...,
		), "connect").ServeHTTP)

		// Disconnect channel and thing
		r.Post("/disconnect", otelhttp.NewHandler(kithttp.NewServer(
			authzMiddleware(disconnectEndpoint(svc)),
			decodeDisconnectRequest,
			api.EncodeResponse,
			opts...,
		), "disconnect").ServeHTTP)
	})

	return r
}

func decodeAssignUsersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := assignUsersRequest{
		token:   apiutil.ExtractBearerToken(r),
		groupID: chi.URLParam(r, "groupID"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
	}

	return req, nil
}

func decodeUnassignUsersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := assignUsersRequest{
		token:   apiutil.ExtractBearerToken(r),
		groupID: chi.URLParam(r, "groupID"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
	}

	return req, nil
}

func decodeAssignUserGroupsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := assignUserGroupsRequest{
		token:   apiutil.ExtractBearerToken(r),
		groupID: chi.URLParam(r, "groupID"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
	}

	return req, nil
}

func decodeUnassignUserGroupsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := assignUserGroupsRequest{
		token:   apiutil.ExtractBearerToken(r),
		groupID: chi.URLParam(r, "groupID"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
	}

	return req, nil
}

func decodeConnectChannelThingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := connectChannelThingRequest{
		token:     apiutil.ExtractBearerToken(r),
		ThingID:   chi.URLParam(r, "thingID"),
		ChannelID: chi.URLParam(r, "groupID"),
	}

	return req, nil
}

func decodeDisconnectChannelThingRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := connectChannelThingRequest{
		token:     apiutil.ExtractBearerToken(r),
		ThingID:   chi.URLParam(r, "thingID"),
		ChannelID: chi.URLParam(r, "groupID"),
	}

	return req, nil
}

func decodeConnectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := connectChannelThingRequest{
		token: apiutil.ExtractBearerToken(r),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}

	return req, nil
}

func decodeDisconnectRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	req := connectChannelThingRequest{
		token: apiutil.ExtractBearerToken(r),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(errors.ErrMalformedEntity, err))
	}

	return req, nil
}

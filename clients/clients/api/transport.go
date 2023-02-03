package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/clients/clients"
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
)

// MakeClientsHandler returns a HTTP handler for API endpoints.
func MakeClientsHandler(svc clients.Service, mux *bone.Mux, logger logger.Logger) {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}

	mux.Post("/clients", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("register_client"))(registrationEndpoint(svc)),
		decodeCreateClientReq,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/clients/:id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("view_client"))(viewClientEndpoint(svc)),
		decodeViewClient,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/clients", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("list_clients"))(listClientsEndpoint(svc)),
		decodeListClients,
		api.EncodeResponse,
		opts...,
	))

	mux.Get("/clients/groups/:groupID/members", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("list_members"))(listMembersEndpoint(svc)),
		decodeListMembersRequest,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/clients/:id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_client_name_and_metadata"))(updateClientEndpoint(svc)),
		decodeUpdateClient,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/clients/:id/tags", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_client_tags"))(updateClientTagsEndpoint(svc)),
		decodeUpdateClientTags,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/clients/:id/identity", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_client_identity"))(updateClientIdentityEndpoint(svc)),
		decodeUpdateClientCredentials,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/clients/:id/secret", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_client_secret"))(updateClientSecretEndpoint(svc)),
		decodeUpdateClientCredentials,
		api.EncodeResponse,
		opts...,
	))

	mux.Patch("/clients/:id/owner", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("update_client_owner"))(updateClientOwnerEndpoint(svc)),
		decodeUpdateClientOwner,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/clients/tokens/issue", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("issue_token"))(issueTokenEndpoint(svc)),
		decodeCredentials,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/clients/tokens/refresh", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("refresh_token"))(refreshTokenEndpoint(svc)),
		decodeRefreshToken,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/clients/:id/enable", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("enable_client"))(enableClientEndpoint(svc)),
		decodeChangeClientStatus,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/clients/:id/disable", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("disable_client"))(disableClientEndpoint(svc)),
		decodeChangeClientStatus,
		api.EncodeResponse,
		opts...,
	))

	mux.GetFunc("/health", mainflux.Health("clients"))
	mux.Handle("/metrics", promhttp.Handler())
}

func decodeViewClient(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewClientReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}

	return req, nil
}

func decodeListClients(_ context.Context, r *http.Request) (interface{}, error) {
	var sid string
	s, err := apiutil.ReadStringQuery(r, api.StatusKey, api.DefClientStatus)
	if err != nil {
		return nil, err
	}
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, err
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, err
	}
	m, err := apiutil.ReadMetadataQuery(r, api.MetadataKey, nil)
	if err != nil {
		return nil, err
	}

	n, err := apiutil.ReadStringQuery(r, api.NameKey, "")
	if err != nil {
		return nil, err
	}
	t, err := apiutil.ReadStringQuery(r, api.TagKey, "")
	if err != nil {
		return nil, err
	}
	oid, err := apiutil.ReadStringQuery(r, api.OwnerKey, "")
	if err != nil {
		return nil, err
	}
	visibility, err := apiutil.ReadStringQuery(r, api.VisibilityKey, api.MyVisibility)
	if err != nil {
		return nil, err
	}
	switch visibility {
	case api.MyVisibility:
		oid = api.MyVisibility
	case api.SharedVisibility:
		sid = api.MyVisibility
	case api.AllVisibility:
		sid = api.MyVisibility
		oid = api.MyVisibility
	}
	st, err := clients.ToStatus(s)
	if err != nil {
		return nil, err
	}
	req := listClientsReq{
		token:    apiutil.ExtractBearerToken(r),
		status:   st,
		offset:   o,
		limit:    l,
		metadata: m,
		name:     n,
		tag:      t,
		sharedBy: sid,
		owner:    oid,
	}
	return req, nil
}

func decodeUpdateClient(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateClientReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateClientTags(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateClientTagsReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateClientCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateClientCredentialsReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateClientOwner(_ context.Context, r *http.Request) (interface{}, error) {
	req := updateClientOwnerReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := loginClientReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeRefreshToken(_ context.Context, r *http.Request) (interface{}, error) {
	req := tokenReq{RefreshToken: apiutil.ExtractBearerToken(r)}

	return req, nil
}
func decodeCreateClientReq(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	var c clients.Client
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	req := createClientReq{
		client: c,
		token:  apiutil.ExtractBearerToken(r),
	}

	return req, nil
}

func decodeChangeClientStatus(_ context.Context, r *http.Request) (interface{}, error) {
	req := changeClientStatusReq{
		token: apiutil.ExtractBearerToken(r),
		id:    bone.GetValue(r, "id"),
	}

	return req, nil
}

func decodeListMembersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	s, err := apiutil.ReadStringQuery(r, api.StatusKey, api.DefClientStatus)
	if err != nil {
		return nil, err
	}
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, err
	}
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, err
	}
	m, err := apiutil.ReadMetadataQuery(r, api.MetadataKey, nil)
	if err != nil {
		return nil, err
	}
	st, err := clients.ToStatus(s)
	if err != nil {
		return nil, err
	}
	req := listMembersReq{
		token: apiutil.ExtractBearerToken(r),
		Page: clients.Page{
			Status:   st,
			Offset:   o,
			Limit:    l,
			Metadata: m,
		},
		groupID: bone.GetValue(r, "groupID"),
	}
	return req, nil
}

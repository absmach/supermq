package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/policies"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakePolicyHandler(csvc clients.Service, psvc policies.Service, mux *bone.Mux, logger logger.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	mux.Post("/connect", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("connect"))(connectEndpoint(psvc)),
		decodeConnectList,
		api.EncodeResponse,
		opts...,
	))

	mux.Put("/disconnect", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("disconnect"))(disconnectEndpoint(psvc)),
		decodeConnectList,
		api.EncodeResponse,
		opts...,
	))

	mux.Put("/channels/:chanId/things/:thingId", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("connect_thing"))(connectThingEndpoint(psvc)),
		decodeConnectThing,
		api.EncodeResponse,
		opts...,
	))

	mux.Delete("/channels/:chanId/things/:thingId", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("disconnect_thing"))(disconnectThingEndpoint(psvc)),
		decodeConnectThing,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/identify", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("identify"))(identifyEndpoint(csvc)),
		decodeIdentify,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/identify/channels/:chanId/access-by-key", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("can_access_by_key"))(canAccessByKeyEndpoint(psvc)),
		decodeCanAccessByKey,
		api.EncodeResponse,
		opts...,
	))

	mux.Post("/identify/channels/:chanId/access-by-id", kithttp.NewServer(
		otelkit.EndpointMiddleware(otelkit.WithOperation("can_access_by_id"))(canAccessByIDEndpoint(psvc)),
		decodeCanAccessByID,
		api.EncodeResponse,
		opts...,
	))
	return mux

}

func decodeConnectThing(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := createPolicyReq{
		token:   apiutil.ExtractBearerToken(r),
		ChanID:  bone.GetValue(r, "chanId"),
		ThingID: bone.GetValue(r, "thingId"),
	}
	return req, nil
}

func decodeConnectList(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := createPoliciesReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeIdentify(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	req := identifyReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeCanAccessByKey(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}

	req := canAccessByKeyReq{
		chanID: bone.GetValue(r, "chanId"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeCanAccessByID(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), api.ContentType) {
		return nil, errors.ErrUnsupportedContentType
	}
	req := canAccessByIDReq{
		chanID: bone.GetValue(r, "chanId"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}

	return req, nil
}

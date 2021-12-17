// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/httputil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/ui"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	contentType = "text/html"
	staticDir   = "ui/web/static"
	token       = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2Mzk2NzczMjUsImlhdCI6MTYzOTY0MTMyNSwiaXNzIjoibWFpbmZsdXguYXV0aCIsInN1YiI6ImZscDFAZW1haWwuY29tIiwiaXNzdWVyX2lkIjoiYzkzY2FmYjMtYjNhNy00ZTdmLWE0NzAtMTVjMTRkOGVkMWUwIiwidHlwZSI6MH0.-tg86JDfSL5hgigNWRhYHEHdg3tQTkZ8hFl2ed-07cM"
	offsetKey   = "offset"
	limitKey    = "limit"
	nameKey     = "name"
	orderKey    = "order"
	dirKey      = "dir"
	metadataKey = "metadata"
	disconnKey  = "disconnected"
	sharedKey   = "shared"
	defOffset   = 0
	defLimit    = 10
)

var (
	errMalformedData     = errors.New("malformed request data")
	errMalformedSubtopic = errors.New("malformed subtopic")
	redirectURL          = ""
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc ui.Service, redirect string, tracer opentracing.Tracer) http.Handler {
	redirectURL = redirect
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	r := bone.New()
	r.Get("/", kithttp.NewServer(
		kitot.TraceServer(tracer, "index")(indexEndpoint(svc)),
		decodeIndexRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_things")(createThingEndpoint(svc)),
		decodeThingCreation,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_thing")(viewThingEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/things/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_thing")(updateThingEndpoint(svc)),
		decodeThingUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/things", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_things")(listThingsEndpoint(svc)),
		decodeListThingsRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id/delete", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_thing")(removeThingEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/channels", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_channels")(createChannelEndpoint(svc)),
		decodeChannelsCreation,
		encodeResponse,
		opts...,
	))

	r.Get("/channels/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_channel")(viewChannelEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/channels/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_channel")(updateChannelEndpoint(svc)),
		decodeChannelUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/channels", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_channels")(listChannelsEndpoint(svc)),
		decodeListChannelsRequest,
		encodeResponse,
		opts...,
	))

	r.Post("/connect", kithttp.NewServer(
		kitot.TraceServer(tracer, "connect_thing")(connectThingEndpoint(svc)),
		decodeConnectThing,
		encodeResponse,
		opts...,
	))

	// r.Post("/connect", kithttp.NewServer(
	// 	kitot.TraceServer(tracer, "connect_thing")(connectEndpoint(svc)),
	// 	decodeConnectList,
	// 	encodeResponse,
	// 	opts...,
	// ))

	r.Get("/channels/:id/delete", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_channel")(removeChannelEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Get("/things/:id/channels", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_channels_by_thing")(listChannelsByThingEndpoint(svc)),
		decodeListByConnection,
		encodeResponse,
		opts...,
	))

	r.Post("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_groups")(createGroupEndpoint(svc)),
		decodeGroupCreation,
		encodeResponse,
		opts...,
	))

	r.Get("/groups", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_groups")(listGroupsEndpoint(svc)),
		decodeListGroupsRequest,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_group")(viewGroupEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/groups/:id", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_group")(updateGroupEndpoint(svc)),
		decodeGroupUpdate,
		encodeResponse,
		opts...,
	))

	r.Get("/groups/:id/delete", kithttp.NewServer(
		kitot.TraceServer(tracer, "remove_group")(removeGroupEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.GetFunc("/version", mainflux.Version("ui"))
	r.Handle("/metrics", promhttp.Handler())

	// Static file handler
	fs := http.FileServer(http.Dir(staticDir))
	r.Handle("/*", fs)

	return r
}

func decodeIndexRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	req := indexReq{
		token: getAuthorization(r),
	}

	return req, nil
}

func decodeThingCreation(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}

	req := createThingsReq{
		token:    getAuthorization(r),
		Name:     r.PostFormValue("name"),
		Metadata: meta,
	}

	return req, nil
}

func getAuthorization(r *http.Request) string {
	return token
	// return r.Header.Get("Authorization")
}

func decodeView(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewResourceReq{
		token: getAuthorization(r),
		id:    bone.GetValue(r, "id"),
	}
	return req, nil
}

func decodeThingUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}

	req := updateThingReq{
		token:    getAuthorization(r),
		id:       bone.GetValue(r, "id"),
		Name:     r.PostFormValue("name"),
		Metadata: meta,
	}
	return req, nil
}

func decodeListThingsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	req := listThingsReq{
		token: getAuthorization(r),
	}

	return req, nil
}

func decodeChannelsCreation(_ context.Context, r *http.Request) (interface{}, error) {

	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}

	req := createChannelsReq{
		token:    getAuthorization(r),
		Name:     r.PostFormValue("name"),
		Metadata: meta,
	}

	return req, nil
}

func decodeChannelUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}

	req := updateChannelReq{
		token:    getAuthorization(r),
		id:       bone.GetValue(r, "id"),
		Name:     r.PostFormValue("name"),
		Metadata: meta,
	}
	return req, nil
}

func decodeListChannelsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	req := listChannelsReq{
		token: getAuthorization(r),
	}

	return req, nil
}

func decodeConnectThing(_ context.Context, r *http.Request) (interface{}, error) {
	r.ParseForm()                  // Parses the request body
	chanId := r.Form.Get("chanId") // x will be "" if parameter is not set
	thingId := r.Form.Get("thingId")
	req := connectThingReq{
		token:   getAuthorization(r),
		ChanID:  chanId,
		ThingID: thingId,
	}
	return req, nil
}

func decodeConnectList(_ context.Context, r *http.Request) (interface{}, error) {
	// if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
	// 	return nil, errors.ErrUnsupportedContentType
	// }
	fmt.Println("decodeConnectList")
	req := connectReq{
		token: getAuthorization(r),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(things.ErrMalformedEntity, err)
	}
	fmt.Printf("req:%v\n", req)
	return req, nil
}

func decodeListByConnection(_ context.Context, r *http.Request) (interface{}, error) {
	o, err := httputil.ReadUintQuery(r, offsetKey, defOffset)
	if err != nil {
		return nil, err
	}

	l, err := httputil.ReadUintQuery(r, limitKey, defLimit)
	if err != nil {
		return nil, err
	}

	c, err := httputil.ReadBoolQuery(r, disconnKey, false)
	if err != nil {
		return nil, err
	}

	or, err := httputil.ReadStringQuery(r, orderKey, "")
	if err != nil {
		return nil, err
	}

	d, err := httputil.ReadStringQuery(r, dirKey, "")
	if err != nil {
		return nil, err
	}

	req := listByConnectionReq{
		token: r.Header.Get("Authorization"),
		id:    bone.GetValue(r, "id"),
		pageMetadata: things.PageMetadata{
			Offset:       o,
			Limit:        l,
			Disconnected: c,
			Order:        or,
			Dir:          d,
		},
	}

	return req, nil
}

func decodeGroupCreation(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}
	req := createGroupsReq{
		token:    getAuthorization(r),
		Name:     r.PostFormValue("name"),
		Metadata: meta,
	}

	return req, nil
}

func decodeListGroupsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	req := listGroupsReq{
		token: getAuthorization(r),
	}

	return req, nil
}

func decodeGroupUpdate(_ context.Context, r *http.Request) (interface{}, error) {
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(r.PostFormValue("metadata")), &meta); err != nil {
		return nil, err
	}

	req := updateGroupReq{
		token:    getAuthorization(r),
		id:       bone.GetValue(r, "id"),
		Name:     r.PostFormValue("name"),
		Metadata: meta,
	}
	return req, nil
}

func decodePayload(body io.ReadCloser) ([]byte, error) {
	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, errMalformedData
	}
	defer body.Close()

	return payload, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)
	ar, ok := response.(uiRes)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	for k, v := range ar.Headers() {
		w.Header().Set(k, v)
	}
	w.WriteHeader(ar.Code())

	if ar.Empty() {
		return nil
	}
	w.Write(ar.html)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch err {
	case errMalformedData, errMalformedSubtopic:
		w.WriteHeader(http.StatusBadRequest)
	case things.ErrUnauthorizedAccess:
		w.WriteHeader(http.StatusForbidden)
	default:
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.PermissionDenied:
				w.WriteHeader(http.StatusForbidden)
			default:
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}

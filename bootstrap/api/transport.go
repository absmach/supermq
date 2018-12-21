package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"nov/bootstrap"
	"strconv"

	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	contentType  = "application/json"
	maxLimit     = 100
	defaultLimit = 10
)

var (
	errUnsupportedContentType = errors.New("unsupported content type")
	errInvalidQueryParams     = errors.New("invalid query params")
	validParams               = []string{"state", "external_id", "mainflux_id", "mainflux_key"}
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc bootstrap.Service, reader bootstrap.ConfigReader) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}
	r := bone.New()

	r.Post("/configs", kithttp.NewServer(
		addEndpoint(svc),
		decodeAddRequest,
		encodeResponse,
		opts...))

	r.Get("/configs/:id", kithttp.NewServer(
		viewEndpoint(svc),
		decodeEntityRequest,
		encodeResponse,
		opts...))

	r.Put("/configs/:id", kithttp.NewServer(
		updateEndpoint(svc),
		decodeUpdateRequest,
		encodeResponse,
		opts...))

	r.Get("/configs", kithttp.NewServer(
		listEndpoint(svc),
		decodeListRequest,
		encodeResponse,
		opts...))

	r.Get("/unknown", kithttp.NewServer(
		listEndpoint(svc),
		decodeUnknownRequest,
		encodeResponse,
		opts...))

	r.Get("/bootstrap/:external_id", kithttp.NewServer(
		bootstrapEndpoint(svc, reader),
		decodeBootstrapRequest,
		encodeResponse,
		opts...))

	r.Put("/state/:id", kithttp.NewServer(
		stateEndpoint(svc),
		decodeStateRequest,
		encodeResponse,
		opts...))

	r.Delete("/configs/:id", kithttp.NewServer(
		removeEndpoint(svc),
		decodeEntityRequest,
		encodeResponse,
		opts...))

	r.GetFunc("/version", mainflux.Version("bootstrap"))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func decodeAddRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Header.Get("Content-Type") != contentType {
		return nil, errUnsupportedContentType
	}

	req := addReq{key: r.Header.Get("Authorization")}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

func decodeUpdateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Header.Get("Content-Type") != contentType {
		return nil, errUnsupportedContentType
	}

	req := updateReq{key: r.Header.Get("Authorization")}
	req.id = bone.GetValue(r, "id")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

func decodeUnknownRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errInvalidQueryParams
	}

	offset, limit, err := parsePagePrams(q)
	if err != nil {
		return nil, err
	}

	req := listReq{
		key:    r.Header.Get("Authorization"),
		filter: bootstrap.Filter{"unknown": "true"},
		offset: offset,
		limit:  limit,
	}

	return req, nil
}

func decodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return nil, errInvalidQueryParams
	}

	offset, limit, err := parsePagePrams(q)
	if err != nil {
		return nil, err
	}

	filter := parseFilter(q)

	req := listReq{
		key:    r.Header.Get("Authorization"),
		filter: filter,
		offset: offset,
		limit:  limit,
	}

	return req, nil
}

func decodeBootstrapRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := bootstrapReq{
		id:  bone.GetValue(r, "external_id"),
		key: r.Header.Get("Authorization"),
	}

	return req, nil
}

func decodeStateRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Header.Get("Content-Type") != contentType {
		return nil, errUnsupportedContentType
	}

	req := changeStateReq{key: r.Header.Get("Authorization")}
	req.id = bone.GetValue(r, "id")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
	}

	return req, nil
}

func decodeEntityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := entityReq{
		key: r.Header.Get("Authorization"),
		id:  bone.GetValue(r, "id"),
	}

	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)
	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}

		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", contentType)

	switch err {
	case errUnsupportedContentType:
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case errInvalidQueryParams, bootstrap.ErrMalformedEntity:
		w.WriteHeader(http.StatusBadRequest)
	case bootstrap.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case bootstrap.ErrUnauthorizedAccess:
		w.WriteHeader(http.StatusForbidden)
	case bootstrap.ErrConflict:
		w.WriteHeader(http.StatusConflict)
	case bootstrap.ErrThings:
		w.WriteHeader(http.StatusServiceUnavailable)
	case io.EOF:
		w.WriteHeader(http.StatusBadRequest)
	default:
		switch err.(type) {
		case *json.SyntaxError:
			w.WriteHeader(http.StatusBadRequest)
		case *json.UnmarshalTypeError:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func parseUint(s string) (uint64, error) {
	if s == "" {
		return 0, nil
	}

	ret, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errInvalidQueryParams
	}
	return ret, nil
}

func parsePagePrams(q url.Values) (uint64, uint64, error) {
	offset, err := parseUint(q.Get("offset"))
	q.Del("offset")
	if err != nil {
		return 0, 0, err
	}

	limit, err := parseUint(q.Get("limit"))
	q.Del("limit")
	if err != nil {
		return 0, 0, err
	}

	if limit < 0 || offset < 0 {
		return 0, 0, bootstrap.ErrMalformedEntity
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	if limit == 0 {
		limit = defaultLimit
	}

	return offset, limit, nil
}

func parseFilter(values url.Values) bootstrap.Filter {
	ret := bootstrap.Filter{}
	for k := range values {
		if contains(validParams, k) {
			ret[k] = values.Get(k)
		}
	}
	return ret
}

func contains(l []string, s string) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
}

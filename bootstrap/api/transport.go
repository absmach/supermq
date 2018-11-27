package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"nov/bootstrap"

	kithttp "github.com/go-kit/kit/transport/http"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const contentType = "application/json"

var (
	errUnsupportedContentType = errors.New("unsupported content type")
	errInvalidQueryParams     = errors.New("invalid query params")
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc bootstrap.Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}
	r := bone.New()

	r.Post("/things", kithttp.NewServer(
		addEndpoint(svc),
		decodeAddRequest,
		encodeResponse,
		opts...))

	r.Get("/bootstrap", kithttp.NewServer(
		bootstrapEndpoint(svc),
		decodeBootstrapRequest,
		encodeResponse,
		opts...))

	r.Put("/status/:id", kithttp.NewServer(
		statusEndpoint(svc),
		decodeStatusRequest,
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

func decodeBootstrapRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := boostrapReq{r.Header.Get("Authorization")}
	return req, nil
}

func decodeStatusRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Header.Get("Content-Type") != contentType {
		return nil, errUnsupportedContentType
	}

	req := changeStatusReq{key: r.Header.Get("Authorization")}
	req.ID = bone.GetValue(r, "id")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, err
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
	case bootstrap.ErrMalformedEntity:
		w.WriteHeader(http.StatusBadRequest)
	case bootstrap.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case bootstrap.ErrUnauthorizedAccess:
		w.WriteHeader(http.StatusForbidden)
	case bootstrap.ErrInvalidID:
		w.WriteHeader(http.StatusServiceUnavailable)
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

package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/users"
)

const contentType = "application/json"

var (
	errUnsupportedContentType = errors.New("unsupported content type")
	logger                    log.Logger
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc users.Service, l log.Logger) http.Handler {
	logger = l

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()

	mux.Post("/users", kithttp.NewServer(
		registrationEndpoint(svc),
		decodeCredentials,
		encodeResponse,
		opts...,
	))

	mux.Post("/tokens", kithttp.NewServer(
		loginEndpoint(svc),
		decodeCredentials,
		encodeResponse,
		opts...,
	))

	return mux
}

func decodeCredentials(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Header.Get("Content-Type") != contentType {
		logger.Warn("Invalid or missing content type.")
		return nil, errUnsupportedContentType
	}

	var user users.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		logger.Warn(fmt.Sprintf("Failed to decode user credentials: %s", err))
		return nil, err
	}

	return userReq{user}, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)

	if ar, ok := response.(apiRes); ok {
		for k, v := range ar.headers() {
			w.Header().Set(k, v)
		}

		w.WriteHeader(ar.code())

		if ar.empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", contentType)

	switch err {
	case users.ErrMalformedEntity:
		w.WriteHeader(http.StatusBadRequest)
	case users.ErrUnauthorizedAccess:
		w.WriteHeader(http.StatusForbidden)
	case users.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case users.ErrConflict:
		w.WriteHeader(http.StatusConflict)
	case errUnsupportedContentType:
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case io.ErrUnexpectedEOF:
		w.WriteHeader(http.StatusBadRequest)
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

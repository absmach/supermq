// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/absmach/supermq"
	api "github.com/absmach/supermq/api/http"
	apiutil "github.com/absmach/supermq/api/http/util"
	smqhttp "github.com/absmach/supermq/http"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	ctSenmlJSON         = "application/senml+json"
	ctSenmlCBOR         = "application/senml+cbor"
	contentType         = "application/json"
	authzHeaderKey      = "Authorization"
	authzQueryKey       = "authorization"
	connHeaderKey       = "Connection"
	connHeaderVal       = "upgrade"
	upgradeHeaderKey    = "Upgrade"
	upgradeHeaderVal    = "websocket"
	readwriteBufferSize = 1024
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  readwriteBufferSize,
		WriteBufferSize: readwriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	errMalformedSubtopic  = errors.New("malformed subtopic")
	errGenSessionID       = errors.New("failed to generate session id")
	errMethodNotAllowed   = errors.New("method not allowed")
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(ctx context.Context, svc smqhttp.Service, resolver messaging.TopicResolver, logger *slog.Logger, instanceID string) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	r := chi.NewRouter()

	r.Handle("/m/{domain}/c/{channel}", messageHandler(ctx, svc, resolver, logger))

	r.Handle("/m/{domain}/c/{channel}/*", messageHandler(ctx, svc, resolver, logger))

	r.Post("/hc/{domain}", otelhttp.NewHandler(kithttp.NewServer(
		healthCheckEndpoint(),
		decodeHealthCheckRequest,
		api.EncodeResponse,
		opts...,
	), "health_check").ServeHTTP)

	r.Get("/health", supermq.Health("http", instanceID))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func decodePublishReq(_ context.Context, r *http.Request) (any, error) {
	ct := r.Header.Get("Content-Type")
	if ct != ctSenmlJSON && ct != contentType && ct != ctSenmlCBOR {
		return nil, errors.Wrap(apiutil.ErrValidation, apiutil.ErrUnsupportedContentType)
	}

	var req publishReq
	_, pass, ok := r.BasicAuth()
	switch {
	case ok:
		req.token = pass
	case !ok:
		req.token = r.Header.Get(authzHeaderKey)
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.ErrMalformedEntity)
	}
	defer r.Body.Close()

	req.msg = &messaging.Message{Payload: payload}

	return req, nil
}

func decodeWSReq(r *http.Request, resolver messaging.TopicResolver, logger *slog.Logger) (connReq, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		switch {
		case r.URL.Query().Get(authzQueryKey) != "":
			password = r.URL.Query().Get(authzQueryKey)
		case r.Header.Get(authzHeaderKey) != "":
			password = r.Header.Get(authzHeaderKey)
		default:
			logger.Debug("Missing authorization key.")
			return connReq{}, errUnauthorizedAccess
		}
	}

	domain := chi.URLParam(r, "domain")
	channel := chi.URLParam(r, "channel")

	domainID, channelID, _, err := resolver.Resolve(r.Context(), domain, channel)
	if err != nil {
		return connReq{}, err
	}

	req := connReq{
		username:  username,
		password:  password,
		channelID: channelID,
		domainID:  domainID,
	}

	subTopic := chi.URLParam(r, "*")

	if subTopic != "" {
		subTopic, err := messaging.ParseSubscribeSubtopic(subTopic)
		if err != nil {
			return connReq{}, err
		}
		req.subtopic = subTopic
	}

	return req, nil
}

func decodeHealthCheckRequest(_ context.Context, r *http.Request) (any, error) {
	var req healthCheckReq
	req.domain = chi.URLParam(r, "domain")
	_, pass, ok := r.BasicAuth()
	switch {
	case ok:
		req.token = pass
	case !ok:
		req.token = r.Header.Get(authzHeaderKey)
	}

	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	return req, nil
}

func encodeError(ctx context.Context, w http.ResponseWriter, err error) {
	switch err {
	case smqhttp.ErrEmptyTopic:
		w.WriteHeader(http.StatusBadRequest)
	case errUnauthorizedAccess:
		w.WriteHeader(http.StatusForbidden)
	case errMalformedSubtopic, errors.ErrMalformedEntity:
		w.WriteHeader(http.StatusBadRequest)
	default:
		api.EncodeError(ctx, err, w)
		return
	}

	if errorVal, ok := err.(errors.Error); ok {
		if err := json.NewEncoder(w).Encode(errorVal); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

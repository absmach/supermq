// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
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
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	ctSenmlJSON      = "application/senml+json"
	ctSenmlCBOR      = "application/senml+cbor"
	contentType      = "application/json"
	connHeaderKey    = "Connection"
	connHeaderVal    = "upgrade"
	upgradeHeaderKey = "Upgrade"
	upgradeHeaderVal = "websocket"

	service             = "ws"
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
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(ctx context.Context, svc smqhttp.Service, logger *slog.Logger, instanceID string) http.Handler {

	r := chi.NewRouter()

	r.HandleFunc("/m/{domainID}/c/{chanID}", messageHandler(ctx, svc, logger))
	r.HandleFunc("/m/{domainID}/c/{chanID}/*", messageHandler(ctx, svc, logger))

	r.Get("/health", supermq.Health("http", instanceID))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func decodePublishReq(_ context.Context, r *http.Request) (interface{}, error) {
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
		req.token = apiutil.ExtractClientSecret(r)
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.ErrMalformedEntity)
	}
	defer r.Body.Close()

	req.msg = &messaging.Message{Payload: payload}

	return req, nil
}

func decodeWSReq(r *http.Request, logger *slog.Logger) (connReq, error) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		authKeys := r.URL.Query()["authorization"]
		if len(authKeys) == 0 {
			logger.Debug("Missing authorization key.")
			return connReq{}, errUnauthorizedAccess
		}
		authKey = authKeys[0]
	}

	domainID := chi.URLParam(r, "domainID")
	chanID := chi.URLParam(r, "chanID")

	req := connReq{
		clientKey: authKey,
		chanID:    chanID,
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
	}
}

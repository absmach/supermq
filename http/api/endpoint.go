// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	api "github.com/absmach/supermq/api/http"
	apiutil "github.com/absmach/supermq/api/http/util"
	smqhttp "github.com/absmach/supermq/http"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/go-kit/kit/endpoint"
)

func messageHandler(ctx context.Context, svc smqhttp.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isWebSocketRequest(r) {
			handleWebSocket(ctx, svc, logger, w, r)
			return
		}
		// Handle HTTP POST for publishing messages
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		req, err := decodePublishReq(ctx, r)
		if err != nil {
			encodeError(ctx, w, err)
			return
		}
		_, err = sendMessageEndpoint()(ctx, req)
		if err != nil {
			encodeError(ctx, w, err)
			return
		}

		err = api.EncodeResponse(ctx, w, publishMessageRes{})
		if err != nil {
			encodeError(ctx, w, err)
		}
	}
}

func sendMessageEndpoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(publishReq)
		if err := req.validate(); err != nil {
			return nil, errors.Wrap(apiutil.ErrValidation, err)
		}

		return publishMessageRes{}, nil
	}
}

func handleWebSocket(ctx context.Context, svc smqhttp.Service, logger *slog.Logger, w http.ResponseWriter, r *http.Request) {
	req, err := decodeWSReq(r, logger)
	if err != nil {
		encodeError(ctx, w, err)
		return
	}

	sessionID, err := generateSessionID()
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to generate session id: %s", err.Error()))
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err.Error()))
		return
	}

	client := smqhttp.NewClient(logger, conn, sessionID)

	client.SetCloseHandler(func(code int, text string) error {
		return svc.Unsubscribe(ctx, sessionID, req.domainID, req.chanID, req.subtopic)
	})

	go client.Start(ctx)

	if err := svc.Subscribe(ctx, sessionID, req.clientKey, req.domainID, req.chanID, req.subtopic, client); err != nil {
		conn.Close()
		return
	}

	logger.Debug(fmt.Sprintf("Successfully upgraded communication to WS on channel %s", req.chanID))
}

func isWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get(connHeaderKey), connHeaderVal) &&
		strings.EqualFold(r.Header.Get(upgradeHeaderKey), upgradeHeaderVal)
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.Wrap(errGenSessionID, err)
	}
	return hex.EncodeToString(b), nil
}

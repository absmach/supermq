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
	"net/url"
	"regexp"
	"strings"

	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/ws"
	"github.com/go-chi/chi/v5"
)

var (
	channelPartRegExp = regexp.MustCompile(`^\/?m\/([\w\-]+)\/c\/([\w\-]+)(\/[^?]*)?(\?.*)?$`)

	errGenSessionID = errors.New("failed to generate session id")
)

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.Wrap(errGenSessionID, err)
	}
	return hex.EncodeToString(b), nil
}

func handshake(ctx context.Context, svc ws.Service, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest(r)
		if err != nil {
			encodeError(w, err)
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

		client := ws.NewClient(logger, conn, sessionID)

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
}

func decodeRequest(r *http.Request) (connReq, error) {
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

	channelParts := channelPartRegExp.FindStringSubmatch(r.RequestURI)
	if len(channelParts) < 3 {
		logger.Warn("Empty channel id or malformed url")
		return connReq{}, errors.ErrMalformedEntity
	}

	subtopic, err := parseSubTopic(channelParts[3])
	if err != nil {
		return connReq{}, err
	}

	req.subtopic = subtopic

	return req, nil
}

func parseSubTopic(subtopic string) (string, error) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", errMalformedSubtopic
	}

	subtopic = strings.ReplaceAll(subtopic, "/", ".")

	elems := strings.Split(subtopic, ".")
	filteredElems := []string{}
	for _, elem := range elems {
		if elem == "" {
			continue
		}

		if len(elem) > 1 && (strings.Contains(elem, "*") || strings.Contains(elem, ">")) {
			return "", errMalformedSubtopic
		}

		filteredElems = append(filteredElems, elem)
	}

	subtopic = strings.Join(filteredElems, ".")

	return subtopic, nil
}

func encodeError(w http.ResponseWriter, err error) {
	var statusCode int

	switch err {
	case ws.ErrEmptyTopic:
		statusCode = http.StatusBadRequest
	case errUnauthorizedAccess:
		statusCode = http.StatusForbidden
	case errMalformedSubtopic, errors.ErrMalformedEntity:
		statusCode = http.StatusBadRequest
	default:
		statusCode = http.StatusNotFound
	}
	logger.Warn(fmt.Sprintf("Failed to authorize: %s", err.Error()))
	w.WriteHeader(statusCode)
}

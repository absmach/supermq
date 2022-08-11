// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/ws"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var channelPartRegExp = regexp.MustCompile(`^/channels/([\w\-]+)/messages(/[^?]*)?(\?.*)?$`)

func handshake(svc ws.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest(r)
		if err != nil {
			encodeError(&req, w, err)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err.Error()))
			return
		}
		req.conn = conn
		client := ws.NewClient(conn, "")

		if err := svc.Subscribe(context.Background(), req.thingKey, req.chanID, req.subtopic, client); err != nil {
			logger.Warn(fmt.Sprintf("Failed to subscribe to broker: %s", err.Error()))
			req.conn.Close()
			return
		}

		// In case previous subscription wasn't closed properly
		if conn == nil {
			conn, err = upgrader.Upgrade(w, r, nil)
			if err != nil {
				logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err.Error()))
				return
			}
			client = ws.NewClient(conn, client.GetID())
		}

		logger.Debug(fmt.Sprintf("Successfully upgraded communication to WS on channel %s", req.chanID))
		msgs := make(chan []byte)

		// Listen for messages received from the chan messages, and publish them to broker
		go client.Process(svc, logger, req.thingKey, req.chanID, req.subtopic, msgs)
		go listen(conn, msgs)
	}
}

func decodeRequest(r *http.Request) (connReq, error) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		authKeys := bone.GetQuery(r, "authorization")
		if len(authKeys) == 0 {
			logger.Debug("Missing authorization key.")
			return connReq{}, errUnauthorizedAccess
		}
		authKey = authKeys[0]
	}

	chanID := bone.GetValue(r, "id")

	req := connReq{
		thingKey: authKey,
		chanID:   chanID,
	}

	channelParts := channelPartRegExp.FindStringSubmatch(r.RequestURI)
	if len(channelParts) < 2 {
		logger.Warn("Empty channel id or malformed url")
		return connReq{}, errors.ErrMalformedEntity
	}

	subtopic, err := parseSubTopic(channelParts[2])
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

	subtopic = strings.Replace(subtopic, "/", ".", -1)

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

func listen(conn *websocket.Conn, msgs chan<- []byte) {
	for {
		// Listen for message from the client, and push them to the msgs channel
		_, payload, err := conn.ReadMessage()

		if websocket.IsUnexpectedCloseError(err) {
			logger.Debug(fmt.Sprintf("Closing WS connection: %s", err.Error()))
			close(msgs)
			return
		}

		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to read message: %s", err.Error()))
			close(msgs)
			return
		}

		msgs <- payload
	}
}

func encodeError(req *connReq, w http.ResponseWriter, err error) {
	statusCode := http.StatusUnauthorized
	if req.conn == nil {
		switch err {
		case ws.ErrEmptyID:
			statusCode = http.StatusServiceUnavailable
		case errUnauthorizedAccess:
			statusCode = http.StatusForbidden
		case errMalformedSubtopic:
			statusCode = http.StatusBadRequest
		case errors.ErrMalformedEntity:
			statusCode = http.StatusBadRequest
		default:
			logger.Warn(fmt.Sprintf("Failed to authorize: %s", err.Error()))
			statusCode = http.StatusServiceUnavailable
		}
		w.WriteHeader(statusCode)
		return
	}

	switch err {
	case ws.ErrEmptyID, ws.ErrEmptyTopic:
		statusCode = http.StatusBadRequest
	case ws.ErrInvalidConnection:
		statusCode = http.StatusUnauthorized
	case ws.ErrFailedConnection:
		statusCode = http.StatusServiceUnavailable
	case errUnauthorizedAccess:
		statusCode = http.StatusForbidden
	case errMalformedSubtopic:
		statusCode = http.StatusBadRequest
	case errors.ErrMalformedEntity:
		statusCode = http.StatusBadRequest
	default:
		logger.Warn(fmt.Sprintf("Failed to authorize: %s", err.Error()))
		switch e, ok := status.FromError(err); {
		case ok:
			switch e.Code() {
			case codes.Unauthenticated:
				statusCode = http.StatusUnauthorized
			case codes.PermissionDenied:
				statusCode = http.StatusForbidden
			case codes.Internal:
				statusCode = http.StatusInternalServerError
			default:
				statusCode = http.StatusInternalServerError
			}
		}
		w.WriteHeader(statusCode)
		if req.conn != nil {
			req.conn.Close()
		}
	}

	if errorVal, ok := err.(errors.Error); ok {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(apiutil.ErrorRes{Err: errorVal.Msg()}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

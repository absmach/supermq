// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
)

var (
	channelPartRegExp = regexp.MustCompile(`^/channels/([\w\-]+)/messages(/[^?]*)?(\?.*)?$`)
)

func handshake(svc ws.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest(r)
		if err != nil {
			switch err {
			case errUnauthorizedAccess:
				w.WriteHeader(http.StatusForbidden)
				return
			default:
				logger.Warn(fmt.Sprintf("Failed to authorize: %s", err.Error()))
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}

		// Upgrade to a new ws connection.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err.Error()))
			return
		}

		// Configure the closing of the ws connection
		conn.SetCloseHandler(func(code int, text string) error {
			return svc.Unsubscribe(context.Background(), req.thingKey, req.chanID, req.subtopic)
		})

		req.conn = conn
		client := ws.NewClient(conn, "")

		// Subscribe using the adapterservice
		if err := svc.Subscribe(context.Background(), req.thingKey, req.chanID, req.subtopic, client); err != nil {
			logger.Warn(fmt.Sprintf("Failed to subscribe to broker: %s", err.Error()))
			if err == ws.ErrFailedConnection {
				req.conn.Close()
				return
			}
		}

		logger.Debug(fmt.Sprintf("Successfully upgraded communication to WS on channel %s", req.chanID))

		msgs := make(chan messaging.Message)

		// Listen for messages received from the chan messages, and publish them to broker
		go publish(svc, msgs, req.thingKey)

		go req.listen(msgs)
	}
}

func publish(svc ws.Service, msgs <-chan messaging.Message, thingKey string) {
	for msg := range msgs {
		svc.Publish(context.Background(), thingKey, msg)
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

	// Starting parsing subtopic from here
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

func (req connReq) listen(msgs chan<- messaging.Message) {
	for {
		// Listen for message from the client, and push them to the msgs channel
		_, payload, err := req.conn.ReadMessage()

		if websocket.IsUnexpectedCloseError(err) {
			logger.Debug(fmt.Sprintf("Closing WS connection: %s", err.Error()))
			return
		}

		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to read message: %s", err.Error()))
			return
		}

		msg := messaging.Message{
			Channel:  req.chanID,
			Subtopic: req.subtopic,
			Protocol: protocol,
			Payload:  payload,
			Created:  time.Now().UnixNano(),
		}

		msgs <- msg
	}
}

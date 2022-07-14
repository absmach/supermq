// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	protocol = "ws"
)

var (
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	errMalformedSubtopic  = errors.New("malformed subtopic")
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	auth              mainflux.ThingsServiceClient
	logger            log.Logger
	channelPartRegExp = regexp.MustCompile(`^/channels/([\w\-]+)/messages(/[^?]*)?(\?.*)?$`)
)

type subscription struct {
	pubID    string
	chanID   string
	subtopic string
	conn     *websocket.Conn
	// client   *ws.Connclient
	channel *ws.Channel
}

// MakeHandler returns http handler with handshake endpoint.
func MakeHandler(svc ws.Service, tc mainflux.ThingsServiceClient, l log.Logger) http.Handler {
	auth = tc
	logger = l

	mux := bone.New()
	mux.GetFunc("/channels/:id/messages", handshake(svc))
	mux.GetFunc("/channels/:id/messages/*", handshake(svc))
	mux.GetFunc("/version", mainflux.Health(protocol))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func handshake(svc ws.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sub, err := authorize(r)
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

		channelParts := channelPartRegExp.FindStringSubmatch(r.RequestURI)
		if len(channelParts) < 2 {
			logger.Warn("Empty channel id or malformed url")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sub.subtopic, err = parseSubTopic(channelParts[2])
		if err != nil {
			logger.Warn("Empty channel id or malformed url")
			w.WriteHeader(http.StatusBadRequest)
		}

		// Create a new ws connection.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err.Error()))
			return
		}

		sub.conn = conn

		c := ws.NewClient(conn, sub.pubID, logger)

		logger.Debug(fmt.Sprintf("Successfully upgraded communication to WS on channel %s", sub.chanID))

		// Subscribe to the channel
		if err := svc.Subscribe(context.Background(), sub.pubID, sub.chanID, sub.subtopic, c); err != nil {
			logger.Warn(fmt.Sprintf("Failed to subscribe to broker: %s", err.Error()))
			if err == ws.ErrFailedConnection {
				sub.conn.Close()
				sub.channel.Closed <- true
				return
			}
		}

		sub.channel = ws.NewChannel()

		wg := new(sync.WaitGroup)
		wg.Add(2)

		go sub.listen(wg)
		go sub.broadcast(wg, svc)

		wg.Wait()

		if err := svc.Unsubscribe(context.Background(), sub.pubID, sub.chanID, sub.subtopic); err != nil {
			logger.Warn(fmt.Sprintf("Failed to unsubscribe from broker: %s", err.Error()))
			sub.conn.Close()
			sub.channel.Closed <- true
			return
		}
	}
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

	subtopic = strings.Join(filteredElems, ",")

	return subtopic, nil
}

func authorize(r *http.Request) (subscription, error) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		authKeys := bone.GetQuery(r, "authorization")
		if len(authKeys) == 0 {
			logger.Debug("Missing authorization key.")
			return subscription{}, errUnauthorizedAccess
		}
		authKey = authKeys[0]
	}

	chanID := bone.GetValue(r, "id")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ar := &mainflux.AccessByKeyReq{
		Token:  authKey,
		ChanID: chanID,
	}

	id, err := auth.CanAccessByKey(ctx, ar)
	if err != nil {
		return subscription{}, err
	}

	logger.Debug(fmt.Sprintf("Successfully authorized client %s on channel %s", id.GetValue(), chanID))

	sub := subscription{
		pubID:  authKey,
		chanID: chanID,
	}

	return sub, nil
}

func (sub subscription) broadcast(wg *sync.WaitGroup, svc ws.Service) {
	defer wg.Done()

	for {
		_, payload, err := sub.conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err) {
			logger.Debug(fmt.Sprintf("Closing WS connection: %s", err.Error()))
			sub.channel.Close()
			return
		}
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to read message: %s", err.Error()))
			return
		}

		msg := messaging.Message{
			Protocol: protocol,
			Channel:  sub.chanID,
			Subtopic: sub.subtopic,
			Payload:  payload,
			Created:  time.Now().UnixNano(),
		}

		if err := svc.Publish(context.Background(), sub.pubID, msg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to publish message to broker(NATS): %s", err.Error()))
			if err == ws.ErrFailedConnection {
				sub.conn.Close()
				sub.channel.Closed <- true
				return
			}
		}

		logger.Debug(fmt.Sprintf("Successfully published message to the channel %s", sub.chanID))
	}
}

func (sub subscription) listen(wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range sub.channel.Messages {
		format := websocket.TextMessage

		if err := sub.conn.WriteMessage(format, msg.Payload); err != nil {
			logger.Warn(fmt.Sprintf("Failed to broadcast message to thing: %s", err))
		}

		logger.Debug("Wrote message successfully")
	}
}

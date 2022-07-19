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
	protocol            = "ws"
	readwriteBufferSize = 1024
)

var (
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	errMalformedSubtopic  = errors.New("malformed subtopic")
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  readwriteBufferSize,
		WriteBufferSize: readwriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	// auth              mainflux.ThingsServiceClient
	logger            log.Logger
	channelPartRegExp = regexp.MustCompile(`^/channels/([\w\-]+)/messages(/[^?]*)?(\?.*)?$`)

	client *ws.Connclient
	// ch     *channel
)

type subscription struct {
	thingKey string // pubID = thingKey (delete this comment later)
	chanID   string
	subtopic string
	conn     *websocket.Conn
}

// MakeHandler returns http handler with handshake endpoint.
func MakeHandler(svc ws.Service, tc mainflux.ThingsServiceClient, l log.Logger) http.Handler {
	// auth = tc
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
		sub, err := getSubscription(r, svc)
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

		getSubstopic(sub, r, w)

		// Create a new ws connection.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err.Error()))
			return
		}

		conn.SetCloseHandler(func(code int, text string) error {
			return svc.Unsubscribe(context.Background(), sub.thingKey, sub.chanID, sub.subtopic)
		})

		sub.conn = conn

		client = ws.NewClient(conn, sub.thingKey, logger)
		// ch = newChannel()

		logger.Debug(fmt.Sprintf("Successfully upgraded communication to WS on channel %s", sub.chanID))

		// start listen() and broadcast()
		sub.startListenBroadcast(svc)
	}
}

func getSubstopic(sub subscription, r *http.Request, w http.ResponseWriter) {
	channelParts := channelPartRegExp.FindStringSubmatch(r.RequestURI)
	if len(channelParts) < 2 {
		logger.Warn("Empty channel id or malformed url")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	subtopic, err := parseSubTopic(channelParts[2])
	if err != nil {
		logger.Warn("Empty channel id or malformed url")
		w.WriteHeader(http.StatusBadRequest)
	}

	sub.subtopic = subtopic
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

func getSubscription(r *http.Request, svc ws.Service) (subscription, error) {
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

	sub := subscription{
		thingKey: authKey,
		chanID:   chanID,
	}

	return sub, nil
}

func (sub subscription) startListenBroadcast(svc ws.Service) {
	if err := svc.Subscribe(context.Background(), sub.thingKey, sub.chanID, sub.subtopic, client); err != nil {
		fmt.Println("Failed to subscribe")
		fmt.Println(err.Error())
		logger.Warn(fmt.Sprintf("Failed to subscribe to broker: %s", err.Error()))
		if err == ws.ErrFailedConnection {
			sub.conn.Close()
			// ch.Close()
			// ch.Closed <- true
			return
		}
	}
	// fmt.Println("subscribed")

	wg := new(sync.WaitGroup)
	wg.Add(1)

	// go sub.listen(wg)
	go sub.broadcast(wg, svc)

	wg.Wait()

	// fmt.Println("unsubscribing")
	// if err := svc.Unsubscribe(context.Background(), sub.thingKey, sub.chanID, sub.subtopic); err != nil {
	// 	logger.Warn(fmt.Sprintf("Failed to unsubscribe from broker: %s", err.Error()))
	// 	sub.conn.Close()
	// 	// ch.Close()
	// 	fmt.Println("ubsubscribed")
	// 	// ch.Closed <- true
	// 	return
	// }
}

func (sub subscription) broadcast(wg *sync.WaitGroup, svc ws.Service) {
	defer wg.Done()

	for {
		// if ch.closed {
		// 	break
		// }

		_, payload, err := sub.conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err) {
			logger.Debug(fmt.Sprintf("Closing WS connection: %s", err.Error()))
			// ch.Close()
			return
		}
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to read message: %s", err.Error()))
			return
		}

		msg := messaging.Message{
			Channel:   sub.chanID,
			Subtopic:  sub.subtopic,
			Publisher: sub.thingKey,
			Protocol:  protocol,
			Payload:   payload,
			Created:   time.Now().UnixNano(),
		}
		fmt.Println("at broadcast(): sub.thingKey = ", sub.thingKey)

		if err := svc.Publish(context.Background(), sub.thingKey, msg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to publish message to broker: %s", err.Error()))
			if err == ws.ErrFailedConnection {
				sub.conn.Close()
				// ch.Close()
				// ch.Closed <- true
				return
			}
		}

		logger.Debug(fmt.Sprintf("Successfully published message to the channel %s", sub.chanID))
	}
}

// func (sub subscription) listen(wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	for msg := range ch.Messages {
// 		if ch.closed {
// 			break
// 		}

// 		format := websocket.TextMessage

// 		if err := sub.conn.WriteMessage(format, msg.Payload); err != nil {
// 			logger.Warn(fmt.Sprintf("Failed to broadcast message to thing: %s", err))
// 		}

// 		logger.Debug("Wrote message successfully")
// 	}
// }

package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	manager "github.com/mainflux/mainflux/manager/client"
	"github.com/mainflux/mainflux/ws"
)

const protocol = "ws"

var (
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	errNotFound           = errors.New("non-existent entity")
	upgrader              = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	auth   manager.ManagerClient
	logger log.Logger
)

// MakeHandler returns http handler with handshake endpoint.
func MakeHandler(svc ws.Service, mc manager.ManagerClient, l log.Logger) http.Handler {
	auth = mc
	logger = l

	mux := bone.New()
	mux.GetFunc("/channels/:id/messages", handshake(svc))

	return mux
}

func handshake(svc ws.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sub, err := authorize(r)
		if err == errNotFound {
			logger.Warn(fmt.Sprintf("Invalid channel id: %s", err))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to authorize: %s", err))
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Create new ws connection.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err))
			return
		}

		// Subscribe to channel
		channel := ws.Channel{make(chan mainflux.RawMessage), make(chan bool)}
		if err = svc.Subscribe(sub.chanID, channel); err != nil {
			logger.Warn(fmt.Sprintf("Failed to subscribe to NATS subject: %s", err))
			w.WriteHeader(http.StatusExpectationFailed)
			return
		}
		go broadcast(conn, channel.Messages)

		// Start listening for messages from NATS.
		go listen(svc, sub, conn, channel)
	}
}

func authorize(r *http.Request) (subscription, error) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		authKeys := bone.GetQuery(r, "authorization")
		if len(authKeys) == 0 {
			return subscription{}, errUnauthorizedAccess
		}
		authKey = authKeys[0]
	}

	// Extract ID from /channels/:id/messages.
	chanID := bone.GetValue(r, "id")
	if !govalidator.IsUUID(chanID) {
		return subscription{}, errNotFound
	}

	pubID, err := auth.CanAccess(chanID, authKey)
	if err != nil {
		return subscription{}, errUnauthorizedAccess
	}

	return subscription{pubID, chanID}, nil
}

func listen(svc ws.Service, sub subscription, conn *websocket.Conn, channel ws.Channel) {
	for {
		_, payload, err := conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err) {
			channel.Closed <- true
			return
		}
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to read message: %s", err))
			return
		}
		msg := mainflux.RawMessage{
			Channel:   sub.chanID,
			Publisher: sub.pubID,
			Protocol:  protocol,
			Payload:   payload,
		}
		if err := svc.Publish(msg); err != nil {
			logger.Warn(fmt.Sprintf("Failed to publish message to NATS: %s", err))
			if err == ws.ErrFailedConnection {
				channel.Closed <- true
				return
			}
		}
	}
}

func broadcast(conn *websocket.Conn, messages <-chan mainflux.RawMessage) {
	for msg := range messages {
		if err := conn.WriteMessage(websocket.TextMessage, msg.Payload); err != nil {
			logger.Warn(fmt.Sprintf("Failed to broadcast message to client: %s", err))
		}
	}
}

type subscription struct {
	pubID  string
	chanID string
}

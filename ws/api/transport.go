package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	manager "github.com/mainflux/mainflux/manager/client"
	"github.com/mainflux/mainflux/ws"
	broker "github.com/nats-io/go-nats"
)

var (
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	upgrader              = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	auth   manager.ManagerClient
	nc     *broker.Conn
	logger log.Logger
)

// MakeHandler returns http handler with handshake endpoint.
func MakeHandler(svc ws.Service, mc manager.ManagerClient, bc *broker.Conn, l log.Logger) http.Handler {
	auth = mc
	nc = bc
	logger = l

	mux := bone.New()
	mux.GetFunc("/channels/:id/messages", handshake(svc))

	return mux
}

func handshake(svc ws.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sub, err := authorize(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Create new ws connection.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Log("error", fmt.Sprintf("Failed to upgrade connection to websocket: %s", err))
			return
		}
		socket := ws.NewSocket(conn)

		brokerSub, err := svc.Subscribe(sub.ChanID, func(msg mainflux.RawMessage) {
			err := svc.Broadcast(msg, func(msg mainflux.RawMessage) error {
				if err := socket.Write(msg); err != nil {
					logger.Log("error", fmt.Sprintf("Failed to send message: %s", err))
					return err
				}
				return nil
			})
			if err != nil {
				logger.Log("error", fmt.Sprintf("Failed to broadcast message: %s", err))
			}
		})
		if err != nil {
			logger.Log("error", fmt.Sprintf("Failed to subscribe to NATS subject: %s", err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Listen on ws connection.
		go svc.Listen(socket, sub, func() {
			brokerSub.Unsubscribe()
		})
	}
}

func authorize(r *http.Request) (ws.Subscription, error) {
	apiKeys := bone.GetQuery(r, "auth")
	if len(apiKeys) == 0 {
		return ws.Subscription{}, errUnauthorizedAccess
	}
	apiKey := apiKeys[0]

	// extract ID from /channels/:id/messages
	chanID := bone.GetValue(r, "id")

	pubID, err := auth.CanAccess(chanID, apiKey)
	if err != nil {
		logger.Log("error", "Failed to authorize: %s", err)
		return ws.Subscription{}, errUnauthorizedAccess
	}

	return ws.Subscription{pubID, chanID}, nil
}

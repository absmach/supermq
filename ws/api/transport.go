package api

import (
	"errors"
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	manager "github.com/mainflux/mainflux/manager/client"
	"github.com/mainflux/mainflux/ws"
)

const protocol = "ws"

var (
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	upgrader              = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	auth manager.ManagerClient
)

// MakeHandler returns http handler with handshake endpoint.
func MakeHandler(svc ws.Service, mc manager.ManagerClient) http.Handler {
	auth = mc

	mux := bone.New()
	mux.GetFunc("/channels/:id/messages", handshake(svc))

	return mux
}

func handshake(svc ws.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		pair, err := authorize(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Create new ws connection.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		svc.AddConnection(pair, conn)

		// Listen on ws connection.
		go func() {
			for {
				_, payload, err := conn.ReadMessage()
				if websocket.IsUnexpectedCloseError(err) {
					return
				}
				if err != nil {
					continue
				}
				msg := mainflux.RawMessage{
					Channel:   pair.ChanID,
					Publisher: pair.PubID,
					Protocol:  protocol,
					Payload:   payload,
				}
				svc.Publish(msg)
			}
		}()
	}
}

func authorize(r *http.Request) (ws.IDPair, error) {
	apiKeys := bone.GetQuery(r, "auth")
	if len(apiKeys) == 0 {
		return ws.IDPair{}, errUnauthorizedAccess
	}
	apiKey := apiKeys[0]

	// extract ID from /channels/:id/messages
	cid := bone.GetValue(r, "id")

	id, err := auth.CanAccess(cid, apiKey)
	if err != nil {
		return ws.IDPair{}, errUnauthorizedAccess
	}

	return ws.IDPair{id, cid}, nil
}

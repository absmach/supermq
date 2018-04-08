package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	manager "github.com/mainflux/mainflux/manager/client"
	broker "github.com/nats-io/go-nats"
	"github.com/go-zoo/bone"
)

var (
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	auth   manager.ManagerClient
	nc     *broker.Conn
	logger log.Logger
)

// MakeHandler returns http handler with handshake endpoint.
func MakeHandler(svc mainflux.MessagePubSub, mc manager.ManagerClient, bc *broker.Conn, l log.Logger) {
	auth = mc
	nc = bc
	logger = l

	// TODO

	return
}

func handshake(svc mainflux.MessagePubSub) {
	// TODO
	return
}

func authorize(r *http.Request) (mainflux.Subscription, error) {
	apiKeys := bone.GetQuery(r, "auth")
	if len(apiKeys) == 0 {
		return mainflux.Subscription{}, errUnauthorizedAccess
	}
	apiKey := apiKeys[0]

	// extract ID from /channels/:id/messages
	chanID := bone.GetValue(r, "id")

	pubID, err := auth.CanAccess(chanID, apiKey)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to authorize: %s", err))
		return mainflux.Subscription{}, errUnauthorizedAccess
	}

	sub := mainflux.Subscription{
		PubID:  pubID,
		ChanID: chanID,
	}
	return sub, nil
}

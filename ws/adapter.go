package ws

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	manager "github.com/mainflux/mainflux/manager/client"
	broker "github.com/nats-io/go-nats"
)

var _ Service = (*adapterService)(nil)

// Service contains publish and subscribe methods necessary for
// message transfer.
type Service interface {
	mainflux.MessagePublisher

	// HandleMessage that is received from message broker.
	HandleMessage(*broker.Msg)

	// AddConnection adds new client ws connection for given client and channel.
	AddConnection(string, string, *websocket.Conn)
}

type adapterService struct {
	pub   mainflux.MessagePublisher
	mc    manager.ManagerClient
	conns map[string]map[string]*websocket.Conn
}

// New instantiates the domain service implementation.
func New(pub mainflux.MessagePublisher) Service {
	return &adapterService{
		pub:   pub,
		conns: make(map[string]map[string]*websocket.Conn),
	}
}

func (as *adapterService) Publish(msg mainflux.RawMessage) error {
	return as.pub.Publish(msg)
}

func (as *adapterService) HandleMessage(msg *broker.Msg) {
	var rawMsg mainflux.RawMessage
	if err := json.Unmarshal(msg.Data, &rawMsg); err != nil {
		return
	}

	cid := rawMsg.Channel
	for pid, conn := range as.conns[cid] {
		if rawMsg.Publisher != pid {
			conn.WriteJSON(rawMsg)
		}
	}

	return
}

func (as *adapterService) AddConnection(channelID, publisherID string, conn *websocket.Conn) {
	if _, ok := as.conns[channelID]; !ok {
		as.conns[channelID] = make(map[string]*websocket.Conn)
	}

	if oldConn, ok := as.conns[channelID][publisherID]; ok {
		oldConn.Close()
	}

	as.conns[channelID][publisherID] = conn

	// On close delete connection from map of connections.
	conn.SetCloseHandler(func(code int, text string) error {
		delete(as.conns[channelID], publisherID)
		return conn.CloseHandler()(code, text)
	})
}

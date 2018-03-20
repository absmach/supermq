package ws

import (
	"errors"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
)

const protocol = "ws"

var _ Service = (*adapterService)(nil)

// ErrFailedMessagePublish indicates that message publishing failed.
var ErrFailedMessagePublish = errors.New("failed to publish message")

// Service contains publish and subscribe methods necessary for
// message transfer.
type Service interface {
	mainflux.MessagePublisher

	// Broadcast broadcasts raw message to channel.
	Broadcast(mainflux.RawMessage)

	// AddConnection adds new client ws connection for given client and channel.
	AddConnection(Subscription, *websocket.Conn)

	// Listen starts loop for receiving messages over connection.
	Listen(Subscription)
}

type adapterService struct {
	pub    mainflux.MessagePublisher
	conns  map[string]map[string]socket
	logger log.Logger
}

// New instantiates the domain service implementation.
func New(pub mainflux.MessagePublisher, logger log.Logger) Service {
	return &adapterService{
		pub:    pub,
		conns:  make(map[string]map[string]socket),
		logger: logger,
	}
}

func (as *adapterService) Publish(msg mainflux.RawMessage) error {
	as.Broadcast(msg)
	if err := as.pub.Publish(msg); err != nil {
		as.logger.Log("error", fmt.Sprintf("Failed to publish message: %s", err))
		return ErrFailedMessagePublish
	}
	return nil
}

func (as *adapterService) Broadcast(msg mainflux.RawMessage) {
	chanID := msg.Channel
	for _, conn := range as.conns[chanID] {
		go func(sock socket) {
			if err := sock.write(msg); err != nil {
				as.logger.Log("error", "Failed to write message: %s", err)
			}
		}(conn)
	}
}

func (as *adapterService) AddConnection(sub Subscription, conn *websocket.Conn) {
	if _, ok := as.conns[sub.ChanID]; !ok {
		as.conns[sub.ChanID] = make(map[string]socket)
	}

	if oldConn, ok := as.conns[sub.ChanID][sub.PubID]; ok {
		oldConn.Close()
	}

	as.conns[sub.ChanID][sub.PubID] = socket{conn, &sync.Mutex{}}

	// On close delete connection from map of connections.
	conn.SetCloseHandler(func(code int, text string) error {
		delete(as.conns[sub.ChanID], sub.PubID)
		return conn.CloseHandler()(code, text)
	})
}

func (as *adapterService) Listen(sub Subscription) {
	conn := as.conns[sub.ChanID][sub.PubID]
	for {
		_, payload, err := conn.ReadMessage()
		if websocket.IsUnexpectedCloseError(err) {
			return
		}
		if err != nil {
			as.logger.Log("error", fmt.Sprintf("Failed to read message: %s", err))
			continue
		}
		msg := mainflux.RawMessage{
			Channel:   sub.ChanID,
			Publisher: sub.PubID,
			Protocol:  protocol,
			Payload:   payload,
		}
		as.Publish(msg)
	}
}

// Subscription contains publisher and channel id.
type Subscription struct {
	PubID  string
	ChanID string
}

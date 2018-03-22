package ws

import (
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
)

const protocol = "ws"

var (
	_ Service = (*adapterService)(nil)

	// ErrFailedMessagePublish indicates that message publishing failed.
	ErrFailedMessagePublish = errors.New("failed to publish message")
	// ErrFailedMessageBroadcast indicates that message broadcast failed.
	ErrFailedMessageBroadcast = errors.New("failed to broadcast message")
)

// Service contains publish and subscribe methods necessary for
// message transfer.
type Service interface {
	mainflux.MessagePublisher

	// Broadcast broadcasts raw message to channel.
	Broadcast(Socket, mainflux.RawMessage) error

	// Listen starts loop for receiving messages over connection.
	Listen(Socket, Subscription, func())
}

type adapterService struct {
	pub    mainflux.MessagePublisher
	logger log.Logger
}

// New instantiates the domain service implementation.
func New(pub mainflux.MessagePublisher, logger log.Logger) Service {
	return &adapterService{
		pub:    pub,
		logger: logger,
	}
}

func (as *adapterService) Publish(msg mainflux.RawMessage) error {
	if err := as.pub.Publish(msg); err != nil {
		as.logger.Log("error", fmt.Sprintf("Failed to publish message: %s", err))
		return ErrFailedMessagePublish
	}
	return nil
}

func (as *adapterService) Broadcast(socket Socket, msg mainflux.RawMessage) error {
	if err := socket.write(msg); err != nil {
		as.logger.Log("error", "Failed to write message: %s", err)
		return ErrFailedMessageBroadcast
	}
	return nil
}

func (as *adapterService) Listen(socket Socket, sub Subscription, onClose func()) {
	for {
		_, payload, err := socket.ReadMessage()
		if websocket.IsUnexpectedCloseError(err) {
			onClose()
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

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
	// ErrFailedSubscription indicates that client couldn't subscribe to specified channel.
	ErrFailedSubscription = errors.New("failed to subscribe to a channel")
)

// Service contains publish and subscribe methods necessary for
// message transfer.
type Service interface {
	mainflux.MessageBroker
	// Listen starts loop for receiving messages over connection.
	Listen(Socket, Subscription, func())
}

type adapterService struct {
	pubsub mainflux.MessagePubSub
	logger log.Logger
}

// New instantiates the domain service implementation.
func New(pubsub mainflux.MessagePubSub, logger log.Logger) Service {
	return &adapterService{pubsub, logger}
}

func (as *adapterService) Publish(msg mainflux.RawMessage) error {
	if err := as.pubsub.Publish(msg); err != nil {
		as.logger.Log("error", fmt.Sprintf("Failed to publish message: %s", err))
		return ErrFailedMessagePublish
	}
	return nil
}

func (as *adapterService) Broadcast(msg mainflux.RawMessage, sendMsg func(msg mainflux.RawMessage) error) error {
	if err := sendMsg(msg); err != nil {
		as.logger.Log("error", fmt.Sprintf("Failed to write message: %s", err))
		return ErrFailedMessageBroadcast
	}
	return nil
}

func (as *adapterService) Subscribe(channel string, onMessage func(mainflux.RawMessage)) (mainflux.Subscription, error) {
	sub, err := as.pubsub.Subscribe(channel, onMessage)
	if err != nil {
		as.logger.Log("error", fmt.Sprintf("Failed to subscribe to a channel: %s", err))
		return nil, ErrFailedSubscription
	}
	return sub, nil
}

func (as *adapterService) Listen(socket Socket, sub Subscription, onClose func()) {
	defer onClose()
	for {
		_, payload, err := socket.ReadMessage()
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

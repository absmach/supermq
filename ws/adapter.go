package ws

import (
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
)

const protocol = "ws"

var _ mainflux.MessagePubSub = (*adapterService)(nil)

var (
	// ErrFailedMessagePublish indicates that message publishing failed.
	ErrFailedMessagePublish = errors.New("failed to publish message")
	// ErrFailedMessageBroadcast indicates that message broadcast failed.
	ErrFailedMessageBroadcast = errors.New("failed to broadcast message")
	// ErrFailedSubscription indicates that client couldn't subscribe to specified channel.
	ErrFailedSubscription = errors.New("failed to subscribe to a channel")
)

type adapterService struct {
	pubsub mainflux.MessagePubSub
	logger log.Logger
}

// New instantiates the domain service implementation.
func New(pubsub mainflux.MessagePubSub, logger log.Logger) mainflux.MessagePubSub {
	return &adapterService{pubsub, logger}
}

func (as *adapterService) Publish(msg mainflux.RawMessage) error {
	if err := as.pubsub.Publish(msg); err != nil {
		as.logger.Log("error", fmt.Sprintf("Failed to publish message: %s", err))
		return ErrFailedMessagePublish
	}
	return nil
}

func (as *adapterService) Subscribe(sub mainflux.Subscription, write mainflux.WriteMessage, read mainflux.ReadMessage) (func(), error) {
	unsubscribe, err := as.pubsub.Subscribe(sub, write, nil)
	if err != nil {
		as.logger.Log("error", fmt.Sprintf("Failed to subscribe to a channel: %s", err))
		return nil, ErrFailedSubscription
	}
	go as.listen(sub, read, func() {
		unsubscribe()
	})
	return nil, nil
}

func (as *adapterService) listen(sub mainflux.Subscription, read mainflux.ReadMessage, onClose func()) {
	defer onClose()
	for {
		payload, err := read()
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
		if err := as.Publish(msg); err != nil {
			as.logger.Log("error", "Failed to publish message to NATS: %s", err)
		}
	}
}

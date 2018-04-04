package ws

import (
	"errors"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
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

func (as *adapterService) Publish(msg mainflux.RawMessage, cfHandler mainflux.ConnFailHandler) error {
	if err := as.pubsub.Publish(msg, cfHandler); err != nil {
		as.logger.Warn(fmt.Sprintf("Failed to publish message: %s", err))
		return ErrFailedMessagePublish
	}
	return nil
}

func (as *adapterService) Subscribe(sub mainflux.Subscription, cfHandler mainflux.ConnFailHandler) (mainflux.Unsubscribe, error) {
	unsubscribe, err := as.pubsub.Subscribe(sub, nil)
	if err != nil {
		as.logger.Warn(fmt.Sprintf("Failed to subscribe to a channel: %s", err))
		return nil, ErrFailedSubscription
	}
	go as.listen(sub, cfHandler, func() {
		go unsubscribe()
	})
	return nil, nil
}

func (as *adapterService) listen(sub mainflux.Subscription, cfHandler mainflux.ConnFailHandler, onClose func()) {
	defer onClose()
	for {
		payload, err := sub.Read()
		if websocket.IsUnexpectedCloseError(err) {
			return
		}
		if err != nil {
			as.logger.Warn(fmt.Sprintf("Failed to read message: %s", err))
			return
		}
		msg := mainflux.RawMessage{
			Channel:   sub.ChanID,
			Publisher: sub.PubID,
			Protocol:  protocol,
			Payload:   payload,
		}
		if err := as.Publish(msg, cfHandler); err != nil {
			as.logger.Warn(fmt.Sprintf("Failed to publish message to NATS: %s", err))
		}
	}
}

package mqtt

import (
	"errors"
	"fmt"

	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
)

const protocol = "mqtt"

var _ mainflux.MessagePubSub = (*adapter)(nil)

var (
	// ErrFailedMessagePublish indicates that message publishing failed.
	ErrFailedMessagePublish = errors.New("failed to publish message")
	// ErrFailedMessageBroadcast indicates that message broadcast failed.
	ErrFailedMessageBroadcast = errors.New("failed to broadcast message")
	// ErrFailedSubscription indicates that client couldn't subscribe to specified channel.
	ErrFailedSubscription = errors.New("failed to subscribe to a channel")
)

type adapter struct {
	pubsub mainflux.MessagePubSub
	logger log.Logger
}

// New instantiates the domain service implementation.
func New(pubsub mainflux.MessagePubSub, logger log.Logger) mainflux.MessagePubSub {
	return &adapter{pubsub, logger}
}

func (a *adapter) Publish(msg mainflux.RawMessage, cfHandler mainflux.ConnFailHandler) error {
	if err := a.pubsub.Publish(msg, cfHandler); err != nil {
		a.logger.Warn(fmt.Sprintf("Failed to publish message: %s", err))
		return ErrFailedMessagePublish
	}
	return nil
}

func (a *adapter) Subscribe(sub mainflux.Subscription, cfHandler mainflux.ConnFailHandler) (mainflux.Unsubscribe, error) {
	unsubscribe, err := a.pubsub.Subscribe(sub, nil)
	if err != nil {
		a.logger.Warn(fmt.Sprintf("Failed to subscribe to a channel: %s", err))
		return nil, ErrFailedSubscription
	}
	go a.listen(sub, cfHandler, func() {
		go unsubscribe()
	})
	return nil, nil
}

func (a *adapter) listen(sub mainflux.Subscription, cfHandler mainflux.ConnFailHandler, onClose func()) {
	defer onClose()
	for {
		payload, err := sub.Read()
		//if websocket.IsUnexpectedCloseError(err) {
		//	return
		//}
		if err != nil {
			a.logger.Warn(fmt.Sprintf("Failed to read message: %s", err))
			return
		}
		msg := mainflux.RawMessage{
			Channel:   sub.ChanID,
			Publisher: sub.PubID,
			Protocol:  protocol,
			Payload:   payload,
		}
		if err := a.Publish(msg, cfHandler); err != nil {
			a.logger.Warn(fmt.Sprintf("Failed to publish message to NATS: %s", err))
		}
	}
}

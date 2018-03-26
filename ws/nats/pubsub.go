// Package nats contains NATS message publisher implementation.
package nats

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	broker "github.com/nats-io/go-nats"
)

const prefix string = "channel"

var _ mainflux.MessagePubSub = (*natsPubSub)(nil)

type natsPubSub struct {
	nc     *broker.Conn
	logger log.Logger
}

// New instantiates NATS message publisher.
func New(nc *broker.Conn, logger log.Logger) mainflux.MessagePubSub {
	return &natsPubSub{nc, logger}
}

func (pubsub *natsPubSub) Publish(msg mainflux.RawMessage) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	return pubsub.nc.Publish(fmt.Sprintf("%s.%s", prefix, msg.Channel), data)
}

func (pubsub *natsPubSub) Subscribe(channel string, onMessage func(mainflux.RawMessage)) (mainflux.Subscription, error) {
	return pubsub.nc.Subscribe(fmt.Sprintf("%s.%s", prefix, channel), func(msg *broker.Msg) {
		if msg == nil {
			pubsub.logger.Log("error", fmt.Sprintf("Received empty message"))
			return
		}
		var rawMsg mainflux.RawMessage
		if err := proto.Unmarshal(msg.Data, &rawMsg); err != nil {
			pubsub.logger.Log("error", fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}
		onMessage(rawMsg)
	})
}

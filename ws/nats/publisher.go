// Package nats contains NATS message publisher implementation.
package nats

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	broker "github.com/nats-io/go-nats"
)

const topic string = "src.ws"

var _ mainflux.MessagePublisher = (*natsPublisher)(nil)

type natsPublisher struct {
	nc *broker.Conn
}

// New instantiates NATS message publisher.
func New(nc *broker.Conn) mainflux.MessagePublisher {
	return &natsPublisher{nc}
}

func (pub *natsPublisher) Publish(msg mainflux.RawMessage) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	return pub.nc.Publish(fmt.Sprintf("%s.%s", topic, msg.Channel), data)
}

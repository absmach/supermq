// Package nats contains NATS message publisher implementation.
package nats

import (
	"fmt"

	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/log"
	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	broker "github.com/nats-io/go-nats"
)

const (
	prefix          = "channel"
	maxFailedReqs   = 3
	maxFailureRatio = 0.6
)

var _ mainflux.MessagePubSub = (*natsPubSub)(nil)

type natsPubSub struct {
	nc     *broker.Conn
	cb     *gobreaker.CircuitBreaker
	logger log.Logger
}

// New instantiates NATS message publisher.
func New(nc *broker.Conn, logger log.Logger) mainflux.MessagePubSub {
	st := gobreaker.Settings{
		Name: "NATS",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			fr := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= maxFailedReqs && fr >= maxFailureRatio
		},
	}
	cb := gobreaker.NewCircuitBreaker(st)
	return &natsPubSub{nc, cb, logger}
}

func (pubsub *natsPubSub) Publish(msg mainflux.RawMessage, cfHandler mainflux.ConnFailHandler) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	err = pubsub.nc.Publish(fmt.Sprintf("%s.%s", prefix, msg.Channel), data)
	if err == broker.ErrConnectionClosed || err == broker.ErrInvalidConnection {
		cfHandler()
	}

	return err
}

func (pubsub *natsPubSub) Subscribe(subscription mainflux.Subscription, _ mainflux.ConnFailHandler) (mainflux.Unsubscribe, error) {
	sub, err := pubsub.nc.Subscribe(fmt.Sprintf("%s.%s", prefix, subscription.ChanID), func(msg *broker.Msg) {
		if msg == nil {
			pubsub.logger.Log("error", fmt.Sprintf("Received empty message"))
			return
		}

		var rawMsg mainflux.RawMessage
		if err := proto.Unmarshal(msg.Data, &rawMsg); err != nil {
			pubsub.logger.Log("error", fmt.Sprintf("Failed to unmarshal received message: %s", err))
			return
		}

		if err := subscription.Write(rawMsg); err != nil {
			pubsub.logger.Log("error", fmt.Sprintf("On message operation failed: %s", err))
		}
	})
	return func() error {
		_, err := pubsub.cb.Execute(func() (interface{}, error) {
			return nil, sub.Unsubscribe()
		})
		if err != nil {
			pubsub.logger.Log("error", fmt.Sprintf("Failed to unsubscribe from channel: %s", err))
		}
		return err
	}, err
}

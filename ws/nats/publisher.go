//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// Package nats contains NATS message publisher implementation.
package nats

import (
	"fmt"

	"github.com/sony/gobreaker"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
	broker "github.com/nats-io/go-nats"
)

const (
	prefix          = "channel"
	maxFailedReqs   = 3
	maxFailureRatio = 0.6
)

var _ ws.Service = (*natsPubSub)(nil)

type natsPubSub struct {
	nc *broker.Conn
	cb *gobreaker.CircuitBreaker
}

// New instantiates NATS message publisher.
func New(nc *broker.Conn) ws.Service {
	st := gobreaker.Settings{
		Name: "NATS",
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			fr := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= maxFailedReqs && fr >= maxFailureRatio
		},
	}
	cb := gobreaker.NewCircuitBreaker(st)
	return &natsPubSub{nc, cb}
}

func (pubsub *natsPubSub) getDestChannel(chanID, subtopic string) string {
	destChannel := fmt.Sprintf("%s.%s", prefix, chanID)
	if subtopic != "" {
		destChannel += "." + subtopic
	}
	return destChannel
}

func (pubsub *natsPubSub) Publish(msg mainflux.RawMessage) error {
	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	// TODO actually if someone subscribe to some channel with jolly chars, publish
	//	does not work, return an error message or silently fail?
	return pubsub.nc.Publish(pubsub.getDestChannel(msg.Channel, msg.Subtopic), data)
}

func (pubsub *natsPubSub) Subscribe(chanID, subtopic string, channel *ws.Channel) error {
	var sub *broker.Subscription

	sub, err := pubsub.nc.Subscribe(pubsub.getDestChannel(chanID, subtopic), func(msg *broker.Msg) {
		if msg == nil {
			return
		}

		var rawMsg mainflux.RawMessage
		if err := proto.Unmarshal(msg.Data, &rawMsg); err != nil {
			return
		}

		// Sends message to messages channel
		channel.Send(rawMsg)
	})

	// Check if subscription should be closed
	go func() {
		<-channel.Closed
		sub.Unsubscribe()
	}()

	return err
}

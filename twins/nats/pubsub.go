//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package nats

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/nats-io/go-nats"
)

const (
	queue         = "twins"
	input         = "channel.>"
	outputUnknown = "out.unknown"
	senML         = "application/senml+json"
)

type pubsub struct {
	nc     *nats.Conn
	logger log.Logger
}

// Subscribe to appropriate NATS topic and normalizes received messages.
func Subscribe(nc *nats.Conn, logger log.Logger) {
	ps := pubsub{
		nc:     nc,
		logger: logger,
	}
	ps.nc.QueueSubscribe(input, queue, ps.handleMsg)
}

func (ps pubsub) handleMsg(m *nats.Msg) {
	var msg mainflux.RawMessage
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshalling failed: %s", err))
		return
	}

	if err := ps.publish(msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Publishing failed: %s", err))
		return
	}
}

func (ps pubsub) publish(msg mainflux.RawMessage) error {
	output := mainflux.OutputSenML

	data, err := proto.Marshal(&msg)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Marshalling failed: %s", err))
		return err
	}

	if err := ps.nc.Publish(output, data); err != nil {
		ps.logger.Warn(fmt.Sprintf("Publishing failed: %s", err))
		return err
	}

	return nil
}

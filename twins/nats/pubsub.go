//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/twins"
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
	tr     twins.TwinRepository
}

// Subscribe to appropriate NATS topic
func Subscribe(nc *nats.Conn, tr twins.TwinRepository, logger log.Logger) {
	ps := pubsub{
		nc:     nc,
		logger: logger,
		tr:     tr,
	}
	ps.nc.QueueSubscribe(input, queue, ps.handleMsg)
}

func (ps pubsub) handleMsg(m *nats.Msg) {
	var msg mainflux.RawMessage
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshalling failed: %s", err))
		return
	}

	twinsSet, err := ps.tr.RetrieveByChannel(context.TODO(), msg.Channel, 10)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Retrieving twins failed: %s", err))
		return
	}

	for _, v := range twinsSet.Twins {
		if err := ps.publish(msg, v); err != nil {
			ps.logger.Warn(fmt.Sprintf("Publishing failed: %s", err))
		}
	}
}

func (ps pubsub) publish(msg mainflux.RawMessage, twin *twins.Twin) error {
	output := mainflux.OutputSenML

	data, err := json.Marshal(msg)
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

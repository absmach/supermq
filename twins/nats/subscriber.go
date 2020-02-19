// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/twins"
	"github.com/nats-io/go-nats"
)

const (
	queue = "twins"
	input = "channel.>"
)

// Subscriber is used to intercept messages and save corresponding twin states
type Subscriber struct {
	natsClient *nats.Conn
	logger     log.Logger
	svc        twins.Service
	channelID  string
}

// NewSubscriber instances Subscriber strucure and subscribes to appropriate NATS topic
func NewSubscriber(nc *nats.Conn, chID string, svc twins.Service, logger log.Logger) *Subscriber {
	ps := Subscriber{
		natsClient: nc,
		logger:     logger,
		svc:        svc,
		channelID:  chID,
	}

	ps.natsClient.QueueSubscribe(input, queue, ps.handleMsg)

	return &ps
}

func (ps *Subscriber) handleMsg(m *nats.Msg) {
	var msg mainflux.Message
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshalling failed: %s", err))
		return
	}

	if err := ps.svc.SaveStates(&msg); err != nil {
		ps.logger.Error(fmt.Sprintf("State save failed: %s", err))
		return
	}
}

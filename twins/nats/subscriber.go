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
	queue  = "twins"
	input  = "channel.>"
	prefix = "channel"
)

type pubsub struct {
	natsClient *nats.Conn
	logger     log.Logger
	svc        twins.Service
	channelID  string
}

// Subscribe to appropriate NATS topic
func Subscribe(nc *nats.Conn, chID string, svc twins.Service, logger log.Logger) {
	ps := pubsub{
		natsClient: nc,
		logger:     logger,
		svc:        svc,
		channelID:  chID,
	}

	ps.natsClient.QueueSubscribe(input, queue, ps.handleMsg)
}

func (ps *pubsub) handleMsg(m *nats.Msg) {
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

func (ps *pubsub) Publish(twinID *string, err *error, succOp, failOp string, payload *[]byte) error {
	if ps.channelID == "" {
		return nil
	}

	op := succOp
	if *err != nil {
		op = failOp
		esb := []byte((*err).Error())
		payload = &esb
	}

	subject := fmt.Sprintf("%s.%s.%s", prefix, ps.channelID, op)

	return ps.natsClient.Publish(subject, *payload)
}

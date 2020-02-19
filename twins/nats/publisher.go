// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"fmt"

	log "github.com/mainflux/mainflux/logger"
	"github.com/nats-io/go-nats"
)

const prefix = "channel"

// Publisher is used to publish twins related notifications
type Publisher struct {
	natsClient *nats.Conn
	logger     log.Logger
	channelID  string
}

// NewPublisher instances Pubsub strucure
func NewPublisher(nc *nats.Conn, chID string, logger log.Logger) *Publisher {
	return &Publisher{
		natsClient: nc,
		logger:     logger,
		channelID:  chID,
	}
}

// Publish sends twins CRUD and state saving related operations
func (ps *Publisher) Publish(twinID *string, err *error, succOp, failOp string, payload *[]byte) error {
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

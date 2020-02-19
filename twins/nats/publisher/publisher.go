// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package publisher

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	"github.com/nats-io/go-nats"
)

const prefix = "channel"

// Publisher is used to publish twins related notifications
type Publisher struct {
	natsClient *nats.Conn
	channelID  string
}

// NewPublisher instances Pubsub strucure
func NewPublisher(nc *nats.Conn, chID string) *Publisher {
	return &Publisher{
		natsClient: nc,
		channelID:  chID,
	}
}

// Publish sends twins CRUD and state saving related operations
func (p *Publisher) Publish(twinID *string, err *error, succOp, failOp string, payload *[]byte) error {
	if p.channelID == "" {
		return nil
	}

	op := succOp
	if *err != nil {
		op = failOp
		esb := []byte((*err).Error())
		payload = &esb
	}

	pl := *payload
	if pl == nil {
		pl = []byte(fmt.Sprintf("{\"deleted\":\"%s\"}", *twinID))
	}
	subject := fmt.Sprintf("%s.%s.%s", prefix, p.channelID, op)
	mc := mainflux.Message{
		Channel:  p.channelID,
		Subtopic: op,
		Payload:  pl,
	}
	b, _ := proto.Marshal(&mc)

	return p.natsClient.Publish(subject, []byte(b))
}

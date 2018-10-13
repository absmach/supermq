//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

// Package nats contains NATS message publisher implementation.
package nats

import (
	"encoding/json"
	"fmt"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/lora"

	"github.com/nats-io/go-nats"
)

const (
	queue         = "lora-adapter"
	natsSubsTopic = "things"
)

type pubsub struct {
	nc     *nats.Conn
	svc    lora.Service
	logger logger.Logger
}

// Subscribe subscribe to Mainflux NATS broker
func Subscribe(svc lora.Service, nc *nats.Conn, log logger.Logger) error {
	ps := pubsub{
		svc:    svc,
		nc:     nc,
		logger: log,
	}

	ps.nc.QueueSubscribe(natsSubsTopic, queue, ps.handleMsg)
	ps.nc.Flush()

	if err := ps.nc.LastError(); err != nil {
		return err
	}
	return nil
}

// MfxHandler triggered when new message is received on Mainflux NATS broker
func (ps *pubsub) handleMsg(m *nats.Msg) {
	msg := lora.EventSourcing{}

	if err := json.Unmarshal(m.Data, msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Failed to unmarshal received event sourcing: %s", err))
		return
	}

	if err := ps.svc.ProvisionRouter(msg); err != nil {
		ps.logger.Error(fmt.Sprintf("Failed to provision Lora app Server: %s", err))
		return
	}
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/transformer"
	"github.com/mainflux/mainflux/transformer/senml"
	"github.com/nats-io/go-nats"
)

const (
	queue         = "transformer"
	input         = "channel.>"
	outputUnknown = "out.unknown"
	senML         = "application/senml+json"
)

type pubsub struct {
	nc     *nats.Conn
	svc    transformer.Transformer
	logger log.Logger
}

// Subscribe to appropriate NATS topic and normalizes received messages.
func Subscribe(svc transformer.Transformer, nc *nats.Conn, logger log.Logger) {
	ps := pubsub{
		nc:     nc,
		svc:    svc,
		logger: logger,
	}
	ps.nc.QueueSubscribe(input, queue, ps.handleMsg)
}

func (ps pubsub) handleMsg(m *nats.Msg) {
	var msg mainflux.Message
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshalling failed: %s", err))
		return
	}

	if err := ps.publish(msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Publishing failed: %s", err))
		return
	}
}

func (ps pubsub) publish(msg mainflux.Message) error {
	output := mainflux.OutputSenML
	t, err := ps.svc.Transform(msg)
	normalized, ok := t.([]senml.Message)
	if !ok {
		errors.New("Invalid type")
	}
	if err != nil {
		switch ct := msg.ContentType; ct {
		case senML:
			return err
		case "":
			output = outputUnknown
		default:
			output = fmt.Sprintf("out.%s", ct)
		}

		if err := ps.nc.Publish(output, msg.GetPayload()); err != nil {
			ps.logger.Warn(fmt.Sprintf("Publishing failed: %s", err))
			return err
		}
	}

	for _, v := range normalized {
		data, err := proto.Marshal(&v)
		if err != nil {
			ps.logger.Warn(fmt.Sprintf("Marshalling failed: %s", err))
			return err
		}

		if err := ps.nc.Publish(output, data); err != nil {
			ps.logger.Warn(fmt.Sprintf("Publishing failed: %s", err))
			return err
		}
	}

	return nil
}

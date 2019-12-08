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
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/paho"
	"github.com/nats-io/go-nats"
)

const (
	queue = "twins"
	input = "channel.>"
)

type pubsub struct {
	natsClient *nats.Conn
	mqttClient paho.Mqtt
	logger     log.Logger
	twins      twins.TwinRepository
}

// Subscribe to appropriate NATS topic
func Subscribe(nc *nats.Conn, mc paho.Mqtt, tr twins.TwinRepository, logger log.Logger) {
	ps := pubsub{
		natsClient: nc,
		mqttClient: mc,
		logger:     logger,
		twins:      tr,
	}
	ps.natsClient.QueueSubscribe(input, queue, ps.handleMsg)
}

func (ps pubsub) handleMsg(m *nats.Msg) {
	var err error

	var msg mainflux.Message
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshalling failed: %s", err))
		return
	}
	if msg.Channel == ps.mqttClient.Topic() {
		return
	}

	b := []byte{}
	id := ""
	defer ps.mqttClient.Publish(&id, &err, "state/success", "state/failure", &b)

	twinsSet, err := ps.twins.RetrieveByThing(context.TODO(), msg.Publisher, 1)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Retrieving twin for %s failed: %s", msg.Publisher, err))
		return
	}
	if len(twinsSet.Twins) < 1 {
		err = twins.ErrNotFound
		ps.logger.Warn(fmt.Sprintf("Retrieving twin for %s failed: %s", msg.Publisher, err))
		return
	}

	tw := twinsSet.Twins[0]

	numStates, err := ps.twins.CountStates(context.TODO(), tw)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Counting states for %s failed: %s", msg.Publisher, err))
		return
	}
	numStates++
	st := twins.State{
		TwinID:     tw.ID,
		ID:         numStates,
		Definition: tw.Definitions[len(tw.Definitions)-1].ID,
		Created:    time.Now(),
		Payload:    msg.Payload,
	}
	if err := ps.twins.SaveState(context.TODO(), st); err != nil {
		ps.logger.Warn(fmt.Sprintf("Updating state for %s failed: %s", msg.Publisher, err))
		return
	}

	id = msg.Publisher
	b = msg.Payload

	ps.logger.Info(fmt.Sprintf("Updating state for %s succeeded", msg.Publisher))
}

func (ps pubsub) publish(msg mainflux.Message, twin *twins.Twin) error {
	data, err := json.Marshal(msg)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Marshalling failed: %s", err))
		return err
	}

	subject := fmt.Sprintf("%s.%s", msg.Channel, msg.Subtopic)
	if err := ps.natsClient.Publish(subject, data); err != nil {
		ps.logger.Warn(fmt.Sprintf("Publishing failed: %s", err))
		return err
	}

	return nil
}

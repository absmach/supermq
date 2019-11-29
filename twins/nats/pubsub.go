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
	b := []byte{}
	id := ""
	var err error
	var msg mainflux.Message
	defer func() {
		if msg.Channel != ps.mqttClient.Topic() {
			ps.mqttClient.Publish(&id, &err, "state/success", "state/failure", &b)
		}
	}()

	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshalling failed: %s", err))
		return
	}

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
	df := tw.Definitions[len(tw.Definitions)-1].Revision
	sr := 0
	if len(tw.States) > 0 {
		sr = tw.States[len(tw.States)-1].Serial + 1
	}
	state := twins.State{
		Definition: df,
		Serial:     sr,
		Created:    time.Now(),
		Payload:    msg.Payload,
	}
	tw.States = append(tw.States, state)

	if err := ps.twins.Update(context.TODO(), tw); err != nil {
		ps.logger.Warn(fmt.Sprintf("Updating twin for %s failed: %s", msg.Publisher, err))
	}

	b, err = json.Marshal(state)

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

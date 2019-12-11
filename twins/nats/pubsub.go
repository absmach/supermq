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
	states     twins.StateRepository
}

// Subscribe to appropriate NATS topic
func Subscribe(nc *nats.Conn, mc paho.Mqtt, tr twins.TwinRepository, sr twins.StateRepository, logger log.Logger) {
	ps := pubsub{
		natsClient: nc,
		mqttClient: mc,
		logger:     logger,
		twins:      tr,
		states:     sr,
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

	tw, err := ps.twins.RetrieveByThing(context.TODO(), msg.Publisher)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Retrieving twin for %s failed: %s", msg.Publisher, err))
		return
	}

	var recs []Record
	if err := json.Unmarshal(msg.Payload, &recs); err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshal payload for %s failed: %s", msg.Publisher, err))
		return
	}

	st, err := ps.states.RetrieveLast(context.TODO(), tw.ID)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Retrieve last state for %s failed: %s", msg.Publisher, err))
		return
	}

	if save := prepareState(&st, &tw, recs, msg); !save {
		return
	}

	if err := ps.states.Save(context.TODO(), st); err != nil {
		ps.logger.Warn(fmt.Sprintf("Updating state for %s failed: %s", msg.Publisher, err))
		return
	}

	id = msg.Publisher
	b = msg.Payload

	ps.logger.Info(fmt.Sprintf("Updating state for %s succeeded", msg.Publisher))
}

func prepareState(st *twins.State, tw *twins.Twin, recs []Record, msg mainflux.Message) bool {
	def := tw.Definitions[len(tw.Definitions)-1]
	st.TwinID = tw.ID
	st.ID++
	st.Created = time.Now()
	st.Definition = def.ID
	if st.Payload == nil {
		st.Payload = make(map[string]interface{})
	}

	save := false

	rec := recs[0]
	for k, a := range def.Attributes {
		if !a.PeristState {
			continue
		}
		if a.ChannelID == msg.Channel && a.Subtopic == msg.Subtopic {
			st.Payload[k] = rec.Value
			save = true
			break
		}
	}

	return save
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

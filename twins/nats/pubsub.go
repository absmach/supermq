// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/mainflux/mainflux/twins/mqtt"
	"github.com/mainflux/senml"
	"github.com/nats-io/go-nats"
)

const (
	queue = "twins"
	input = "channel.>"
)

var mqttOp = map[string]string{
	"stateSucc": "state/success",
	"stateFail": "state/failure",
}

type pubsub struct {
	natsClient *nats.Conn
	mqttClient mqtt.Mqtt
	logger     log.Logger
	twins      twins.TwinRepository
	states     twins.StateRepository
}

// Subscribe to appropriate NATS topic
func Subscribe(nc *nats.Conn, mc mqtt.Mqtt, tr twins.TwinRepository, sr twins.StateRepository, logger log.Logger) {
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
	var msg mainflux.Message
	err := proto.Unmarshal(m.Data, &msg)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Unmarshalling failed: %s", err))
		return
	}
	if msg.Channel == ps.mqttClient.Topic() {
		return
	}

	var b []byte
	var id string
	defer ps.mqttClient.Publish(&id, &err, mqttOp["stateSucc"], mqttOp["stateFail"], &b)

	tw, err := ps.twins.RetrieveByThing(context.TODO(), msg.Publisher)
	if err != nil {
		ps.logger.Warn(fmt.Sprintf("Retrieving twin for %s failed: %s", msg.Publisher, err))
		return
	}

	var recs []senml.Record
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
		ps.logger.Info(fmt.Sprintf("No persistent attributes for %s for %s", msg.Subtopic, msg.Publisher))
		return
	}

	if err := ps.states.Save(context.TODO(), st); err != nil {
		ps.logger.Warn(fmt.Sprintf("Updating state for %s failed: %s", msg.Publisher, err))
		return
	}

	id = msg.Publisher
	b = msg.Payload

	ps.logger.Info(fmt.Sprintf("Updating attribute %s for %s succeeded", msg.Subtopic, msg.Publisher))
}

func prepareState(st *twins.State, tw *twins.Twin, recs []senml.Record, msg mainflux.Message) bool {
	def := tw.Definitions[len(tw.Definitions)-1]
	st.TwinID = tw.ID
	st.ID++
	st.Created = time.Now()
	st.Definition = def.ID
	if st.Payload == nil {
		st.Payload = make(map[string]interface{})
	}

	save := false
	for k, a := range def.Attributes {
		if !a.PersistState {
			continue
		}
		if a.Channel == msg.Channel && a.Subtopic == msg.Subtopic {
			st.Payload[k] = recs[0].Value
			save = true
			break
		}
	}

	return save
}

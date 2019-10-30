// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"github.com/gocql/gocql"
	"github.com/mainflux/mainflux/transformer/senml"
	"github.com/mainflux/mainflux/writers"
)

var _ writers.MessageRepository = (*cassandraRepository)(nil)

type cassandraRepository struct {
	session *gocql.Session
}

// New instantiates Cassandra message repository.
func New(session *gocql.Session) writers.MessageRepository {
	return &cassandraRepository{session}
}

func (cr *cassandraRepository) Save(messages ...senml.Message) error {
	cql := `INSERT INTO messages (id, channel, subtopic, publisher, protocol,
			name, unit, value, string_value, bool_value, data_value, value_sum,
			time, update_time, link)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	id := gocql.TimeUUID()

	for _, msg := range messages {
		var floatVal, valSum *float64
		var strVal, dataVal *string
		var boolVal *bool
		switch msg.Value.(type) {
		case *senml.Message_FloatValue:
			v := msg.GetFloatValue()
			floatVal = &v
		case *senml.Message_StringValue:
			v := msg.GetStringValue()
			strVal = &v
		case *senml.Message_DataValue:
			v := msg.GetDataValue()
			dataVal = &v
		case *senml.Message_BoolValue:
			v := msg.GetBoolValue()
			boolVal = &v
		}

		if msg.GetValueSum() != nil {
			v := msg.GetValueSum().GetValue()
			valSum = &v
		}

		err := cr.session.Query(cql, id, msg.GetChannel(), msg.GetSubtopic(), msg.GetPublisher(),
			msg.GetProtocol(), msg.GetName(), msg.GetUnit(), floatVal,
			strVal, boolVal, dataVal, valSum, msg.GetTime(), msg.GetUpdateTime(), msg.GetLink()).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

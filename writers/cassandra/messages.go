// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gocql/gocql"
	"github.com/mainflux/mainflux/pkg/errors"
	mfjson "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

var (
	errSaveMessage = errors.New("failed to save message to cassandra database")
	errNoTable     = errors.New("table does not exist")
)
var _ writers.MessageRepository = (*cassandraRepository)(nil)

type cassandraRepository struct {
	session *gocql.Session
}

// New instantiates Cassandra message repository.
func New(session *gocql.Session) writers.MessageRepository {
	return &cassandraRepository{session}
}

func (cr *cassandraRepository) Save(message interface{}) error {
	switch m := message.(type) {
	case mfjson.Message:
		return cr.saveJSON(m)
	default:
		return cr.saveSenml(m)
	}
}

func (cr *cassandraRepository) saveSenml(messages interface{}) error {
	msgs, ok := messages.([]senml.Message)
	if !ok {
		return errSaveMessage
	}
	cql := `INSERT INTO messages (id, channel, subtopic, publisher, protocol,
            name, unit, value, string_value, bool_value, data_value, sum,
            time, update_time)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	id := gocql.TimeUUID()

	for _, msg := range msgs {
		err := cr.session.Query(cql, id, msg.Channel, msg.Subtopic, msg.Publisher,
			msg.Protocol, msg.Name, msg.Unit, msg.Value, msg.StringValue,
			msg.BoolValue, msg.DataValue, msg.Sum, msg.Time, msg.UpdateTime).Exec()
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}
	}

	return nil
}

func (cr *cassandraRepository) saveJSON(msg mfjson.Message) error {
	table := strings.Split(msg.Subtopic, ".")[0]
	if err := cr.insertJSON(table, msg); err != nil {
		if err == errNoTable {
			if err := cr.creteTable(table); err != nil {
				return err
			}
			return cr.insertJSON(table, msg)
		}
		return err
	}
	return nil
}

func (cr *cassandraRepository) insertJSON(table string, msg mfjson.Message) error {
	pld, err := json.Marshal(msg.Payload)
	if err != nil {
		return err
	}
	cql := `INSERT INTO %s (id, channel, created, subtopic, publisher, payload)
            VALUES (?, ?, ?, ?, ?, ?)`
	cql = fmt.Sprintf(cql, table)
	id := gocql.TimeUUID()

	err = cr.session.Query(cql, id, msg.Channel, msg.Created, msg.Subtopic, msg.Publisher, string(pld)).Exec()
	if err != nil {
		if err.Error() == fmt.Sprintf("unconfigured table %s", table) {
			return errNoTable
		}
		return errors.Wrap(errSaveMessage, err)
	}

	return nil
}

func (cr *cassandraRepository) creteTable(name string) error {
	q := `CREATE TABLE IF NOT EXISTS %s (
        id uuid,
        channel text,
        subtopic text,
        publisher text,
        protocol text,
        created bigint,
        payload text,
        PRIMARY KEY (channel, created, id)
    ) WITH CLUSTERING ORDER BY (created DESC)`

	q = fmt.Sprintf(q, name)
	return cr.session.Query(q).Exec()
}

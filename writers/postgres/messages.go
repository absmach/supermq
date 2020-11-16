// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq" // required for DB access
	"github.com/mainflux/mainflux/pkg/errors"
	mfjson "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

const (
	errInvalid        = "invalid_text_representation"
	errUndefinedTable = "undefined_table"
)

var (
	// ErrInvalidMessage indicates that service received message that
	// doesn't fit required format.
	ErrInvalidMessage = errors.New("invalid message representation")
	errSaveMessage    = errors.New("failed to save message to postgres database")
	errTransRollback  = errors.New("failed to rollback transaction")
	errMessageFormat  = errors.New("invalid message format")
	errNoMessages     = errors.New("empty message")
	errNoTable        = errors.New("relation does not exist")
)

var _ writers.MessageRepository = (*postgresRepo)(nil)

type postgresRepo struct {
	db *sqlx.DB
}

// New returns new PostgreSQL writer.
func New(db *sqlx.DB) writers.MessageRepository {
	return &postgresRepo{db: db}
}

func (pr postgresRepo) Save(message interface{}) (err error) {
	switch m := message.(type) {
	case []mfjson.Message:
		return pr.saveJSON(m)
	default:
		return pr.saveSenml(m)
	}
}

func (pr postgresRepo) saveSenml(messages interface{}) error {
	msgs, ok := messages.([]senml.Message)
	if !ok {
		return errSaveMessage
	}
	q := `INSERT INTO senml (id, channel, subtopic, publisher, protocol,
          name, unit, value, string_value, bool_value, data_value, sum,
          time, update_time)
          VALUES (:id, :channel, :subtopic, :publisher, :protocol, :name, :unit,
          :value, :string_value, :bool_value, :data_value, :sum,
          :time, :update_time);`

	tx, err := pr.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = errors.Wrap(err, errors.Wrap(errTransRollback, txErr))
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(errSaveMessage, err)
		}
		return
	}()

	for _, msg := range msgs {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		m := senmlMessage{Message: msg, ID: id.String()}
		if _, err := tx.NamedExec(q, m); err != nil {
			pqErr, ok := err.(*pq.Error)
			if ok {
				switch pqErr.Code.Name() {
				case errInvalid:
					return errors.Wrap(errSaveMessage, ErrInvalidMessage)
				}
			}

			return errors.Wrap(errSaveMessage, err)
		}
	}
	return err
}

func (pr postgresRepo) saveJSON(msgs []mfjson.Message) error {
	if len(msgs) == 0 {
		return nil
	}
	table := strings.Split(msgs[0].Subtopic, ".")[0]
	if err := pr.insertJSON(table, msgs); err != nil {
		if err == errNoTable {
			if err := pr.createTable(table); err != nil {
				return err
			}
			return pr.insertJSON(table, msgs)
		}
		return err
	}
	return nil
}

func (pr postgresRepo) insertJSON(table string, msgs []mfjson.Message) error {
	tx, err := pr.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	defer func() {
		if err != nil {
			if txErr := tx.Rollback(); txErr != nil {
				err = errors.Wrap(err, errors.Wrap(errTransRollback, txErr))
			}
			return
		}

		if err = tx.Commit(); err != nil {
			err = errors.Wrap(errSaveMessage, err)
		}
		return
	}()

	q := `INSERT INTO %s (id, channel, created, subtopic, publisher, protocol, payload)
          VALUES (:id, :channel, :created, :subtopic, :publisher, :protocol, :payload);`
	q = fmt.Sprintf(q, table)

	for _, m := range msgs {
		var dbmsg jsonMessage
		dbmsg, err = toJSONMessage(m)
		if err != nil {
			return errors.Wrap(errSaveMessage, err)
		}
		if _, err = tx.NamedExec(q, dbmsg); err != nil {
			pqErr, ok := err.(*pq.Error)
			if ok {
				switch pqErr.Code.Name() {
				case errInvalid:
					return errors.Wrap(errSaveMessage, ErrInvalidMessage)
				case errUndefinedTable:
					return errNoTable
				}
			}
			return err
		}
	}
	return nil
}

func (pr postgresRepo) createTable(name string) error {
	q := `CREATE TABLE IF NOT EXISTS %s (
                        id            UUID,
                        created       BIGINT,
                        channel       UUID,
                        subtopic      VARCHAR(254),
                        publisher     UUID,
                        protocol      TEXT,
                        payload       JSONB,
                        PRIMARY KEY (id)
                    )`
	q = fmt.Sprintf(q, name)

	_, err := pr.db.Exec(q)
	return err
}

type senmlMessage struct {
	senml.Message
	ID string `db:"id"`
}

type jsonMessage struct {
	ID        string `db:"id"`
	Channel   string `db:"channel"`
	Created   int64  `db:"created"`
	Subtopic  string `db:"subtopic"`
	Publisher string `db:"publisher"`
	Protocol  string `db:"protocol"`
	Payload   []byte `db:"payload"`
}

func toJSONMessage(msg mfjson.Message) (jsonMessage, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return jsonMessage{}, err
	}

	data := []byte("{}")
	if msg.Payload != nil {
		b, err := json.Marshal(msg.Payload)
		if err != nil {
			return jsonMessage{}, errors.Wrap(errSaveMessage, err)
		}
		data = b
	}

	m := jsonMessage{
		ID:        id.String(),
		Channel:   msg.Channel,
		Created:   msg.Created,
		Subtopic:  msg.Subtopic,
		Publisher: msg.Publisher,
		Protocol:  msg.Protocol,
		Payload:   data,
	}

	return m, nil
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/gofrs/uuid"

	"github.com/jmoiron/sqlx" // required for DB access
	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

// ErrInvalidMessage indicates that service received message that
// doesn't fit required format.
var ErrInvalidMessage = errors.New("invalid message representation")

var _ writers.MessageRepository = (*clickhouseRepo)(nil)

type clickhouseRepo struct {
	db *sqlx.DB
}

// New returns new Clickhouse writer.
func New(db *sqlx.DB) writers.MessageRepository {
	return &clickhouseRepo{db: db}
}

func (cr clickhouseRepo) Save(messages ...senml.Message) error {
	q := `INSERT INTO messages (id, channel, subtopic, publisher, protocol,
    name, unit, value, string_value, bool_value, data_value, sum,
    time, update_time)
    VALUES (:id, :channel, :subtopic, :publisher, :protocol, :name, :unit,
    :value, :string_value, :bool_value, :data_value, :sum,
    :time, :update_time);`

	tx, err := cr.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareNamed(q)
	if err != nil {
		return err
	}

	for _, msg := range messages {
		dbth, err := toDBMessage(msg)
		if err != nil {
			return err
		}

		if _, err := stmt.Exec(dbth); err != nil {
			return err
		}
	}

	return tx.Commit()
}

type dbMessage struct {
	ID          string    `db:"id"`
	Channel     string    `db:"channel"`
	Subtopic    string    `db:"subtopic"`
	Publisher   string    `db:"publisher"`
	Protocol    string    `db:"protocol"`
	Name        string    `db:"name"`
	Unit        string    `db:"unit"`
	Value       *float64  `db:"value"`
	StringValue *string   `db:"string_value"`
	BoolValue   *bool     `db:"bool_value"`
	DataValue   *string   `db:"data_value"`
	Sum         *float64  `db:"sum"`
	Time        time.Time `db:"time"`
	UpdateTime  float64   `db:"update_time"`
}

func toDBMessage(msg senml.Message) (dbMessage, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return dbMessage{}, err
	}

	m := dbMessage{
		ID:         id.String(),
		Channel:    msg.Channel,
		Subtopic:   msg.Subtopic,
		Publisher:  msg.Publisher,
		Protocol:   msg.Protocol,
		Name:       msg.Name,
		Unit:       msg.Unit,
		Time:       floatToTime(msg.Time),
		UpdateTime: msg.UpdateTime,
		Sum:        msg.Sum,
	}

	switch {
	case msg.Value != nil:
		m.Value = msg.Value
	case msg.StringValue != nil:
		m.StringValue = msg.StringValue
	case msg.DataValue != nil:
		m.DataValue = msg.DataValue
	case msg.BoolValue != nil:
		m.BoolValue = msg.BoolValue
	}

	return m, nil
}

func floatToTime(ft float64) time.Time {
	sec, nsec := math.Modf(ft)
	return time.Unix(int64(sec), int64(nsec*1e9))
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx" // required for DB access
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/transformers/senml"
)

var errReadMessages = errors.New("faled to read messages from clickhouse database")

var _ readers.MessageRepository = (*clickhouseRepository)(nil)

type clickhouseRepository struct {
	db *sqlx.DB
}

// New returns new Clickhouse writer.
func New(db *sqlx.DB) readers.MessageRepository {
	return &clickhouseRepository{
		db: db,
	}
}

func (cr clickhouseRepository) ReadAll(chanID string, offset, limit uint64, query map[string]string) (readers.MessagesPage, error) {
	q := fmt.Sprintf(`SELECT * FROM messages
    WHERE %s ORDER BY time DESC
    LIMIT :offset, :limit;`, fmtCondition(query))

	params := map[string]interface{}{
		"channel":   chanID,
		"limit":     limit,
		"offset":    offset,
		"subtopic":  query["subtopic"],
		"publisher": query["publisher"],
		"name":      query["name"],
		"protocol":  query["protocol"],
	}

	rows, err := cr.db.NamedQuery(q, params)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	defer rows.Close()

	page := readers.MessagesPage{
		Offset:   offset,
		Limit:    limit,
		Messages: []senml.Message{},
	}
	for rows.Next() {
		dbm := dbMessage{Channel: chanID}
		if err := rows.StructScan(&dbm); err != nil {
			return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
		}

		msg := toMessage(dbm)
		page.Messages = append(page.Messages, msg)
	}

	q = `SELECT COUNT(*) FROM messages WHERE channel = ?`
	qParams := []interface{}{chanID}

	if query["subtopic"] != "" {
		q = `SELECT COUNT(*) FROM messages WHERE channel = ? AND subtopic = ?`
		qParams = append(qParams, query["subtopic"])
	}

	if err := cr.db.QueryRow(q, qParams...).Scan(&page.Total); err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}

	return page, nil
}

func fmtCondition(query map[string]string) string {
	condition := `channel = :channel`
	for name := range query {
		switch name {
		case
			"subtopic",
			"publisher",
			"name",
			"protocol":
			condition = fmt.Sprintf(`%s AND %s = :%s`, condition, name, name)
		}
	}
	return condition
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

func toMessage(dbm dbMessage) senml.Message {
	msg := senml.Message{
		Channel:    dbm.Channel,
		Subtopic:   dbm.Subtopic,
		Publisher:  dbm.Publisher,
		Protocol:   dbm.Protocol,
		Name:       dbm.Name,
		Unit:       dbm.Unit,
		Time:       timeToFloat(dbm.Time),
		UpdateTime: dbm.UpdateTime,
		Sum:        dbm.Sum,
	}

	switch {
	case dbm.Value != nil:
		msg.Value = dbm.Value
	case dbm.StringValue != nil:
		msg.StringValue = dbm.StringValue
	case dbm.DataValue != nil:
		msg.DataValue = dbm.DataValue
	case dbm.BoolValue != nil:
		msg.BoolValue = dbm.BoolValue
	}

	return msg
}

func timeToFloat(t time.Time) float64 {
	return float64(t.UnixNano()) / 1e9
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx" // required for DB access
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
)

const errInvalid = "invalid_text_representation"

var errReadMessages = errors.New("failed to read messages from postgres database")

var _ readers.MessageRepository = (*postgresRepository)(nil)

type postgresRepository struct {
	db *sqlx.DB
}

// New returns new PostgreSQL writer.
func New(db *sqlx.DB) readers.MessageRepository {
	return &postgresRepository{
		db: db,
	}
}

func (tr postgresRepository) ReadAll(chanID string, offset, limit uint64, query map[string]string) (readers.MessagesPage, error) {
	q := fmt.Sprintf(`SELECT * FROM messages
    WHERE %s ORDER BY time DESC
    LIMIT :limit OFFSET :offset;`, fmtCondition(chanID, query))

	params := map[string]interface{}{
		"channel":   chanID,
		"limit":     limit,
		"offset":    offset,
		"subtopic":  query["subtopic"],
		"publisher": query["publisher"],
		"name":      query["name"],
		"protocol":  query["protocol"],
	}

	rows, err := tr.db.NamedQuery(q, params)
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
		msg := dbMessage{Message: senml.Message{Channel: chanID}}
		if err := rows.StructScan(&msg); err != nil {
			return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
		}

		page.Messages = append(page.Messages, msg.Message)
	}

	q = `SELECT COUNT(*) FROM messages WHERE channel = $1;`
	qParams := []interface{}{chanID}

	if query["subtopic"] != "" {
		q = `SELECT COUNT(*) FROM messages WHERE channel = $1 AND subtopic = $2;`
		qParams = append(qParams, query["subtopic"])
	}

	if err := tr.db.QueryRow(q, qParams...).Scan(&page.Total); err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}

	return page, nil
}

func fmtCondition(chanID string, query map[string]string) string {
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
	ID string `db:"id"`
	senml.Message
}

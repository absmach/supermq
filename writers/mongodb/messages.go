// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mongodb

import (
	"context"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

const (
	senmlCollection string = "senml"
	jsonCollection  string = "json"
)

var (
	errSaveMessage   = errors.New("failed to save message to mongodb database")
	errMessageFormat = errors.New("invalid message format")
)

var _ writers.MessageRepository = (*mongoRepo)(nil)

type mongoRepo struct {
	db *mongo.Database
}

// New returns new MongoDB writer.
func New(db *mongo.Database) writers.MessageRepository {
	return &mongoRepo{db}
}

func (repo *mongoRepo) Save(message interface{}) error {
	switch m := message.(type) {
	case json.Message:
		return repo.saveJSON(m)
	default:
		return repo.saveSenml(m)
	}

}

func (repo *mongoRepo) saveSenml(messages interface{}) error {
	msgs, ok := messages.([]senml.Message)
	if !ok {
		return errSaveMessage
	}
	coll := repo.db.Collection(senmlCollection)
	var dbMsgs []interface{}
	for _, msg := range msgs {
		dbMsgs = append(dbMsgs, msg)
	}

	_, err := coll.InsertMany(context.Background(), dbMsgs)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	return nil
}

func (repo *mongoRepo) saveJSON(message json.Message) error {
	msgs := []interface{}{}
	switch pld := message.Payload.(type) {
	case map[string]interface{}:
		msgs = append(msgs, message)
	case []map[string]interface{}:
		for _, p := range pld {
			add := message
			add.Payload = p
			msgs = append(msgs, add)
		}
	}

	coll := repo.db.Collection(strings.Split(msgs[0].(json.Message).Subtopic, ".")[0])

	_, err := coll.InsertMany(context.Background(), msgs)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	return nil
}

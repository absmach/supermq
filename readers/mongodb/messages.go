// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mongodb

import (
	"context"
	"encoding/json"

	"github.com/mainflux/mainflux/pkg/errors"
	jsont "github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	format = "format"
	// Collection for SenML messages
	defCollection = "messages"
)

var errReadMessages = errors.New("failed to read messages from mongodb database")

var _ readers.MessageRepository = (*mongoRepository)(nil)

type mongoRepository struct {
	db *mongo.Database
}

// New returns new MongoDB reader.
func New(db *mongo.Database) readers.MessageRepository {
	return mongoRepository{
		db: db,
	}
}

func (repo mongoRepository) ReadAll(chanID string, rpm readers.PageMetadata) (readers.MessagesPage, error) {
	if rpm.Format == "" {
		rpm.Format = defCollection
	}

	col := repo.db.Collection(rpm.Format)

	sortMap := map[string]interface{}{
		"time": -1,
	}
	// Remove format filter and format the rest properly.
	filter := fmtCondition(chanID, rpm)
	cursor, err := col.Find(context.Background(), filter, options.Find().SetSort(sortMap).SetLimit(int64(rpm.Limit)).SetSkip(int64(rpm.Offset)))
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	defer cursor.Close(context.Background())

	var messages []readers.Message
	switch rpm.Format {
	case defCollection:
		for cursor.Next(context.Background()) {
			var m senml.Message
			if err := cursor.Decode(&m); err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}

			messages = append(messages, m)
		}
	default:
		for cursor.Next(context.Background()) {
			var m map[string]interface{}
			if err := cursor.Decode(&m); err != nil {
				return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
			}
			m["payload"] = jsont.ParseFlat(m["payload"])

			messages = append(messages, m)
		}
	}

	total, err := col.CountDocuments(context.Background(), filter)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	if total < 0 {
		return readers.MessagesPage{}, nil
	}

	mp := readers.MessagesPage{
		PageMetadata: rpm,
		Total:        uint64(total),
		Messages:     messages,
	}

	return mp, nil
}

func fmtCondition(chanID string, rpm readers.PageMetadata) bson.D {
	filter := bson.D{
		bson.E{
			Key:   "channel",
			Value: chanID,
		},
	}

	var query map[string]interface{}
	meta, err := json.Marshal(rpm)
	if err != nil {
		return filter
	}
	json.Unmarshal(meta, &query)

	for name, value := range query {
		switch name {
		case
			"channel",
			"subtopic",
			"publisher",
			"name",
			"protocol":
			filter = append(filter, bson.E{Key: name, Value: value})
		case "v":
			filter = append(filter, bson.E{Key: "value", Value: value})
		case "vb":
			filter = append(filter, bson.E{Key: "bool_value", Value: value})
		case "vs":
			filter = append(filter, bson.E{Key: "string_value", Value: value})
		case "vd":
			filter = append(filter, bson.E{Key: "data_value", Value: value})
		case "from":
			filter = append(filter, bson.E{Key: "time", Value: bson.M{"$gte": value}})
		case "to":
			filter = append(filter, bson.E{Key: "time", Value: bson.M{"$lt": value}})
		}
	}

	return filter
}

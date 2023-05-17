// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mongodb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/readers"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection for SenML messages.
const defCollection = "messages"

// mongoDB comparison operators different than default mainflux
// readers comparison operators. This map maps mainflux comparators
// to mongoDB comparators.
var mongoComparators = map[string]string{
	readers.EqualKey:            "eq",
	readers.LowerThanKey:        "lt",
	readers.LowerThanEqualKey:   "lte",
	readers.GreaterThanKey:      "gt",
	readers.GreaterThanEqualKey: "gte",
}

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
	format := defCollection
	order := "time"
	if rpm.Format != "" && rpm.Format != defCollection {
		order = "created"
		format = rpm.Format
	}

	col := repo.db.Collection(format)

	sortMap := map[string]interface{}{
		order: -1,
	}
	// Remove format filter and format the rest properly.
	filter := fmtCondition(chanID, rpm)
	fmt.Println("Filter:-")
	fmt.Println("###")
	fmt.Println(filter)
	fmt.Println()
	fmt.Println()
	fmt.Println()
	fmt.Println("###")
	cursor, err := col.Find(context.Background(), filter, options.Find().SetSort(sortMap).SetLimit(int64(rpm.Limit)).SetSkip(int64(rpm.Offset)))
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(readers.ErrReadMessages, err)
	}
	defer cursor.Close(context.Background())

	var messages []readers.Message
	switch format {
	case defCollection:
		for cursor.Next(context.Background()) {
			var m senml.Message
			if err := cursor.Decode(&m); err != nil {
				return readers.MessagesPage{}, errors.Wrap(readers.ErrReadMessages, err)
			}

			messages = append(messages, m)
		}
	default:
		for cursor.Next(context.Background()) {
			var m map[string]interface{}
			if err := cursor.Decode(&m); err != nil {
				return readers.MessagesPage{}, errors.Wrap(readers.ErrReadMessages, err)
			}

			messages = append(messages, m)
		}
	}

	total, err := col.CountDocuments(context.Background(), filter)
	if err != nil {
		return readers.MessagesPage{}, errors.Wrap(readers.ErrReadMessages, err)
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
	if err := json.Unmarshal(meta, &query); err != nil {
		return filter
	}

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
			bsonFilter := value
			if val, ok := query["comparator"].(string); ok {
				comparator := fmt.Sprintf("$%s", mongoComparators[val]) // mongoDB comparison operator
				bsonFilter = bson.M{comparator: value}
			}
			filter = append(filter, bson.E{Key: "value", Value: bsonFilter})
		case "vb":
			filter = append(filter, bson.E{Key: "bool_value", Value: value})
		case "vs":
			bsonFilter := value
			stringValue := "string_value"
			if val, ok := query["comparator"].(string); ok {
				fmt.Println("val = ", val)
				comparator := fmt.Sprintf("$%s", mongoComparators[val])
				fmt.Println("comparator = ", comparator)
				switch comparator {
				case "$eq":
					// comparator = fmt.Sprintf("$%s", mongoComparators[val]) // mongoDB comparison operator
					bsonFilter = bson.M{comparator: value}
				case "$gt":
					notBsonFilter := bson.M{"$eq": value}
					filter = append(filter, bson.E{Key: "string_value", Value: bson.M{"$not": notBsonFilter}})
					value = fmt.Sprintf("/%s/", value)
					comparator = "$regex"
					bsonFilter = value
				case "$gte":
					value = fmt.Sprintf("/%s/", value)
					bsonFilter = value
				case "$lte":
					notBsonFilter := bson.M{"$eq": value}
					filter = append(filter, bson.E{Key: "string_value", Value: bson.M{"$not": notBsonFilter}})
					stringValue = fmt.Sprintf("/%s/", stringValue)
					bsonFilter = value
				case "$lt":
					stringValue = fmt.Sprintf("/%s/", stringValue)
					bsonFilter = value
				}
			}
			filter = append(filter, bson.E{Key: stringValue, Value: bsonFilter})
		case "vd":
			bsonFilter := value
			if val, ok := query["comparator"].(string); ok {
				comparator := fmt.Sprintf("$%s", mongoComparators[val])
				bsonFilter = bson.M{comparator: value}
			}
			filter = append(filter, bson.E{Key: "data_value", Value: bsonFilter})
		case "from":
			filter = append(filter, bson.E{Key: "time", Value: bson.M{"$gte": value}})
		case "to":
			filter = append(filter, bson.E{Key: "time", Value: bson.M{"$lt": value}})
		}
	}

	return filter
}

// string_value = value/field in the mongo database , value = vs -> from query

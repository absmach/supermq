// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mongodb

import (
	"context"

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

var errSaveMessage = errors.New("failed to save message to mongodb database")

var _ writers.MessageRepository = (*mongoRepo)(nil)

type mongoRepo struct {
	db *mongo.Database
}

// New returns new MongoDB writer.
func New(db *mongo.Database) writers.MessageRepository {
	return &mongoRepo{db}
}

func (repo *mongoRepo) Save(messages interface{}) error {

	switch messages.(type) {
	case json.Message:
		return repo.saveJSON(messages)
	default:
		return repo.saveSenml(messages)
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
		m := message{
			Channel:    msg.Channel,
			Subtopic:   msg.Subtopic,
			Publisher:  msg.Publisher,
			Protocol:   msg.Protocol,
			Name:       msg.Name,
			Unit:       msg.Unit,
			Time:       msg.Time,
			UpdateTime: msg.UpdateTime,
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
		m.Sum = msg.Sum

		dbMsgs = append(dbMsgs, m)
	}

	_, err := coll.InsertMany(context.Background(), dbMsgs)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	return nil
}

func (repo *mongoRepo) saveJSON(messages interface{}) error {
	msg, ok := messages.(json.Message)
	if !ok {
		return errSaveMessage
	}
	coll := repo.db.Collection(jsonCollection)
	m := message{
		Channel:   msg.Channel,
		Subtopic:  msg.Subtopic,
		Publisher: msg.Publisher,
		Protocol:  msg.Protocol,
		Payload:   msg.Payload,
	}
	_, err := coll.InsertOne(context.Background(), m)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	return nil
}

type message struct {
	Channel     string                 `bson:"channel,omitempty"`
	Subtopic    string                 `bson:"subtopic,omitempty"`
	Publisher   string                 `bson:"publisher,omitempty"`
	Protocol    string                 `bson:"protocol,omitempty"`
	Name        string                 `bson:"name,omitempty"`
	Unit        string                 `bson:"unit,omitempty"`
	Value       *float64               `bson:"value,omitempty"`
	StringValue *string                `bson:"stringValue,omitempty"`
	BoolValue   *bool                  `bson:"boolValue,omitempty"`
	DataValue   *string                `bson:"dataValue,omitempty"`
	Sum         *float64               `bson:"sum,omitempty"`
	Time        float64                `bson:"time,omitempty"`
	UpdateTime  float64                `bson:"updateTime,omitempty"`
	Payload     map[string]interface{} `bson:"payload,omitempty"`
}

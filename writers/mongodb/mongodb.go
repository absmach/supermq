package mongodb

import (
	"context"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/writers"
)

const (
	collectionName string = "mainflux"
)

var _ writers.MessageRepository = (*mongoRepo)(nil)

type mongoRepo struct {
	db *mongo.Database
}

// New returns new MongoDB writer.
func New(db *mongo.Database) (writers.MessageRepository, error) {
	return &mongoRepo{db}, nil
}

func (repo *mongoRepo) Save(msg mainflux.Message) error {
	coll := repo.db.Collection(collectionName)

	_, err := coll.InsertOne(
		context.Background(),
		bson.NewDocument(
			bson.EC.Int64("channel", int64(msg.Channel)),
			bson.EC.Int64("publisher", int64(msg.Publisher)),
			bson.EC.String("protocol", msg.Protocol),
			bson.EC.String("unit", msg.Unit),
			bson.EC.Double("value", msg.Value),
			bson.EC.String("stringValue", msg.StringValue),
			bson.EC.Boolean("boolValue", msg.BoolValue),
			bson.EC.String("dataValue", msg.DataValue),
			bson.EC.Double("valueSum", msg.ValueSum),
			bson.EC.Double("time", msg.Time),
			bson.EC.Double("updateTime", msg.UpdateTime),
			bson.EC.String("link", msg.Link),
		))

	return err
}

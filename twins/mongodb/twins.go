//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package mongodb

import (
	"context"

	"github.com/mainflux/mainflux/twins"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const collectionName string = "mainflux"

type twinRepository struct {
	db *mongo.Database
}

var _ twins.TwinRepository = (*twinRepository)(nil)

// NewTwinRepository instantiates a MongoDB implementation of twin
// repository.
func NewTwinRepository(db *mongo.Database) twins.TwinRepository {
	return &twinRepository{
		db: db,
	}
}

// Save persists the twin. Successful operation is indicated by a nil
// error response.
func (tr *twinRepository) Save(ctx context.Context, tw twins.Twin) error {
	coll := tr.db.Collection(collectionName)

	if _, err := tr.RetrieveByKey(ctx, tw.Key); err != nil {
		return twins.ErrConflict
	}

	dbtw := toDBTwin(tw)
	if _, err := coll.InsertOne(context.Background(), dbtw); err != nil {
		return err
	}

	return nil
}

// Update performs an update to the existing twins. A non-nil error is
// returned to indicate operation failure.
func (tr *twinRepository) Update(_ context.Context, tw twins.Twin) error {
	coll := tr.db.Collection(collectionName)

	filter := bson.D{{"id", tw.ID}}
	dbtw := toDBTwin(tw)
	if _, err := coll.UpdateOne(context.Background(), filter, dbtw); err != nil {
		return err
	}

	return nil
}

// RetrieveByID retrieves the twin having the provided identifier
func (tr *twinRepository) RetrieveByID(_ context.Context, id string) (twins.Twin, error) {
	coll := tr.db.Collection(collectionName)
	var tw twins.Twin

	filter := bson.D{{"id", id}}

	if err := coll.FindOne(context.Background(), filter).Decode(&tw); err != nil {
		return tw, err
	}

	return tw, nil
}

// RetrieveByKey retrieves the twin having the provided key
func (tr *twinRepository) RetrieveByKey(_ context.Context, key string) (twins.Twin, error) {
	coll := tr.db.Collection(collectionName)
	var tw twins.Twin

	filter := bson.D{{"key", key}}

	if err := coll.FindOne(context.Background(), filter).Decode(&tw); err != nil {
		return tw, err
	}

	return tw, nil
}

// Remove removes the twin having the provided id
func (tr *twinRepository) RemoveByID(_ context.Context, id string) error {
	coll := tr.db.Collection(collectionName)

	filter := bson.D{{"id", id}}

	if _, err := coll.DeleteOne(context.Background(), filter); err != nil {
		return err
	}

	return nil
}

// Remove removes the twin having the provided key
func (tr *twinRepository) RemoveByKey(_ context.Context, key string) error {
	coll := tr.db.Collection(collectionName)

	filter := bson.D{{"id", key}}

	if _, err := coll.DeleteOne(context.Background(), filter); err != nil {
		return err
	}

	return nil
}

func toDBTwin(tw twins.Twin) bson.D {
	return bson.D{
		{"id", tw.ID},
		{"owner", tw.Owner},
		{"name", tw.Name},
		{"key", tw.Key},
		{"metadata", tw.Metadata},
	}
}

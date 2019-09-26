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
	"github.com/mainflux/mainflux/twins/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	maxNameSize           = 1024
	collectionName string = "mainflux"
)

const (
	maxNameSize           = 1024
	collectionName string = "mainflux"
)

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
func (tr *twinRepository) Save(ctx context.Context, tw twins.Twin) (string, error) {
	coll := tr.db.Collection(collectionName)

	if _, err := tr.RetrieveByID(ctx, tw.Owner, tw.ID); err == nil {
		return "", twins.ErrConflict
	}
	if _, err := tr.RetrieveByKey(ctx, tw.Key); err == nil {
		return "", twins.ErrConflict
	}

	dbtw, err := toDBTwin(tw)
	if err != nil {
		return "", err
	}

	if _, err := coll.InsertOne(context.Background(), dbtw); err != nil {
		return "", err
	}

	return tw.ID, nil
}

// Update performs an update to the existing twins. A non-nil error is
// returned to indicate operation failure.
func (tr *twinRepository) Update(ctx context.Context, tw twins.Twin) error {
	coll := tr.db.Collection(collectionName)

	if _, err := tr.RetrieveByID(ctx, tw.Owner, tw.ID); err != nil {
		return twins.ErrNotFound
	}

	dbtw, err := toDBTwin(tw)
	if err != nil {
		return err
	}

	filter := bson.D{{"id", tw.ID}}
	update := bson.D{{"$set", dbtw}}
	if _, err := coll.UpdateOne(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}

// UpdateKey performs an update key of the existing twin. A non-nil error is
// returned to indicate operation failure.
func (tr *twinRepository) UpdateKey(ctx context.Context, owner, id, key string) error {
	coll := tr.db.Collection(collectionName)

	if _, err := tr.RetrieveByID(ctx, owner, id); err != nil {
		return twins.ErrNotFound
	}

	if err := uuid.New().IsValid(key); err != nil {
		return err
	}

	filter := bson.D{{"id", id}}
	update := bson.D{{"$set", bson.D{{"key", key}}}}

	if _, err := coll.UpdateOne(context.Background(), filter, update); err != nil {
		return err
	}

	return nil
}

// RetrieveByID retrieves the twin having the provided identifier
func (tr *twinRepository) RetrieveByID(_ context.Context, owner, id string) (twins.Twin, error) {
	coll := tr.db.Collection(collectionName)
	var tw twins.Twin

	if err := uuid.New().IsValid(id); err != nil {
		return tw, err
	}

	filter := bson.D{{"id", id}}
	if err := coll.FindOne(context.Background(), filter).Decode(&tw); err != nil {
		return tw, twins.ErrNotFound
	}

	return tw, nil
}

// RetrieveByKey retrieves the twin having the provided key
func (tr *twinRepository) RetrieveByKey(_ context.Context, key string) (string, error) {
	coll := tr.db.Collection(collectionName)
	var tw twins.Twin

	filter := bson.D{{"key", key}}

	if err := coll.FindOne(context.Background(), filter).Decode(&tw); err != nil {
		return "", err
	}

	return tw.ID, nil
}

// Remove removes the twin having the provided id
func (tr *twinRepository) Remove(ctx context.Context, owner, id string) error {
	coll := tr.db.Collection(collectionName)

	if err := uuid.New().IsValid(id); err != nil {
		return err
	}

	if _, err := tr.RetrieveByID(ctx, "owner", id); err != nil {
		return twins.ErrNotFound
	}

	filter := bson.D{{"id", id}}
	if _, err := coll.DeleteOne(context.Background(), filter); err != nil {
		return err
	}

	return nil
}

func toDBTwin(tw twins.Twin) (bson.D, error) {
	// invalid name
	if len(tw.Name) > maxNameSize {
		return bson.D{}, twins.ErrMalformedEntity
	}

	// invalid id
	if err := uuid.New().IsValid(tw.ID); err != nil {
		return bson.D{}, err
	}

	// invalid key
	if err := uuid.New().IsValid(tw.Key); err != nil {
		return bson.D{}, err
	}

	return bson.D{
		{"id", tw.ID},
		{"owner", tw.Owner},
		{"name", tw.Name},
		{"key", tw.Key},
		{"metadata", tw.Metadata},
	}, nil
}

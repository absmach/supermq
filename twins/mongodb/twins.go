//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package mongodb

import (
	"context"
	"fmt"

	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	maxNameSize            = 1024
	twinsCollection string = "twins"
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

// Save persists the twin
func (tr *twinRepository) Save(ctx context.Context, tw twins.Twin) (string, error) {
	coll := tr.db.Collection(twinsCollection)

	if _, err := tr.RetrieveByID(ctx, tw.Owner, tw.ID); err == nil {
		return "", twins.ErrConflict
	}
	if _, err := tr.RetrieveByKey(ctx, tw.Key); err == nil {
		return "", twins.ErrConflict
	}

	if err := validate(tw); err != nil {
		return "", err
	}

	if _, err := coll.InsertOne(context.Background(), tw); err != nil {
		return "", err
	}

	return tw.ID, nil
}

// Update performs an update to the existing twins. A non-nil error is
// returned to indicate operation failure.
func (tr *twinRepository) Update(ctx context.Context, tw twins.Twin) error {
	if err := validate(tw); err != nil {
		return err
	}

	coll := tr.db.Collection(twinsCollection)

	filter := bson.D{{"id", tw.ID}}
	update := bson.D{{"$set", tw}}
	res, err := coll.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount < 1 {
		return twins.ErrNotFound
	}

	return nil
}

// UpdateKey performs an update key of the existing twin. A non-nil error is
// returned to indicate operation failure.
func (tr *twinRepository) UpdateKey(ctx context.Context, owner, id, key string) error {
	coll := tr.db.Collection(twinsCollection)

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
	coll := tr.db.Collection(twinsCollection)
	var tw twins.Twin

	if err := uuid.New().IsValid(id); err != nil {
		return tw, err
	}

	filter := bson.D{{"id", id}}

	if err := coll.FindOne(context.Background(), filter).Decode(&tw); err != nil {
		fmt.Printf("%s\n", err)
		return tw, twins.ErrNotFound
	}

	return tw, nil
}

// RetrieveByKey retrieves the twin having the provided key
func (tr *twinRepository) RetrieveByKey(_ context.Context, key string) (string, error) {
	coll := tr.db.Collection(twinsCollection)
	var tw twins.Twin

	filter := bson.D{{"key", key}}

	if err := coll.FindOne(context.Background(), filter).Decode(&tw); err != nil {
		return "", err
	}

	return tw.ID, nil
}

func decodeTwins(ctx context.Context, cur *mongo.Cursor) ([]twins.Twin, error) {
	defer cur.Close(ctx)
	var results []twins.Twin
	for cur.Next(ctx) {
		var elem twins.Twin
		err := cur.Decode(&elem)
		if err != nil {
			return []twins.Twin{}, nil
		}
		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return []twins.Twin{}, nil
	}
	return results, nil
}

func (tr *twinRepository) RetrieveAll(ctx context.Context, owner string, offset uint64, limit uint64, name string, metadata twins.Metadata) (twins.TwinsPage, error) {
	coll := tr.db.Collection(twinsCollection)

	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))

	filter := bson.D{}

	if owner != "" {
		filter = append(filter, bson.E{"owner", owner})
	}
	if name != "" {
		filter = append(filter, bson.E{"name", name})
	}
	if len(metadata) > 0 {
		filter = append(filter, bson.E{"metadata", metadata})
	}
	cur, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return twins.TwinsPage{}, err
	}

	results, err := decodeTwins(ctx, cur)
	if err != nil {
		return twins.TwinsPage{}, err
	}

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return twins.TwinsPage{}, err
	}

	return twins.TwinsPage{
		Twins: results,
		PageMetadata: twins.PageMetadata{
			Total:  uint64(total),
			Offset: offset,
			Limit:  limit,
		},
	}, nil
}

func (tr *twinRepository) RetrieveByThing(ctx context.Context, thing string, offset uint64, limit uint64) (twins.TwinsPage, error) {
	if err := uuid.New().IsValid(thing); err != nil {
		return twins.TwinsPage{}, twins.ErrNotFound
	}

	coll := tr.db.Collection(twinsCollection)

	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))

	filter := bson.D{{"thingid", thing}}
	cur, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return twins.TwinsPage{}, err
	}

	results, err := decodeTwins(ctx, cur)
	if err != nil {
		return twins.TwinsPage{}, err
	}

	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return twins.TwinsPage{}, err
	}

	return twins.TwinsPage{
		Twins: results,
		PageMetadata: twins.PageMetadata{
			Total:  uint64(total),
			Offset: offset,
			Limit:  limit,
		},
	}, nil
}

// Remove removes the twin having the provided id
func (tr *twinRepository) Remove(ctx context.Context, owner, id string) error {
	coll := tr.db.Collection(twinsCollection)

	if err := uuid.New().IsValid(id); err != nil {
		return err
	}

	filter := bson.D{{"id", id}}
	res, err := coll.DeleteOne(context.Background(), filter)
	if err != nil {
		return err
	}

	if res.DeletedCount < 1 {
		return twins.ErrNotFound
	}

	return nil
}

func validate(tw twins.Twin) error {
	if len(tw.Name) > maxNameSize {
		return twins.ErrMalformedEntity
	}
	if err := uuid.New().IsValid(tw.ID); err != nil {
		return twins.ErrMalformedEntity
	}
	if err := uuid.New().IsValid(tw.Key); err != nil {
		return twins.ErrMalformedEntity
	}
	return nil
}

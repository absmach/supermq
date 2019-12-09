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

const (
	statesCollection string = "states"
)

type stateRepository struct {
	db *mongo.Database
}

var _ twins.StateRepository = (*stateRepository)(nil)

// NewStateRepository instantiates a MongoDB implementation of state
// repository.
func NewStateRepository(db *mongo.Database) twins.StateRepository {
	return &stateRepository{
		db: db,
	}
}

// SaveState persists the state
func (sr *stateRepository) Save(ctx context.Context, st twins.State) error {
	coll := sr.db.Collection(statesCollection)

	if _, err := coll.InsertOne(context.Background(), st); err != nil {
		return err
	}

	return nil
}

// CountStates returns the number of states related to twin
func (sr *stateRepository) Count(ctx context.Context, tw twins.Twin) (int64, error) {
	coll := sr.db.Collection(statesCollection)

	filter := bson.D{{"twinid", tw.ID}}
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return total, nil
}

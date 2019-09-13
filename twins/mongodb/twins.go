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
	"go.mongodb.org/mongo-driver/mongo"
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

// Save persists the twin. Successful operation is indicated by non-nil
// error response.
func (tr *twinRepository) Save(context.Context, twins.Twin) (string, error) {
	return "", nil
}

// Update performs an update to the existing twins. A non-nil error is
// returned to indicate operation failure.
func (tr *twinRepository) Update(context.Context, twins.Twin) error {
	return nil
}

// RetrieveByID retrieves the twin having the provided identifier, that is owned
// by the specified thing.
func (tr *twinRepository) RetrieveByID(context.Context, string, string) (twins.Twin, error) {
	return twins.Twin{}, nil
}

// RetrieveByKey retrieves the twin having the provided key, that is owned
// by the specified thing.
func (tr *twinRepository) RetrieveByKey(context.Context, string) (string, error) {
	return "", nil
}

// Remove removes the twin having the provided identifier, that is owned
// by the specified thing.
func (tr *twinRepository) Remove(context.Context, string, string) error {
	return nil
}

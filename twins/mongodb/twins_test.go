//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package mongodb_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/mongodb"
	"github.com/mainflux/mainflux/twins/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	port       string
	addr       string
	testLog, _ = log.New(os.Stdout, log.Info.String())
	testDB     = "test"
	collection = "mainflux"
	db         mongo.Database
	msgsNum    = 10
	owner      = "mainflux@mainflux.com"
	name       = "twin"
)

func TestTwinSave(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)
	repo := mongodb.NewTwinRepository(db)

	for i := 0; i < msgsNum; i++ {
		twid, err := uuid.New().ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
		twkey, err := uuid.New().ID()
		require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

		tw := twins.Twin{
			ID:    twid,
			Owner: string(i) + owner,
			Name:  name + string(i),
			Key:   twkey,
		}

		err = repo.Save(context.TODO(), tw)
	}

	count, err := db.Collection(collection).CountDocuments(context.Background(), bson.D{})

	assert.Nil(t, err, fmt.Sprintf("Querying database expected to succeed: %s.\n", err))
	assert.Nil(t, err, fmt.Sprintf("Save operation expected to succeed: %s.\n", err))
	assert.Equal(t, int64(msgsNum), count, fmt.Sprintf("Expected to have %d value, found %d instead.\n", msgsNum, count))
}

// func TestThingUpdate(t *testing.T) {
// 	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
// 	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

// 	db := client.Database(testDB)
// 	repo := mongodb.NewTwinRepository(db)

// 	email := "twin-update@example.com"
// 	validName := "mfx_shadow"

// 	twid, err := uuid.New().ID()
// 	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
// 	twkey, err := uuid.New().ID()
// 	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

// 	twin := twins.Twin{
// 		ID:    twid,
// 		Owner: email,
// 		Name:  validName,
// 		Key:   twkey,
// 	}

// }

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
	"strings"
	"testing"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/mongodb"
	"github.com/mainflux/mainflux/twins/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const maxNameSize = 1024

var (
	port        string
	addr        string
	testLog, _  = log.New(os.Stdout, log.Info.String())
	testDB      = "test"
	collection  = "mainflux"
	db          mongo.Database
	msgsNum     = 10
	invalidName = strings.Repeat("m", maxNameSize+1)
)

func TestTwinSave(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)
	repo := mongodb.NewTwinRepository(db)

	email := "thing-save@example.com"

	twid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	twkey, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	nonexistentTwinKey, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	twin := twins.Twin{
		ID:    twid,
		Owner: email,
		Key:   twkey,
	}

	cases := []struct {
		desc string
		twin twins.Twin
		err  error
	}{
		{
			desc: "create new twin",
			twin: twin,
			err:  nil,
		},
		{
			desc: "create twin with conflicting key",
			twin: twin,
			err:  twins.ErrConflict,
		},
		{
			desc: "create twin with invalid ID",
			twin: twins.Twin{
				ID:    "invalid",
				Owner: email,
				Key:   twkey,
			},
			err: twins.ErrMalformedEntity,
		},
		{
			desc: "create twin with invalid Key",
			twin: twins.Twin{
				ID:    twid,
				Owner: email,
				Key:   nonexistentTwinKey,
			},
			err: twins.ErrConflict,
		},
		{
			desc: "create twin with invalid name",
			twin: twins.Twin{
				ID:    twid,
				Owner: email,
				Key:   twkey,
				Name:  invalidName,
			},
			err: twins.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		err := repo.Save(context.Background(), tc.twin)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
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

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

const (
	maxNameSize = 1024
	msgsNum     = 10
	testDB      = "test"
	collection  = "mainflux"
)

var (
	port        string
	addr        string
	testLog, _  = log.New(os.Stdout, log.Info.String())
	idp         = uuid.New()
	db          mongo.Database
	invalidName = strings.Repeat("m", maxNameSize+1)
)

func TestTwinSave(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)
	repo := mongodb.NewTwinRepository(db)

	email := "twin-save@example.com"

	twid, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	twkey, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	nonexistentTwinID, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	nonexistentTwinKey, err := idp.ID()
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
			desc: "create twin with existing ID",
			twin: twins.Twin{
				ID:    twid,
				Owner: email,
				Key:   nonexistentTwinKey,
			},
			err: twins.ErrConflict,
		},
		{
			desc: "create twin with existing Key",
			twin: twins.Twin{
				ID:    nonexistentTwinID,
				Owner: email,
				Key:   twkey,
			},
			err: twins.ErrConflict,
		},
		{
			desc: "create twin with invalid name",
			twin: twins.Twin{
				ID:    nonexistentTwinID,
				Owner: email,
				Key:   nonexistentTwinKey,
				Name:  invalidName,
			},
			err: twins.ErrMalformedEntity,
		},
		{
			desc: "create twin with invalid ID",
			twin: twins.Twin{
				ID:    "invalid",
				Owner: email,
				Key:   nonexistentTwinKey,
			},
			err: twins.ErrMalformedEntity,
		},
		{
			desc: "create twin with invalid Key",
			twin: twins.Twin{
				ID:    nonexistentTwinID,
				Owner: email,
				Key:   "invalid",
			},
			err: twins.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		err := repo.Save(context.Background(), tc.twin)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestThingUpdate(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)
	repo := mongodb.NewTwinRepository(db)

	email := "twin-update@example.com"
	validName := "mfx_shadow"

	twid, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	twkey, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	nonexistentTwinID, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	twin := twins.Twin{
		ID:    twid,
		Owner: email,
		Name:  validName,
		Key:   twkey,
	}

	if err := repo.Save(context.Background(), twin); err != nil {
		testLog.Error(err.Error())
	}

	cases := []struct {
		desc string
		twin twins.Twin
		err  error
	}{
		{
			desc: "update existing twin",
			twin: twin,
			err:  nil,
		},
		{
			desc: "update non-existing twin",
			twin: twins.Twin{
				ID:    nonexistentTwinID,
				Owner: email,
				Name:  validName,
				Key:   twkey,
			},
			err: twins.ErrNotFound,
		},
		{
			desc: "update twin with invalid name",
			twin: twins.Twin{
				ID:    twid,
				Owner: email,
				Name:  validName,
				Key:   "invalid",
			},
			err: twins.ErrMalformedEntity,
		},
		{
			desc: "update twin with invalid name",
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
		err := repo.Update(context.Background(), tc.twin)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestTwinUpdateKey(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)
	repo := mongodb.NewTwinRepository(db)

	email := "twin-update@example.com"
	validName := "mfx_shadow"

	twid, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	twkey, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	nonexistentTwinID, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	nonexistentTwinKey, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	twin := twins.Twin{
		ID:    twid,
		Owner: email,
		Name:  validName,
		Key:   twkey,
	}

	if err := repo.Save(context.Background(), twin); err != nil {
		testLog.Error(err.Error())
	}

	cases := []struct {
		desc string
		id   string
		key  string
		err  error
	}{
		{
			desc: "update key of an existing twin",
			id:   twin.ID,
			key:  nonexistentTwinKey,
			err:  nil,
		},
		{
			desc: "update key of a non-existing twin",
			id:   nonexistentTwinID,
			key:  nonexistentTwinKey,
			err:  twins.ErrNotFound,
		},
		{
			desc: "update key with an invalid key",
			id:   twin.ID,
			key:  "invalid",
			err:  twins.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		err := repo.UpdateKey(context.Background(), tc.id, tc.key)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestTwinRetrieveByID(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)
	repo := mongodb.NewTwinRepository(db)

	email := "twin-retrievebyid@example.com"
	validName := "mfx_shadow"

	twid, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	twkey, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	nonexistentTwinID, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	twin := twins.Twin{
		ID:    twid,
		Owner: email,
		Name:  validName,
		Key:   twkey,
	}

	if err := repo.Save(context.Background(), twin); err != nil {
		testLog.Error(err.Error())
	}

	cases := []struct {
		desc string
		id   string
		err  error
	}{
		{
			desc: "retreive an existing twin",
			id:   twin.ID,
			err:  nil,
		},
		{
			desc: "retrieve a non-existing twin",
			id:   nonexistentTwinID,
			err:  twins.ErrNotFound,
		},
		{
			desc: "retrieve with an invalid id",
			id:   "invalid",
			err:  twins.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		_, err := repo.RetrieveByID(context.Background(), tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestTwinRemove(t *testing.T) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)
	repo := mongodb.NewTwinRepository(db)

	email := "twin-remove@example.com"
	validName := "mfx_shadow"

	twid, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	twkey, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	nonexistentTwinID, err := idp.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	twin := twins.Twin{
		ID:    twid,
		Owner: email,
		Name:  validName,
		Key:   twkey,
	}

	if err := repo.Save(context.Background(), twin); err != nil {
		testLog.Error(err.Error())
	}

	cases := []struct {
		desc string
		id   string
		err  error
	}{
		{
			desc: "remove an existing twin",
			id:   twin.ID,
			err:  nil,
		},
		{
			desc: "remove a non-existing twin",
			id:   nonexistentTwinID,
			err:  twins.ErrNotFound,
		},
		{
			desc: "retrieve with an invalid id",
			id:   "invalid",
			err:  twins.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		err := repo.Remove(context.Background(), tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

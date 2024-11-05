// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/0x6flab/namegenerator"
	"github.com/absmach/magistrala/clients"
	"github.com/absmach/magistrala/clients/postgres"
	"github.com/absmach/magistrala/internal/testsutil"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const maxNameSize = 1024

var (
	invalidName     = strings.Repeat("m", maxNameSize+10)
	thingIdentity   = "thing-identity@example.com"
	thingName       = "thing name"
	invalidDomainID = strings.Repeat("m", maxNameSize+10)
	namegen         = namegenerator.NewGenerator()
)

func TestClientsSave(t *testing.T) {
	t.Cleanup(func() {
		_, err := db.Exec("DELETE FROM clients")
		require.Nil(t, err, fmt.Sprintf("clean clients unexpected error: %s", err))
	})
	repo := postgres.NewRepository(database)

	uid := testsutil.GenerateUUID(t)
	domainID := testsutil.GenerateUUID(t)
	secret := testsutil.GenerateUUID(t)

	cases := []struct {
		desc   string
		things []clients.Client
		err    error
	}{
		{
			desc: "add new thing successfully",
			things: []clients.Client{
				{
					ID:     uid,
					Domain: domainID,
					Name:   thingName,
					Credentials: clients.Credentials{
						Identity: thingIdentity,
						Secret:   secret,
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: nil,
		},
		{
			desc: "add multiple things successfully",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: testsutil.GenerateUUID(t),
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Secret: testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: testsutil.GenerateUUID(t),
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Secret: testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: testsutil.GenerateUUID(t),
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Secret: testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: nil,
		},
		{
			desc: "add new thing with duplicate secret",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: domainID,
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Identity: thingIdentity,
						Secret:   secret,
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "add multiple things with one thing having duplicate secret",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: testsutil.GenerateUUID(t),
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Secret: testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: domainID,
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Identity: thingIdentity,
						Secret:   secret,
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "add new thing without domain id",
			things: []clients.Client{
				{
					ID:   testsutil.GenerateUUID(t),
					Name: thingName,
					Credentials: clients.Credentials{
						Identity: "withoutdomain-thing@example.com",
						Secret:   testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: nil,
		},
		{
			desc: "add thing with invalid thing id",
			things: []clients.Client{
				{
					ID:     invalidName,
					Domain: domainID,
					Name:   thingName,
					Credentials: clients.Credentials{
						Identity: "invalidid-thing@example.com",
						Secret:   testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "add multiple things with one thing having invalid thing id",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: testsutil.GenerateUUID(t),
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Secret: testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
				{
					ID:     invalidName,
					Domain: testsutil.GenerateUUID(t),
					Name:   namegen.Generate(),
					Credentials: clients.Credentials{
						Secret: testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "add thing with invalid thing name",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Name:   invalidName,
					Domain: domainID,
					Credentials: clients.Credentials{
						Identity: "invalidname-thing@example.com",
						Secret:   testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "add thing with invalid thing domain id",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: invalidDomainID,
					Credentials: clients.Credentials{
						Identity: "invaliddomainid-thing@example.com",
						Secret:   testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "add thing with invalid thing identity",
			things: []clients.Client{
				{
					ID:   testsutil.GenerateUUID(t),
					Name: thingName,
					Credentials: clients.Credentials{
						Identity: invalidName,
						Secret:   testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
					Status:   clients.EnabledStatus,
				},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "add thing with a missing thing identity",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: testsutil.GenerateUUID(t),
					Name:   "missing-thing-identity",
					Credentials: clients.Credentials{
						Identity: "",
						Secret:   testsutil.GenerateUUID(t),
					},
					Metadata: clients.Metadata{},
				},
			},
			err: nil,
		},
		{
			desc: "add thing with a missing thing secret",
			things: []clients.Client{
				{
					ID:     testsutil.GenerateUUID(t),
					Domain: testsutil.GenerateUUID(t),
					Credentials: clients.Credentials{
						Identity: "missing-thing-secret@example.com",
						Secret:   "",
					},
					Metadata: clients.Metadata{},
				},
			},
			err: nil,
		},
		{
			desc: "add a thing with invalid metadata",
			things: []clients.Client{
				{
					ID:   testsutil.GenerateUUID(t),
					Name: namegen.Generate(),
					Credentials: clients.Credentials{
						Identity: fmt.Sprintf("%s@example.com", namegen.Generate()),
						Secret:   testsutil.GenerateUUID(t),
					},
					Metadata: map[string]interface{}{
						"key": make(chan int),
					},
				},
			},
			err: errors.ErrMalformedEntity,
		},
	}
	for _, tc := range cases {
		rThings, err := repo.Save(context.Background(), tc.things...)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			for i := range rThings {
				tc.things[i].Credentials.Secret = rThings[i].Credentials.Secret
			}
			assert.Equal(t, tc.things, rThings, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.things, rThings))
		}
	}
}

func TestThingsRetrieveBySecret(t *testing.T) {
	t.Cleanup(func() {
		_, err := db.Exec("DELETE FROM clients")
		require.Nil(t, err, fmt.Sprintf("clean clients unexpected error: %s", err))
	})
	repo := postgres.NewRepository(database)

	thing := clients.Client{
		ID:   testsutil.GenerateUUID(t),
		Name: thingName,
		Credentials: clients.Credentials{
			Identity: thingIdentity,
			Secret:   testsutil.GenerateUUID(t),
		},
		Metadata: clients.Metadata{},
		Status:   clients.EnabledStatus,
	}

	_, err := repo.Save(context.Background(), thing)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc     string
		secret   string
		response clients.Client
		err      error
	}{
		{
			desc:     "retrieve thing by secret successfully",
			secret:   thing.Credentials.Secret,
			response: thing,
			err:      nil,
		},
		{
			desc:     "retrieve thing by invalid secret",
			secret:   "non-existent-secret",
			response: clients.Client{},
			err:      repoerr.ErrNotFound,
		},
		{
			desc:     "retrieve thing by empty secret",
			secret:   "",
			response: clients.Client{},
			err:      repoerr.ErrNotFound,
		},
	}

	for _, tc := range cases {
		res, err := repo.RetrieveBySecret(context.Background(), tc.secret)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, res, tc.response, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, res))
	}
}

func TestRetrieveByID(t *testing.T) {
	t.Cleanup(func() {
		_, err := db.Exec("DELETE FROM clients")
		require.Nil(t, err, fmt.Sprintf("clean clients unexpected error: %s", err))
	})
	repo := postgres.NewRepository(database)

	thing := clients.Client{
		ID:   testsutil.GenerateUUID(t),
		Name: thingName,
		Credentials: clients.Credentials{
			Identity: thingIdentity,
			Secret:   testsutil.GenerateUUID(t),
		},
		Metadata: clients.Metadata{},
		Status:   clients.EnabledStatus,
	}

	_, err := repo.Save(context.Background(), thing)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc     string
		id       string
		response clients.Client
		err      error
	}{
		{
			desc:     "successfully",
			id:       thing.ID,
			response: thing,
			err:      nil,
		},
		{
			desc:     "with invalid id",
			id:       testsutil.GenerateUUID(t),
			response: clients.Client{},
			err:      repoerr.ErrNotFound,
		},
		{
			desc:     "with empty id",
			id:       "",
			response: clients.Client{},
			err:      repoerr.ErrNotFound,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			cli, err := repo.RetrieveByID(context.Background(), c.id)
			assert.True(t, errors.Contains(err, c.err), fmt.Sprintf("expected %s got %s\n", c.err, err))
			if err == nil {
				assert.Equal(t, thing.ID, cli.ID)
				assert.Equal(t, thing.Name, cli.Name)
				assert.Equal(t, thing.Metadata, cli.Metadata)
				assert.Equal(t, thing.Credentials.Identity, cli.Credentials.Identity)
				assert.Equal(t, thing.Credentials.Secret, cli.Credentials.Secret)
				assert.Equal(t, thing.Status, cli.Status)
			}
		})
	}
}

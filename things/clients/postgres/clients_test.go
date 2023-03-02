package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/internal/testsutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things/clients"
	cpostgres "github.com/mainflux/mainflux/things/clients/postgres"
	"github.com/mainflux/mainflux/things/groups"
	gpostgres "github.com/mainflux/mainflux/things/groups/postgres"
	"github.com/mainflux/mainflux/things/policies"
	ppostgres "github.com/mainflux/mainflux/things/policies/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	maxNameSize = 1024
)

var (
	idProvider     = uuid.New()
	invalidName    = strings.Repeat("m", maxNameSize+10)
	clientIdentity = "client-identity@example.com"
	clientName     = "client name"
	wrongName      = "wrong-name"
	wrongID        = "wrong-id"
)

func TestClientsSave(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)

	uid := testsutil.GenerateUUID(t, idProvider)

	cases := []struct {
		desc   string
		client clients.Client
		err    error
	}{
		{
			desc: "add new client successfully",
			client: clients.Client{
				ID:   uid,
				Name: clientName,
				Credentials: clients.Credentials{
					Identity: clientIdentity,
					Secret:   testsutil.GenerateUUID(t, idProvider),
				},
				Metadata: clients.Metadata{},
				Status:   clients.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "add new client with an owner",
			client: clients.Client{
				ID:    testsutil.GenerateUUID(t, idProvider),
				Owner: uid,
				Name:  clientName,
				Credentials: clients.Credentials{
					Identity: "withowner-client@example.com",
					Secret:   testsutil.GenerateUUID(t, idProvider),
				},
				Metadata: clients.Metadata{},
				Status:   clients.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "add client with invalid client id",
			client: clients.Client{
				ID:   invalidName,
				Name: clientName,
				Credentials: clients.Credentials{
					Identity: "invalidid-client@example.com",
					Secret:   testsutil.GenerateUUID(t, idProvider),
				},
				Metadata: clients.Metadata{},
				Status:   clients.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add client with invalid client name",
			client: clients.Client{
				ID:   testsutil.GenerateUUID(t, idProvider),
				Name: invalidName,
				Credentials: clients.Credentials{
					Identity: "invalidname-client@example.com",
					Secret:   testsutil.GenerateUUID(t, idProvider),
				},
				Metadata: clients.Metadata{},
				Status:   clients.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add client with invalid client owner",
			client: clients.Client{
				ID:    testsutil.GenerateUUID(t, idProvider),
				Owner: invalidName,
				Credentials: clients.Credentials{
					Identity: "invalidowner-client@example.com",
					Secret:   testsutil.GenerateUUID(t, idProvider),
				},
				Metadata: clients.Metadata{},
				Status:   clients.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add client with invalid client identity",
			client: clients.Client{
				ID:   testsutil.GenerateUUID(t, idProvider),
				Name: clientName,
				Credentials: clients.Credentials{
					Identity: invalidName,
					Secret:   testsutil.GenerateUUID(t, idProvider),
				},
				Metadata: clients.Metadata{},
				Status:   clients.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "add client with a missing client identity",
			client: clients.Client{
				ID: testsutil.GenerateUUID(t, idProvider),
				Credentials: clients.Credentials{
					Identity: "",
					Secret:   testsutil.GenerateUUID(t, idProvider),
				},
				Metadata: clients.Metadata{},
			},
			err: nil,
		},
		{
			desc: "add client with a missing client secret",
			client: clients.Client{
				ID: testsutil.GenerateUUID(t, idProvider),
				Credentials: clients.Credentials{
					Identity: "missing-client-secret@example.com",
					Secret:   "",
				},
				Metadata: clients.Metadata{},
			},
			err: nil,
		},
	}
	for _, tc := range cases {
		rClient, err := repo.Save(context.Background(), tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			rClient[0].Credentials.Secret = tc.client.Credentials.Secret
			assert.Equal(t, tc.client, rClient[0], fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.client, rClient[0]))
		}
	}
}

func TestClientsRetrieveByID(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)

	client := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: clientName,
		Credentials: clients.Credentials{
			Identity: clientIdentity,
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Status: clients.EnabledStatus,
	}

	_, err := repo.Save(context.Background(), client)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := map[string]struct {
		ID  string
		err error
	}{
		"retrieve existing client":     {client.ID, nil},
		"retrieve non-existing client": {wrongID, errors.ErrNotFound},
	}

	for desc, tc := range cases {
		cli, err := repo.RetrieveByID(context.Background(), tc.ID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		if err == nil {
			assert.Equal(t, client.ID, cli.ID, fmt.Sprintf("retrieve client by ID : client ID : expected %s got %s\n", client.ID, cli.ID))
			assert.Equal(t, client.Name, cli.Name, fmt.Sprintf("retrieve client by ID : client Name : expected %s got %s\n", client.Name, cli.Name))
			assert.Equal(t, client.Credentials.Identity, cli.Credentials.Identity, fmt.Sprintf("retrieve client by ID : client Identity : expected %s got %s\n", client.Credentials.Identity, cli.Credentials.Identity))
			assert.Equal(t, client.Status, cli.Status, fmt.Sprintf("retrieve client by ID : client Status : expected %d got %d\n", client.Status, cli.Status))
		}
	}
}

func TestClientsRetrieveAll(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)
	grepo := gpostgres.NewRepository(database)
	prepo := ppostgres.NewRepository(database)

	var nClients = uint64(200)
	var ownerID string

	meta := clients.Metadata{
		"admin": "true",
	}
	wrongMeta := clients.Metadata{
		"admin": "false",
	}
	var expectedClients = []clients.Client{}

	var sharedGroup = groups.Group{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "shared-group",
	}
	_, err := grepo.Save(context.Background(), sharedGroup)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	for i := uint64(0); i < nClients; i++ {
		identity := fmt.Sprintf("TestRetrieveAll%d@example.com", i)
		client := clients.Client{
			ID:   testsutil.GenerateUUID(t, idProvider),
			Name: identity,
			Credentials: clients.Credentials{
				Identity: identity,
				Secret:   testsutil.GenerateUUID(t, idProvider),
			},
			Metadata: clients.Metadata{},
			Status:   clients.EnabledStatus,
		}
		if i == 1 {
			ownerID = client.ID
		}
		if i%10 == 0 {
			client.Owner = ownerID
			client.Metadata = meta
			client.Tags = []string{"Test"}
		}
		if i%50 == 0 {
			client.Status = clients.DisabledStatus
		}
		_, err := repo.Save(context.Background(), client)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		expectedClients = append(expectedClients, client)
		var policy = policies.Policy{
			Subject: client.ID,
			Object:  sharedGroup.ID,
			Actions: []string{"c_list"},
		}
		_, err = prepo.Save(context.Background(), policy)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	}

	cases := map[string]struct {
		size     uint64
		pm       clients.Page
		response []clients.Client
	}{
		"retrieve all clients empty page": {
			pm:       clients.Page{},
			response: []clients.Client{},
			size:     0,
		},
		"retrieve all clients": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Status: clients.AllStatus,
			},
			response: expectedClients,
			size:     200,
		},
		"retrieve all clients with limit": {
			pm: clients.Page{
				Offset: 0,
				Limit:  50,
				Status: clients.AllStatus,
			},
			response: expectedClients[0:50],
			size:     50,
		},
		"retrieve all clients with offset": {
			pm: clients.Page{
				Offset: 50,
				Limit:  nClients,
				Status: clients.AllStatus,
			},
			response: expectedClients[50:200],
			size:     150,
		},
		"retrieve all clients with limit and offset": {
			pm: clients.Page{
				Offset: 50,
				Limit:  50,
				Status: clients.AllStatus,
			},
			response: expectedClients[50:100],
			size:     50,
		},
		"retrieve all clients with limit and offset not full": {
			pm: clients.Page{
				Offset: 170,
				Limit:  50,
				Status: clients.AllStatus,
			},
			response: expectedClients[170:200],
			size:     30,
		},
		"retrieve all clients by metadata": {
			pm: clients.Page{
				Offset:   0,
				Limit:    nClients,
				Total:    nClients,
				Metadata: meta,
				Status:   clients.AllStatus,
			},
			response: []clients.Client{expectedClients[0], expectedClients[10], expectedClients[20], expectedClients[30], expectedClients[40], expectedClients[50], expectedClients[60],
				expectedClients[70], expectedClients[80], expectedClients[90], expectedClients[100], expectedClients[110], expectedClients[120], expectedClients[130],
				expectedClients[140], expectedClients[150], expectedClients[160], expectedClients[170], expectedClients[180], expectedClients[190],
			},
			size: 20,
		},
		"retrieve clients by wrong metadata": {
			pm: clients.Page{
				Offset:   0,
				Limit:    nClients,
				Total:    nClients,
				Metadata: wrongMeta,
				Status:   clients.AllStatus,
			},
			response: []clients.Client{},
			size:     0,
		},
		"retrieve all clients by name": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Name:   "TestRetrieveAll3@example.com",
				Status: clients.AllStatus,
			},
			response: []clients.Client{expectedClients[3]},
			size:     1,
		},
		"retrieve clients by wrong name": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Name:   wrongName,
				Status: clients.AllStatus,
			},
			response: []clients.Client{},
			size:     0,
		},
		"retrieve all clients by owner": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Owner:  ownerID,
				Status: clients.AllStatus,
			},
			response: []clients.Client{expectedClients[10], expectedClients[20], expectedClients[30], expectedClients[40], expectedClients[50], expectedClients[60],
				expectedClients[70], expectedClients[80], expectedClients[90], expectedClients[100], expectedClients[110], expectedClients[120], expectedClients[130],
				expectedClients[140], expectedClients[150], expectedClients[160], expectedClients[170], expectedClients[180], expectedClients[190],
			},
			size: 19,
		},
		"retrieve clients by wrong owner": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Owner:  wrongID,
				Status: clients.AllStatus,
			},
			response: []clients.Client{},
			size:     0,
		},
		"retrieve all clients by disabled status": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Status: clients.DisabledStatus,
			},
			response: []clients.Client{expectedClients[0], expectedClients[50], expectedClients[100], expectedClients[150]},
			size:     4,
		},
		"retrieve all clients by combined status": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Status: clients.AllStatus,
			},
			response: expectedClients,
			size:     200,
		},
		"retrieve clients by the wrong status": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Status: 10,
			},
			response: []clients.Client{},
			size:     0,
		},
		"retrieve all clients by tags": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Tag:    "Test",
				Status: clients.AllStatus,
			},
			response: []clients.Client{expectedClients[0], expectedClients[10], expectedClients[20], expectedClients[30], expectedClients[40], expectedClients[50], expectedClients[60],
				expectedClients[70], expectedClients[80], expectedClients[90], expectedClients[100], expectedClients[110], expectedClients[120], expectedClients[130],
				expectedClients[140], expectedClients[150], expectedClients[160], expectedClients[170], expectedClients[180], expectedClients[190],
			},
			size: 20,
		},
		"retrieve clients by wrong tags": {
			pm: clients.Page{
				Offset: 0,
				Limit:  nClients,
				Total:  nClients,
				Tag:    "wrongTags",
				Status: clients.AllStatus,
			},
			response: []clients.Client{},
			size:     0,
		},
	}
	for desc, tc := range cases {
		page, err := repo.RetrieveAll(context.Background(), tc.pm)
		size := uint64(len(page.Clients))
		assert.ElementsMatch(t, page.Clients, tc.response, fmt.Sprintf("%s: expected %v got %v\n", desc, tc.response, page.Clients))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.size, size))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestClientsUpdateMetadata(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)

	client1 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-client",
		Credentials: clients.Credentials{
			Identity: "client1-update@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Metadata: clients.Metadata{
			"name": "enabled-client",
		},
		Tags:   []string{"enabled", "tag1"},
		Status: clients.EnabledStatus,
	}

	client2 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-client",
		Credentials: clients.Credentials{
			Identity: "client2-update@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Metadata: clients.Metadata{
			"name": "disabled-client",
		},
		Tags:   []string{"disabled", "tag1"},
		Status: clients.DisabledStatus,
	}

	_, err := repo.Save(context.Background(), client1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new client with metadata: expected %v got %s\n", nil, err))
	_, err = repo.Save(context.Background(), client2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled client: expected %v got %s\n", nil, err))

	ucases := []struct {
		desc   string
		update string
		client clients.Client
		err    error
	}{
		{
			desc:   "update metadata for enabled client",
			update: "metadata",
			client: clients.Client{
				ID: client1.ID,
				Metadata: clients.Metadata{
					"update": "metadata",
				},
			},
			err: nil,
		},
		{
			desc:   "update metadata for disabled client",
			update: "metadata",
			client: clients.Client{
				ID: client2.ID,
				Metadata: clients.Metadata{
					"update": "metadata",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name for enabled client",
			update: "name",
			client: clients.Client{
				ID:   client1.ID,
				Name: "updated name",
			},
			err: nil,
		},
		{
			desc:   "update name for disabled client",
			update: "name",
			client: clients.Client{
				ID:   client2.ID,
				Name: "updated name",
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name and metadata for enabled client",
			update: "both",
			client: clients.Client{
				ID:   client1.ID,
				Name: "updated name and metadata",
				Metadata: clients.Metadata{
					"update": "name and metadata",
				},
			},
			err: nil,
		},
		{
			desc:   "update name and metadata for a disabled client",
			update: "both",
			client: clients.Client{
				ID:   client2.ID,
				Name: "updated name and metadata",
				Metadata: clients.Metadata{
					"update": "name and metadata",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update metadata for invalid client",
			update: "metadata",
			client: clients.Client{
				ID: wrongID,
				Metadata: clients.Metadata{
					"update": "metadata",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name for invalid client",
			update: "name",
			client: clients.Client{
				ID:   wrongID,
				Name: "updated name",
			},
			err: errors.ErrNotFound,
		},
		{
			desc:   "update name and metadata for invalid client",
			update: "both",
			client: clients.Client{
				ID:   client2.ID,
				Name: "updated name and metadata",
				Metadata: clients.Metadata{
					"update": "name and metadata",
				},
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		expected, err := repo.Update(context.Background(), tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			if tc.client.Name != "" {
				assert.Equal(t, expected.Name, tc.client.Name, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, expected.Name, tc.client.Name))
			}
			if tc.client.Metadata != nil {
				assert.Equal(t, expected.Metadata, tc.client.Metadata, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, expected.Metadata, tc.client.Metadata))
			}

		}
	}
}

func TestClientsUpdateTags(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)

	client1 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-client-with-tags",
		Credentials: clients.Credentials{
			Identity: "client1-update-tags@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Tags:   []string{"test", "enabled"},
		Status: clients.EnabledStatus,
	}
	client2 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-client-with-tags",
		Credentials: clients.Credentials{
			Identity: "client2-update-tags@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Tags:   []string{"test", "disabled"},
		Status: clients.DisabledStatus,
	}

	_, err := repo.Save(context.Background(), client1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new client with tags: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, client1.ID, client1.ID, fmt.Sprintf("add new client with tags: expected %v got %s\n", nil, err))
	}
	_, err = repo.Save(context.Background(), client2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled client with tags: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, client2.ID, client2.ID, fmt.Sprintf("add new disabled client with tags: expected %v got %s\n", nil, err))
	}
	ucases := []struct {
		desc   string
		client clients.Client
		err    error
	}{
		{
			desc: "update tags for enabled client",
			client: clients.Client{
				ID:   client1.ID,
				Tags: []string{"updated"},
			},
			err: nil,
		},
		{
			desc: "update tags for disabled client",
			client: clients.Client{
				ID:   client2.ID,
				Tags: []string{"updated"},
			},
			err: errors.ErrNotFound,
		},
		{
			desc: "update tags for invalid client",
			client: clients.Client{
				ID:   wrongID,
				Tags: []string{"updated"},
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		expected, err := repo.UpdateTags(context.Background(), tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.client.Tags, expected.Tags, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.client.Tags, expected.Tags))
		}
	}
}

func TestClientsUpdateSecret(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)

	client1 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-client",
		Credentials: clients.Credentials{
			Identity: "client1-update@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Status: clients.EnabledStatus,
	}
	client2 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-client",
		Credentials: clients.Credentials{
			Identity: "client2-update@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Status: clients.DisabledStatus,
	}

	rClient1, err := repo.Save(context.Background(), client1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new client: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, client1.ID, rClient1[0].ID, fmt.Sprintf("add new client: expected %v got %s\n", nil, err))
	}
	rClient2, err := repo.Save(context.Background(), client2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled client: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, client2.ID, rClient2[0].ID, fmt.Sprintf("add new disabled client: expected %v got %s\n", nil, err))
	}

	ucases := []struct {
		desc   string
		client clients.Client
		err    error
	}{
		{
			desc: "update secret for enabled client",
			client: clients.Client{
				ID: client1.ID,
				Credentials: clients.Credentials{
					Identity: "client1-update@example.com",
					Secret:   "newpassword",
				},
			},
			err: nil,
		},
		{
			desc: "update secret for disabled client",
			client: clients.Client{
				ID: client2.ID,
				Credentials: clients.Credentials{
					Identity: "client2-update@example.com",
					Secret:   "newpassword",
				},
			},
			err: errors.ErrNotFound,
		},
		{
			desc: "update secret for invalid client",
			client: clients.Client{
				ID: wrongID,
				Credentials: clients.Credentials{
					Identity: "client3-update@example.com",
					Secret:   "newpassword",
				},
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		_, err := repo.UpdateSecret(context.Background(), tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			c, err := repo.RetrieveByID(context.Background(), tc.client.ID)
			require.Nil(t, err, fmt.Sprintf("retrieve client by id during update of secret unexpected error: %s", err))
			assert.Equal(t, tc.client.Credentials.Secret, c.Credentials.Secret, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.client.Credentials.Secret, c.Credentials.Secret))
		}
	}
}

func TestClientsUpdateOwner(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)

	client1 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-client-with-owner",
		Credentials: clients.Credentials{
			Identity: "client1-update-owner@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Owner:  testsutil.GenerateUUID(t, idProvider),
		Status: clients.EnabledStatus,
	}
	client2 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "disabled-client-with-owner",
		Credentials: clients.Credentials{
			Identity: "client2-update-owner@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Owner:  testsutil.GenerateUUID(t, idProvider),
		Status: clients.DisabledStatus,
	}

	_, err := repo.Save(context.Background(), client1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new client with owner: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, client1.ID, client1.ID, fmt.Sprintf("add new client with owner: expected %v got %s\n", nil, err))
	}
	_, err = repo.Save(context.Background(), client2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled client with owner: expected %v got %s\n", nil, err))
	if err == nil {
		assert.Equal(t, client2.ID, client2.ID, fmt.Sprintf("add new disabled client with owner: expected %v got %s\n", nil, err))
	}
	ucases := []struct {
		desc   string
		client clients.Client
		err    error
	}{
		{
			desc: "update owner for enabled client",
			client: clients.Client{
				ID:    client1.ID,
				Owner: testsutil.GenerateUUID(t, idProvider),
			},
			err: nil,
		},
		{
			desc: "update owner for disabled client",
			client: clients.Client{
				ID:    client2.ID,
				Owner: testsutil.GenerateUUID(t, idProvider),
			},
			err: errors.ErrNotFound,
		},
		{
			desc: "update owner for invalid client",
			client: clients.Client{
				ID:    wrongID,
				Owner: testsutil.GenerateUUID(t, idProvider),
			},
			err: errors.ErrNotFound,
		},
	}
	for _, tc := range ucases {
		expected, err := repo.UpdateOwner(context.Background(), tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.client.Owner, expected.Owner, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.client.Owner, expected.Owner))
		}
	}
}

func TestClientsChangeStatus(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := cpostgres.NewRepository(database)

	client1 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "enabled-client",
		Credentials: clients.Credentials{
			Identity: "client1-update@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Status: clients.EnabledStatus,
	}

	_, err := repo.Save(context.Background(), client1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new client: expected %v got %s\n", nil, err))

	ucases := []struct {
		desc   string
		client clients.Client
		err    error
	}{
		{
			desc: "change client status for an enabled client",
			client: clients.Client{
				ID:     client1.ID,
				Status: 0,
			},
			err: nil,
		},
		{
			desc: "change client status for a disabled client",
			client: clients.Client{
				ID:     client1.ID,
				Status: 1,
			},
			err: nil,
		},
		{
			desc: "change client status for non-existing client",
			client: clients.Client{
				ID:     "invalid",
				Status: 2,
			},
			err: errors.ErrNotFound,
		},
	}

	for _, tc := range ucases {
		expected, err := repo.ChangeStatus(context.Background(), tc.client.ID, tc.client.Status)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.client.Status, expected.Status, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.client.Status, expected.Status))
		}
	}
}

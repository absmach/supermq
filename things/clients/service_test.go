package clients_test

import (
	context "context"
	fmt "fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/internal/apiutil"
	mfclients "github.com/mainflux/mainflux/internal/mainflux/clients"
	"github.com/mainflux/mainflux/internal/testsutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/clients/mocks"
	cmocks "github.com/mainflux/mainflux/things/clients/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	idProvider     = uuid.New()
	secret         = "strongsecret"
	validCMetadata = clients.Metadata{"role": "client"}
	ID             = testsutil.GenerateUUID(&testing.T{}, idProvider)
	client         = clients.Client{
		ID:          ID,
		Name:        "clientname",
		Tags:        []string{"tag1", "tag2"},
		Credentials: clients.Credentials{Identity: "clientidentity", Secret: secret},
		Metadata:    validCMetadata,
		Status:      mfclients.EnabledStatus,
	}
	inValidToken   = "invalidToken"
	withinDuration = 5 * time.Second
	adminEmail     = "admin@example.com"
	token          = "token"
)

func newService(tokens map[string]string) (clients.Service, *cmocks.ClientRepository) {
	adminPolicy := mocks.MockSubjectSet{Object: ID, Relation: clients.AdminRelationKey}
	auth := mocks.NewAuthService(tokens, map[string][]mocks.MockSubjectSet{token: {adminPolicy}})
	thingCache := mocks.NewClientCache()
	idProvider := uuid.NewMock()
	cRepo := new(cmocks.ClientRepository)

	return clients.NewService(auth, cRepo, thingCache, idProvider), cRepo
}

func TestRegisterClient(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	cases := []struct {
		desc   string
		client clients.Client
		token  string
		err    error
	}{
		{
			desc:   "register new client",
			client: client,
			token:  token,
			err:    nil,
		},
		{
			desc:   "register existing client",
			client: client,
			token:  token,
			err:    errors.ErrConflict,
		},
		{
			desc: "register a new enabled client with name",
			client: clients.Client{
				Name: "clientWithName",
				Credentials: clients.Credentials{
					Identity: "newclientwithname@example.com",
					Secret:   secret,
				},
				Status: mfclients.EnabledStatus,
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new disabled client with name",
			client: clients.Client{
				Name: "clientWithName",
				Credentials: clients.Credentials{
					Identity: "newclientwithname@example.com",
					Secret:   secret,
				},
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new enabled client with tags",
			client: clients.Client{
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newclientwithtags@example.com",
					Secret:   secret,
				},
				Status: mfclients.EnabledStatus,
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new disabled client with tags",
			client: clients.Client{
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newclientwithtags@example.com",
					Secret:   secret,
				},
				Status: mfclients.DisabledStatus,
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new enabled client with metadata",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithmetadata@example.com",
					Secret:   secret,
				},
				Metadata: validCMetadata,
				Status:   mfclients.EnabledStatus,
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new disabled client with metadata",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithmetadata@example.com",
					Secret:   secret,
				},
				Metadata: validCMetadata,
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new disabled client",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithvalidstatus@example.com",
					Secret:   secret,
				},
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new client with valid disabled status",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithvalidstatus@example.com",
					Secret:   secret,
				},
				Status: mfclients.DisabledStatus,
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new client with all fields",
			client: clients.Client{
				Name: "newclientwithallfields",
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newclientwithallfields@example.com",
					Secret:   secret,
				},
				Metadata: clients.Metadata{
					"name": "newclientwithallfields",
				},
				Status: mfclients.EnabledStatus,
			},
			err:   nil,
			token: token,
		},
		{
			desc: "register a new client with missing identity",
			client: clients.Client{
				Name: "clientWithMissingIdentity",
				Credentials: clients.Credentials{
					Secret: secret,
				},
			},
			err:   errors.ErrMalformedEntity,
			token: token,
		},
		{
			desc: "register a new client with invalid owner",
			client: clients.Client{
				Owner: mocks.WrongID,
				Credentials: clients.Credentials{
					Identity: "newclientwithinvalidowner@example.com",
					Secret:   secret,
				},
			},
			err:   errors.ErrMalformedEntity,
			token: token,
		},
		{
			desc: "register a new client with empty secret",
			client: clients.Client{
				Owner: testsutil.GenerateUUID(t, idProvider),
				Credentials: clients.Credentials{
					Identity: "newclientwithemptysecret@example.com",
				},
			},
			err:   apiutil.ErrMissingSecret,
			token: token,
		},
		{
			desc: "register a new client with invalid status",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithinvalidstatus@example.com",
					Secret:   secret,
				},
				Status: mfclients.AllStatus,
			},
			err:   apiutil.ErrInvalidStatus,
			token: token,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("Save", context.Background(), mock.Anything).Return(&clients.Client{}, tc.err)
		registerTime := time.Now()
		expected, err := svc.CreateThings(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.NotEmpty(t, expected[0].ID, fmt.Sprintf("%s: expected %s not to be empty\n", tc.desc, expected[0].ID))
			assert.WithinDuration(t, expected[0].CreatedAt, registerTime, withinDuration, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, expected[0].CreatedAt, registerTime))
			tc.client.ID = expected[0].ID
			tc.client.CreatedAt = expected[0].CreatedAt
			tc.client.UpdatedAt = expected[0].UpdatedAt
			tc.client.Credentials.Secret = expected[0].Credentials.Secret
			tc.client.Owner = expected[0].Owner
			tc.client.UpdatedBy = expected[0].UpdatedBy
			assert.Equal(t, tc.client, expected[0], fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.client, expected[0]))
		}
		repoCall.Unset()
	}
}

func TestViewClient(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	cases := []struct {
		desc     string
		token    string
		clientID string
		response clients.Client
		err      error
	}{
		{
			desc:     "view client successfully",
			response: client,
			token:    token,
			clientID: client.ID,
			err:      nil,
		},
		{
			desc:     "view client with an invalid token",
			response: clients.Client{},
			token:    inValidToken,
			clientID: "",
			err:      errors.ErrAuthorization,
		},
		{
			desc:     "view client with valid token and invalid client id",
			response: clients.Client{},
			token:    token,
			clientID: mocks.WrongID,
			err:      errors.ErrNotFound,
		},
		{
			desc:     "view client with an invalid token and invalid client id",
			response: clients.Client{},
			token:    inValidToken,
			clientID: mocks.WrongID,
			err:      errors.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall1 := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.response, tc.err)
		rClient, err := svc.ViewClient(context.Background(), tc.token, tc.clientID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, rClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, rClient))
		repoCall1.Unset()
	}
}

func TestListClients(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	var nClients = uint64(200)
	var aClients = []clients.Client{}
	var OwnerID = testsutil.GenerateUUID(t, idProvider)
	for i := uint64(1); i < nClients; i++ {
		identity := fmt.Sprintf("TestListClients_%d@example.com", i)
		client := clients.Client{
			Name: identity,
			Credentials: clients.Credentials{
				Identity: identity,
				Secret:   "password",
			},
			Tags:     []string{"tag1", "tag2"},
			Metadata: clients.Metadata{"role": "client"},
		}
		if i%50 == 0 {
			client.Owner = OwnerID
			client.Owner = testsutil.GenerateUUID(t, idProvider)
		}
		aClients = append(aClients, client)
	}

	cases := []struct {
		desc     string
		token    string
		page     clients.Page
		response clients.ClientsPage
		size     uint64
		err      error
	}{
		{
			desc:  "list clients with authorized token",
			token: token,

			page: clients.Page{
				Status: mfclients.AllStatus,
			},
			size: 0,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{},
			},
			err: nil,
		},
		{
			desc:  "list clients with an invalid token",
			token: inValidToken,
			page: clients.Page{
				Status: mfclients.AllStatus,
			},
			size: 0,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
			},
			err: errors.ErrAuthentication,
		},
		{
			desc:  "list clients that are shared with me",
			token: token,
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Status:   mfclients.EnabledStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  4,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{aClients[0], aClients[50], aClients[100], aClients[150]},
			},
			size: 4,
		},
		{
			desc:  "list clients that are shared with me with a specific name",
			token: token,
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Name:     "TestListClients3",
				Status:   mfclients.EnabledStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  4,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{aClients[0], aClients[50], aClients[100], aClients[150]},
			},
			size: 4,
		},
		{
			desc:  "list clients that are shared with me with an invalid name",
			token: token,
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Name:     "notpresentclient",
				Status:   mfclients.EnabledStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{},
			},
			size: 0,
		},
		{
			desc:  "list clients that I own",
			token: token,
			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Owner:  clients.MyKey,
				Status: mfclients.EnabledStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  4,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{aClients[0], aClients[50], aClients[100], aClients[150]},
			},
			size: 4,
		},
		{
			desc:  "list clients that I own with a specific name",
			token: token,
			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Owner:  clients.MyKey,
				Name:   "TestListClients3",
				Status: mfclients.AllStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  4,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{aClients[0], aClients[50], aClients[100], aClients[150]},
			},
			size: 4,
		},
		{
			desc:  "list clients that I own with an invalid name",
			token: token,
			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Owner:  clients.MyKey,
				Name:   "notpresentclient",
				Status: mfclients.AllStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{},
			},
			size: 0,
		},
		{
			desc:  "list clients that I own and are shared with me",
			token: token,
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				Owner:    clients.MyKey,
				SharedBy: clients.MyKey,
				Status:   mfclients.AllStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  4,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{aClients[0], aClients[50], aClients[100], aClients[150]},
			},
			size: 4,
		},
		{
			desc:  "list clients that I own and are shared with me with a specific name",
			token: token,
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Owner:    clients.MyKey,
				Name:     "TestListClients3",
				Status:   mfclients.AllStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  4,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{aClients[0], aClients[50], aClients[100], aClients[150]},
			},
			size: 4,
		},
		{
			desc:  "list clients that I own and are shared with me with an invalid name",
			token: token,
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Owner:    clients.MyKey,
				Name:     "notpresentclient",
				Status:   mfclients.AllStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
				Clients: []clients.Client{},
			},
			size: 0,
		},
		{
			desc:  "list clients with offset and limit",
			token: token,

			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Status: mfclients.AllStatus,
			},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  nClients - 6,
					Offset: 0,
					Limit:  0,
				},
				Clients: aClients[6:nClients],
			},
			size: nClients - 6,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(clients.Client{}, tc.err)
		repoCall1 := cRepo.On("RetrieveAll", context.Background(), mock.Anything).Return(tc.response, tc.err)
		page, err := svc.ListClients(context.Background(), tc.token, tc.page)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, page, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, page))
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestUpdateClient(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	client1 := client
	client2 := client
	client1.Name = "Updated client"
	client2.Metadata = clients.Metadata{"role": "test"}

	cases := []struct {
		desc     string
		client   clients.Client
		response clients.Client
		token    string
		err      error
	}{
		{
			desc:     "update client name with valid token",
			client:   client1,
			response: client1,
			token:    token,
			err:      nil,
		},
		{
			desc:     "update client name with invalid token",
			client:   client1,
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthorization,
		},
		{
			desc: "update client name with invalid ID",
			client: clients.Client{
				ID:   mocks.WrongID,
				Name: "Updated Client",
			},
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthorization,
		},
		{
			desc:     "update client metadata with valid token",
			client:   client2,
			response: client2,
			token:    token,
			err:      nil,
		},
		{
			desc:     "update client metadata with invalid token",
			client:   client2,
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(clients.Client{}, tc.err)
		repoCall1 := cRepo.On("Update", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClient(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		repoCall1.Unset()
		repoCall.Unset()
	}
}

func TestUpdateClientTags(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	client.Tags = []string{"updated"}

	cases := []struct {
		desc     string
		client   clients.Client
		response clients.Client
		token    string
		err      error
	}{
		{
			desc:     "update client tags with valid token",
			client:   client,
			token:    token,
			response: client,
			err:      nil,
		},
		{
			desc:     "update client tags with invalid token",
			client:   client,
			token:    "non-existent",
			response: clients.Client{},
			err:      errors.ErrAuthorization,
		},
		{
			desc: "update client name with invalid ID",
			client: clients.Client{
				ID:   mocks.WrongID,
				Name: "Updated name",
			},
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(clients.Client{}, tc.err)
		repoCall1 := cRepo.On("UpdateTags", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClientTags(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		repoCall1.Unset()
		repoCall.Unset()
	}
}

func TestUpdateClientOwner(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	client.Owner = "newowner@mail.com"

	cases := []struct {
		desc     string
		client   clients.Client
		response clients.Client
		token    string
		err      error
	}{
		{
			desc:     "update client owner with valid token",
			client:   client,
			token:    token,
			response: client,
			err:      nil,
		},
		{
			desc:     "update client owner with invalid token",
			client:   client,
			token:    "non-existent",
			response: clients.Client{},
			err:      errors.ErrAuthorization,
		},
		{
			desc: "update client owner with invalid ID",
			client: clients.Client{
				ID:    mocks.WrongID,
				Owner: "updatedowner@mail.com",
			},
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(clients.Client{}, tc.err)
		repoCall1 := cRepo.On("UpdateOwner", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClientOwner(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		repoCall1.Unset()
		repoCall.Unset()
	}
}

func TestUpdateClientSecret(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	cases := []struct {
		desc      string
		id        string
		newSecret string
		token     string
		response  clients.Client
		err       error
	}{
		{
			desc:      "update client secret with valid token",
			id:        client.ID,
			newSecret: "newSecret",
			token:     token,
			response:  client,
			err:       nil,
		},
		{
			desc:      "update client secret with invalid token",
			id:        client.ID,
			newSecret: "newPassword",
			token:     "non-existent",
			response:  clients.Client{},
			err:       errors.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall1 := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.response, tc.err)
		repoCall2 := cRepo.On("RetrieveByIdentity", context.Background(), mock.Anything).Return(tc.response, tc.err)
		repoCall3 := cRepo.On("UpdateSecret", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClientSecret(context.Background(), tc.token, tc.id, tc.newSecret)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		repoCall1.Unset()
		repoCall2.Unset()
		repoCall3.Unset()
	}
}

func TestEnableClient(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	enabledClient1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "client1@example.com", Secret: "password"}, Status: mfclients.EnabledStatus}
	disabledClient1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "client3@example.com", Secret: "password"}, Status: mfclients.DisabledStatus}
	endisabledClient1 := disabledClient1
	endisabledClient1.Status = mfclients.EnabledStatus

	cases := []struct {
		desc     string
		id       string
		token    string
		client   clients.Client
		response clients.Client
		err      error
	}{
		{
			desc:     "enable disabled client",
			id:       disabledClient1.ID,
			token:    token,
			client:   disabledClient1,
			response: endisabledClient1,
			err:      nil,
		},
		{
			desc:     "enable enabled client",
			id:       enabledClient1.ID,
			token:    token,
			client:   enabledClient1,
			response: enabledClient1,
			err:      mfclients.ErrStatusAlreadyAssigned,
		},
		{
			desc:     "enable non-existing client",
			id:       mocks.WrongID,
			token:    token,
			client:   clients.Client{},
			response: clients.Client{},
			err:      errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		repoCall1 := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.client, tc.err)
		repoCall2 := cRepo.On("ChangeStatus", context.Background(), mock.Anything).Return(tc.response, tc.err)
		_, err := svc.EnableClient(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall1.Unset()
		repoCall2.Unset()
	}

	cases2 := []struct {
		desc     string
		status   mfclients.Status
		size     uint64
		response clients.ClientsPage
	}{
		{
			desc:   "list enabled clients",
			status: mfclients.EnabledStatus,
			size:   2,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{enabledClient1, endisabledClient1},
			},
		},
		{
			desc:   "list disabled clients",
			status: mfclients.DisabledStatus,
			size:   1,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  1,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{disabledClient1},
			},
		},
		{
			desc:   "list enabled and disabled clients",
			status: mfclients.AllStatus,
			size:   3,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  3,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{enabledClient1, disabledClient1, endisabledClient1},
			},
		},
	}

	for _, tc := range cases2 {
		pm := clients.Page{
			Offset: 0,
			Limit:  100,
			Status: tc.status,
		}
		repoCall := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(clients.Client{}, nil)
		repoCall1 := cRepo.On("RetrieveAll", context.Background(), mock.Anything).Return(tc.response, nil)
		page, err := svc.ListClients(context.Background(), token, pm)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		size := uint64(len(page.Clients))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", tc.desc, tc.size, size))
		repoCall1.Unset()
		repoCall.Unset()
	}
}

func TestDisableClient(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	enabledClient1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "client1@example.com", Secret: "password"}, Status: mfclients.EnabledStatus}
	disabledClient1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "client3@example.com", Secret: "password"}, Status: mfclients.DisabledStatus}
	disenabledClient1 := enabledClient1
	disenabledClient1.Status = mfclients.DisabledStatus

	cases := []struct {
		desc     string
		id       string
		token    string
		client   clients.Client
		response clients.Client
		err      error
	}{
		{
			desc:     "disable enabled client",
			id:       enabledClient1.ID,
			token:    token,
			client:   enabledClient1,
			response: disenabledClient1,
			err:      nil,
		},
		{
			desc:     "disable disabled client",
			id:       disabledClient1.ID,
			token:    token,
			client:   disabledClient1,
			response: clients.Client{},
			err:      mfclients.ErrStatusAlreadyAssigned,
		},
		{
			desc:     "disable non-existing client",
			id:       mocks.WrongID,
			client:   clients.Client{},
			token:    token,
			response: clients.Client{},
			err:      errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		_ = cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.client, tc.err)
		repoCall1 := cRepo.On("ChangeStatus", context.Background(), mock.Anything).Return(tc.response, tc.err)
		_, err := svc.DisableClient(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall1.Unset()
	}

	cases2 := []struct {
		desc     string
		status   mfclients.Status
		size     uint64
		response clients.ClientsPage
	}{
		{
			desc:   "list enabled clients",
			status: mfclients.EnabledStatus,
			size:   1,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  1,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{enabledClient1},
			},
		},
		{
			desc:   "list disabled clients",
			status: mfclients.DisabledStatus,
			size:   2,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{disenabledClient1, disabledClient1},
			},
		},
		{
			desc:   "list enabled and disabled clients",
			status: mfclients.AllStatus,
			size:   3,
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  3,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{enabledClient1, disabledClient1, disenabledClient1},
			},
		},
	}

	for _, tc := range cases2 {
		pm := clients.Page{
			Offset: 0,
			Limit:  100,
			Status: tc.status,
		}
		repoCall := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(clients.Client{}, nil)
		repoCall1 := cRepo.On("RetrieveAll", context.Background(), mock.Anything).Return(tc.response, nil)
		page, err := svc.ListClients(context.Background(), token, pm)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		size := uint64(len(page.Clients))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", tc.desc, tc.size, size))
		repoCall1.Unset()
		repoCall.Unset()
	}
}

func TestListMembers(t *testing.T) {
	svc, cRepo := newService(map[string]string{token: adminEmail})

	var nClients = uint64(10)
	var aClients = []clients.Client{}
	for i := uint64(1); i < nClients; i++ {
		identity := fmt.Sprintf("member_%d@example.com", i)
		client := clients.Client{
			ID:   testsutil.GenerateUUID(t, idProvider),
			Name: identity,
			Credentials: clients.Credentials{
				Identity: identity,
				Secret:   "password",
			},
			Tags:     []string{"tag1", "tag2"},
			Metadata: clients.Metadata{"role": "client"},
		}
		aClients = append(aClients, client)
	}
	validToken := token

	cases := []struct {
		desc     string
		token    string
		groupID  string
		page     clients.Page
		response clients.MembersPage
		err      error
	}{
		{
			desc:    "list clients with authorized token",
			token:   validToken,
			groupID: testsutil.GenerateUUID(t, idProvider),
			page: clients.Page{
				Subject: adminEmail,
				Owner:   adminEmail,
				Action:  "g_list",
			},
			response: clients.MembersPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
				Members: []clients.Client{},
			},
			err: nil,
		},
		{
			desc:    "list clients with offset and limit",
			token:   validToken,
			groupID: testsutil.GenerateUUID(t, idProvider),
			page: clients.Page{
				Offset:  6,
				Limit:   nClients,
				Status:  mfclients.AllStatus,
				Subject: adminEmail,
				Owner:   adminEmail,
				Action:  "g_list",
			},
			response: clients.MembersPage{
				Page: clients.Page{
					Total: nClients - 6 - 1,
				},
				Members: aClients[6 : nClients-1],
			},
		},
		{
			desc:    "list clients with an invalid token",
			token:   inValidToken,
			groupID: testsutil.GenerateUUID(t, idProvider),
			page: clients.Page{
				Subject: adminEmail,
				Action:  "g_list",
				Owner:   adminEmail,
			},
			response: clients.MembersPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
			},
			err: errors.ErrAuthentication,
		},
		{
			desc:    "list clients with an invalid id",
			token:   validToken,
			groupID: mocks.WrongID,
			page: clients.Page{
				Subject: adminEmail,
				Action:  "g_list",
				Owner:   adminEmail,
			},
			response: clients.MembersPage{
				Page: clients.Page{
					Total:  0,
					Offset: 0,
					Limit:  0,
				},
			},
			err: errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(clients.Client{}, tc.err)
		repoCall1 := cRepo.On("Members", context.Background(), tc.groupID, tc.page).Return(tc.response, tc.err)
		page, err := svc.ListClientsByGroup(context.Background(), tc.token, tc.groupID, tc.page)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, page, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, page))
		repoCall.Unset()
		repoCall1.Unset()
	}
}

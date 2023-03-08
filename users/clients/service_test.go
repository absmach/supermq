package clients_test

import (
	context "context"
	fmt "fmt"
	"regexp"
	"testing"
	"time"

	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/internal/testsutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users/clients"
	"github.com/mainflux/mainflux/users/clients/mocks"
	cmocks "github.com/mainflux/mainflux/users/clients/mocks"
	"github.com/mainflux/mainflux/users/hasher"
	"github.com/mainflux/mainflux/users/jwt"
	pmocks "github.com/mainflux/mainflux/users/policies/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	idProvider     = uuid.New()
	phasher        = hasher.New()
	secret         = "strongsecret"
	validCMetadata = clients.Metadata{"role": "client"}
	client         = clients.Client{
		ID:          testsutil.GenerateUUID(&testing.T{}, idProvider),
		Name:        "clientname",
		Tags:        []string{"tag1", "tag2"},
		Credentials: clients.Credentials{Identity: "clientidentity", Secret: secret},
		Metadata:    validCMetadata,
		Status:      clients.EnabledStatus,
	}
	inValidToken    = "invalidToken"
	withinDuration  = 5 * time.Second
	passRegex       = regexp.MustCompile("^.{8,}$")
	accessDuration  = time.Minute * 1
	refreshDuration = time.Minute * 10
)

func generateValidToken(t *testing.T, clientID string, svc clients.Service, cRepo *mocks.ClientRepository) string {
	client := clients.Client{
		ID:   clientID,
		Name: "validtoken",
		Credentials: clients.Credentials{
			Identity: "validtoken",
			Secret:   secret,
		},
		Status: clients.EnabledStatus,
	}
	rClient := client
	rClient.Credentials.Secret, _ = phasher.Hash(client.Credentials.Secret)

	repoCall := cRepo.On("RetrieveByIdentity", context.Background(), client.Credentials.Identity).Return(rClient, nil)
	token, err := svc.IssueToken(context.Background(), client.Credentials.Identity, client.Credentials.Secret)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("Create token expected nil got %s\n", err))
	if !repoCall.Parent.AssertCalled(t, "RetrieveByIdentity", context.Background(), client.Credentials.Identity) {
		assert.Fail(t, "RetrieveByIdentity was not called on creating token")
	}
	repoCall.Unset()
	return token.AccessToken
}

func TestRegisterClient(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

	cases := []struct {
		desc   string
		client clients.Client
		token  string
		err    error
	}{
		{
			desc:   "register new client",
			client: client,
			token:  generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			err:    nil,
		},
		{
			desc:   "register existing client",
			client: client,
			token:  generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
				Status: clients.EnabledStatus,
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
		},
		{
			desc: "register a new enabled client with tags",
			client: clients.Client{
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newclientwithtags@example.com",
					Secret:   secret,
				},
				Status: clients.EnabledStatus,
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
		},
		{
			desc: "register a new disabled client with tags",
			client: clients.Client{
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newclientwithtags@example.com",
					Secret:   secret,
				},
				Status: clients.DisabledStatus,
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
		},
		{
			desc: "register a new enabled client with metadata",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithmetadata@example.com",
					Secret:   secret,
				},
				Metadata: validCMetadata,
				Status:   clients.EnabledStatus,
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
		},
		{
			desc: "register a new client with valid disabled status",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithvalidstatus@example.com",
					Secret:   secret,
				},
				Status: clients.DisabledStatus,
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
				Status: clients.EnabledStatus,
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
		},
		{
			desc: "register a new client with invalid status",
			client: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newclientwithinvalidstatus@example.com",
					Secret:   secret,
				},
				Status: clients.AllStatus,
			},
			err:   apiutil.ErrInvalidStatus,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("Save", context.Background(), mock.Anything).Return(&clients.Client{}, tc.err)
		registerTime := time.Now()
		expected, err := svc.RegisterClient(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.NotEmpty(t, expected.ID, fmt.Sprintf("%s: expected %s not to be empty\n", tc.desc, expected.ID))
			assert.WithinDuration(t, expected.CreatedAt, registerTime, withinDuration, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, expected.CreatedAt, registerTime))
			tc.client.ID = expected.ID
			tc.client.CreatedAt = expected.CreatedAt
			tc.client.UpdatedAt = expected.UpdatedAt
			tc.client.Credentials.Secret = expected.Credentials.Secret
			tc.client.Owner = expected.Owner
			assert.Equal(t, tc.client, expected, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.client, expected))
			if !repoCall.Parent.AssertCalled(t, "Save", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("Save was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
	}
}

func TestViewClient(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

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
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			clientID: client.ID,
			err:      nil,
		},
		{
			desc:     "view client with an invalid token",
			response: clients.Client{},
			token:    inValidToken,
			clientID: "",
			err:      errors.ErrAuthentication,
		},
		{
			desc:     "view client with valid token and invalid client id",
			response: clients.Client{},
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			clientID: mocks.WrongID,
			err:      errors.ErrNotFound,
		},
		{
			desc:     "view client with an invalid token and invalid client id",
			response: clients.Client{},
			token:    inValidToken,
			clientID: mocks.WrongID,
			err:      errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("Evaluate", context.Background(), "client", mock.Anything).Return(nil)
		repoCall1 := cRepo.On("RetrieveByID", context.Background(), tc.clientID).Return(tc.response, tc.err)
		rClient, err := svc.ViewClient(context.Background(), tc.token, tc.clientID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, rClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, rClient))
		repoCall.Unset()
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "Evaluate", context.Background(), "client", mock.Anything) {
				assert.Fail(t, fmt.Sprintf("Evaluate was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "RetrieveByID", context.Background(), tc.clientID) {
				assert.Fail(t, fmt.Sprintf("RetrieveByID was not called on %s", tc.desc))
			}
		}
		repoCall1.Unset()
	}
}

func TestListClients(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),

			page: clients.Page{
				Status: clients.AllStatus,
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
				Status: clients.AllStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Status:   clients.EnabledStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Name:     "TestListClients3",
				Status:   clients.EnabledStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Name:     "notpresentclient",
				Status:   clients.EnabledStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Owner:  clients.MyKey,
				Status: clients.EnabledStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Owner:  clients.MyKey,
				Name:   "TestListClients3",
				Status: clients.AllStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Owner:  clients.MyKey,
				Name:   "notpresentclient",
				Status: clients.AllStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				Owner:    clients.MyKey,
				SharedBy: clients.MyKey,
				Status:   clients.AllStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Owner:    clients.MyKey,
				Name:     "TestListClients3",
				Status:   clients.AllStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			page: clients.Page{
				Offset:   6,
				Limit:    nClients,
				SharedBy: clients.MyKey,
				Owner:    clients.MyKey,
				Name:     "notpresentclient",
				Status:   clients.AllStatus,
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
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),

			page: clients.Page{
				Offset: 6,
				Limit:  nClients,
				Status: clients.AllStatus,
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
		repoCall := cRepo.On("RetrieveAll", context.Background(), mock.Anything).Return(tc.response, tc.err)
		page, err := svc.ListClients(context.Background(), tc.token, tc.page)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, page, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, page))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "RetrieveAll", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("RetrieveAll was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
	}
}

func TestUpdateClient(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

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
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			err:      nil,
		},
		{
			desc:     "update client name with invalid token",
			client:   client1,
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthentication,
		},
		{
			desc: "update client name with invalid ID",
			client: clients.Client{
				ID:   mocks.WrongID,
				Name: "Updated Client",
			},
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthentication,
		},
		{
			desc:     "update client metadata with valid token",
			client:   client2,
			response: client2,
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			err:      nil,
		},
		{
			desc:     "update client metadata with invalid token",
			client:   client2,
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("CheckAdmin", context.Background(), mock.Anything).Return(nil)
		repoCall1 := cRepo.On("Update", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClient(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "CheckAdmin", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("CheckAdmin was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "Update", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("Update was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestUpdateClientTags(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

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
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			response: client,
			err:      nil,
		},
		{
			desc:     "update client tags with invalid token",
			client:   client,
			token:    "non-existent",
			response: clients.Client{},
			err:      errors.ErrAuthentication,
		},
		{
			desc: "update client name with invalid ID",
			client: clients.Client{
				ID:   mocks.WrongID,
				Name: "Updated name",
			},
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("CheckAdmin", context.Background(), mock.Anything).Return(nil)
		repoCall1 := cRepo.On("UpdateTags", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClientTags(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "CheckAdmin", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("CheckAdmin was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "UpdateTags", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("UpdateTags was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestUpdateClientIdentity(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

	client2 := client
	client2.Credentials.Identity = "updated@example.com"

	cases := []struct {
		desc     string
		identity string
		response clients.Client
		token    string
		id       string
		err      error
	}{
		{
			desc:     "update client identity with valid token",
			identity: "updated@example.com",
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			id:       client.ID,
			response: client2,
			err:      nil,
		},
		{
			desc:     "update client identity with invalid id",
			identity: "updated@example.com",
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			id:       mocks.WrongID,
			response: clients.Client{},
			err:      errors.ErrNotFound,
		},
		{
			desc:     "update client identity with invalid token",
			identity: "updated@example.com",
			token:    "non-existent",
			id:       client2.ID,
			response: clients.Client{},
			err:      errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("CheckAdmin", context.Background(), mock.Anything).Return(nil)
		repoCall1 := cRepo.On("UpdateIdentity", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClientIdentity(context.Background(), tc.token, tc.id, tc.identity)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "CheckAdmin", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("CheckAdmin was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "UpdateIdentity", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("UpdateIdentity was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestUpdateClientOwner(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

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
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			response: client,
			err:      nil,
		},
		{
			desc:     "update client owner with invalid token",
			client:   client,
			token:    "non-existent",
			response: clients.Client{},
			err:      errors.ErrAuthentication,
		},
		{
			desc: "update client owner with invalid ID",
			client: clients.Client{
				ID:    mocks.WrongID,
				Owner: "updatedowner@mail.com",
			},
			response: clients.Client{},
			token:    "non-existent",
			err:      errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("CheckAdmin", context.Background(), mock.Anything).Return(nil)
		repoCall1 := cRepo.On("UpdateOwner", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClientOwner(context.Background(), tc.token, tc.client)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "CheckAdmin", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("CheckAdmin was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "UpdateOwner", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("UpdateOwner was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestUpdateClientSecret(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

	rClient := client
	rClient.Credentials.Secret, _ = phasher.Hash(client.Credentials.Secret)

	repoCall := cRepo.On("RetrieveByIdentity", context.Background(), client.Credentials.Identity).Return(rClient, nil)
	token, err := svc.IssueToken(context.Background(), client.Credentials.Identity, client.Credentials.Secret)
	assert.Nil(t, err, fmt.Sprintf("Issue token expected nil got %s\n", err))
	repoCall.Unset()

	cases := []struct {
		desc      string
		oldSecret string
		newSecret string
		token     string
		response  clients.Client
		err       error
	}{
		{
			desc:      "update client secret with valid token",
			oldSecret: client.Credentials.Secret,
			newSecret: "newSecret",
			token:     token.AccessToken,
			response:  rClient,
			err:       nil,
		},
		{
			desc:      "update client secret with invalid token",
			oldSecret: client.Credentials.Secret,
			newSecret: "newPassword",
			token:     "non-existent",
			response:  clients.Client{},
			err:       errors.ErrAuthentication,
		},
		{
			desc:      "update client secret with wrong old secret",
			oldSecret: "oldSecret",
			newSecret: "newSecret",
			token:     token.AccessToken,
			response:  clients.Client{},
			err:       apiutil.ErrInvalidSecret,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("RetrieveByID", context.Background(), client.ID).Return(tc.response, tc.err)
		repoCall1 := cRepo.On("RetrieveByIdentity", context.Background(), client.Credentials.Identity).Return(tc.response, tc.err)
		repoCall2 := cRepo.On("UpdateSecret", context.Background(), mock.Anything).Return(tc.response, tc.err)
		updatedClient, err := svc.UpdateClientSecret(context.Background(), tc.token, tc.oldSecret, tc.newSecret)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, updatedClient, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, updatedClient))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "RetrieveByID", context.Background(), tc.response.ID) {
				assert.Fail(t, fmt.Sprintf("RetrieveByID was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "RetrieveByIdentity", context.Background(), tc.response.Credentials.Identity) {
				assert.Fail(t, fmt.Sprintf("RetrieveByIdentity was not called on %s", tc.desc))
			}
			if !repoCall2.Parent.AssertCalled(t, "UpdateSecret", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("UpdateSecret was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
	}
}

func TestEnableClient(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

	enabledClient1 := clients.Client{ID: testsutil.GenerateUUID(t, idProvider), Credentials: clients.Credentials{Identity: "client1@example.com", Secret: "password"}, Status: clients.EnabledStatus}
	disabledClient1 := clients.Client{ID: testsutil.GenerateUUID(t, idProvider), Credentials: clients.Credentials{Identity: "client3@example.com", Secret: "password"}, Status: clients.DisabledStatus}
	endisabledClient1 := disabledClient1
	endisabledClient1.Status = clients.EnabledStatus

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
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			client:   disabledClient1,
			response: endisabledClient1,
			err:      nil,
		},
		{
			desc:     "enable enabled client",
			id:       enabledClient1.ID,
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			client:   enabledClient1,
			response: enabledClient1,
			err:      clients.ErrStatusAlreadyAssigned,
		},
		{
			desc:     "enable non-existing client",
			id:       mocks.WrongID,
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			client:   clients.Client{},
			response: clients.Client{},
			err:      errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("CheckAdmin", context.Background(), mock.Anything).Return(nil)
		repoCall1 := cRepo.On("RetrieveByID", context.Background(), tc.id).Return(tc.client, tc.err)
		repoCall2 := cRepo.On("ChangeStatus", context.Background(), tc.id, clients.EnabledStatus).Return(tc.response, tc.err)
		_, err := svc.EnableClient(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "CheckAdmin", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("CheckAdmin was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "RetrieveByID", context.Background(), tc.id) {
				assert.Fail(t, fmt.Sprintf("RetrieveByID was not called on %s", tc.desc))
			}
			if !repoCall2.Parent.AssertCalled(t, "ChangeStatus", context.Background(), tc.id, clients.EnabledStatus) {
				assert.Fail(t, fmt.Sprintf("ChangeStatus was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
	}

	cases2 := []struct {
		desc     string
		status   clients.Status
		size     uint64
		response clients.ClientsPage
	}{
		{
			desc:   "list enabled clients",
			status: clients.EnabledStatus,
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
			status: clients.DisabledStatus,
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
			status: clients.AllStatus,
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
			Action: "c_list",
		}
		repoCall := cRepo.On("RetrieveAll", context.Background(), pm).Return(tc.response, nil)
		page, err := svc.ListClients(context.Background(), generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo), pm)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		size := uint64(len(page.Clients))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", tc.desc, tc.size, size))
		repoCall.Unset()
	}
}

func TestDisableClient(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

	enabledClient1 := clients.Client{ID: testsutil.GenerateUUID(t, idProvider), Credentials: clients.Credentials{Identity: "client1@example.com", Secret: "password"}, Status: clients.EnabledStatus}
	disabledClient1 := clients.Client{ID: testsutil.GenerateUUID(t, idProvider), Credentials: clients.Credentials{Identity: "client3@example.com", Secret: "password"}, Status: clients.DisabledStatus}
	disenabledClient1 := enabledClient1
	disenabledClient1.Status = clients.DisabledStatus

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
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			client:   enabledClient1,
			response: disenabledClient1,
			err:      nil,
		},
		{
			desc:     "disable disabled client",
			id:       disabledClient1.ID,
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			client:   disabledClient1,
			response: clients.Client{},
			err:      clients.ErrStatusAlreadyAssigned,
		},
		{
			desc:     "disable non-existing client",
			id:       mocks.WrongID,
			client:   clients.Client{},
			token:    generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			response: clients.Client{},
			err:      errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("CheckAdmin", context.Background(), mock.Anything).Return(nil)
		repoCall1 := cRepo.On("RetrieveByID", context.Background(), tc.id).Return(tc.client, tc.err)
		repoCall2 := cRepo.On("ChangeStatus", context.Background(), tc.id, clients.DisabledStatus).Return(tc.response, tc.err)
		_, err := svc.DisableClient(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "CheckAdmin", context.Background(), mock.Anything) {
				assert.Fail(t, fmt.Sprintf("CheckAdmin was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "RetrieveByID", context.Background(), tc.id) {
				assert.Fail(t, fmt.Sprintf("RetrieveByID was not called on %s", tc.desc))
			}
			if !repoCall2.Parent.AssertCalled(t, "ChangeStatus", context.Background(), tc.id, clients.DisabledStatus) {
				assert.Fail(t, fmt.Sprintf("ChangeStatus was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
	}

	cases2 := []struct {
		desc     string
		status   clients.Status
		size     uint64
		response clients.ClientsPage
	}{
		{
			desc:   "list enabled clients",
			status: clients.EnabledStatus,
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
			status: clients.DisabledStatus,
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
			status: clients.AllStatus,
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
			Action: "c_list",
		}
		repoCall := cRepo.On("RetrieveAll", context.Background(), pm).Return(tc.response, nil)
		page, err := svc.ListClients(context.Background(), generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo), pm)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		size := uint64(len(page.Clients))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", tc.desc, tc.size, size))
		repoCall.Unset()
	}
}

func TestListMembers(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

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
	validID := testsutil.GenerateUUID(t, idProvider)
	validToken := generateValidToken(t, validID, svc, cRepo)

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
				Subject: validID,
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
				Status:  clients.AllStatus,
				Subject: validID,
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
				Subject: validID,
				Action:  "g_list",
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
				Subject: validID,
				Action:  "g_list",
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
		repoCall := pRepo.On("CheckAdmin", context.Background(), validID).Return(nil)
		repoCall1 := cRepo.On("Members", context.Background(), tc.groupID, tc.page).Return(tc.response, tc.err)
		page, err := svc.ListMembers(context.Background(), tc.token, tc.groupID, tc.page)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, page, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, page))
		if tc.err == nil {
			if !repoCall.Parent.AssertCalled(t, "CheckAdmin", context.Background(), validID) {
				assert.Fail(t, fmt.Sprintf("CheckAdmin was not called on %s", tc.desc))
			}
			if !repoCall1.Parent.AssertCalled(t, "Members", context.Background(), tc.groupID, tc.page) {
				assert.Fail(t, fmt.Sprintf("Members was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestIssueToken(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

	rClient := client
	rClient2 := client
	rClient3 := client
	rClient.Credentials.Secret, _ = phasher.Hash(client.Credentials.Secret)
	rClient2.Credentials.Secret = "wrongsecret"
	rClient3.Credentials.Secret, _ = phasher.Hash("wrongsecret")

	cases := []struct {
		desc    string
		client  clients.Client
		rClient clients.Client
		err     error
	}{
		{
			desc:    "issue token for an existing client",
			client:  client,
			rClient: rClient,
			err:     nil,
		},
		{
			desc:    "issue token for a non-existing client",
			client:  client,
			rClient: clients.Client{},
			err:     errors.ErrAuthentication,
		},
		{
			desc:    "issue token for a client with wrong secret",
			client:  rClient2,
			rClient: rClient3,
			err:     errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall := cRepo.On("RetrieveByIdentity", context.Background(), tc.client.Credentials.Identity).Return(tc.rClient, tc.err)
		token, err := svc.IssueToken(context.Background(), tc.client.Credentials.Identity, tc.client.Credentials.Secret)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.NotEmpty(t, token.AccessToken, fmt.Sprintf("%s: expected %s not to be empty\n", tc.desc, token.AccessToken))
			assert.NotEmpty(t, token.RefreshToken, fmt.Sprintf("%s: expected %s not to be empty\n", tc.desc, token.RefreshToken))
			if !repoCall.Parent.AssertCalled(t, "RetrieveByIdentity", context.Background(), tc.client.Credentials.Identity) {
				assert.Fail(t, fmt.Sprintf("RetrieveByIdentity was not called on %s", tc.desc))
			}
		}
		repoCall.Unset()
	}
}

func TestRefreshToken(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret), accessDuration, refreshDuration)
	e := mocks.NewEmailer()
	svc := clients.NewService(cRepo, pRepo, tokenizer, e, phasher, idProvider, passRegex)

	rClient := client
	rClient.Credentials.Secret, _ = phasher.Hash(client.Credentials.Secret)

	repoCall := cRepo.On("RetrieveByIdentity", context.Background(), client.Credentials.Identity).Return(rClient, nil)
	token, err := svc.IssueToken(context.Background(), client.Credentials.Identity, client.Credentials.Secret)
	assert.Nil(t, err, fmt.Sprintf("Issue token expected nil got %s\n", err))
	repoCall.Unset()

	cases := []struct {
		desc   string
		token  string
		client clients.Client
		err    error
	}{
		{
			desc:   "refresh token with refresh token for an existing client",
			token:  token.RefreshToken,
			client: client,
			err:    nil,
		},
		{
			desc:   "refresh token with refresh token for a non-existing client",
			token:  token.RefreshToken,
			client: clients.Client{},
			err:    errors.ErrAuthentication,
		},
		{
			desc:   "refresh token with access token for an existing client",
			token:  token.AccessToken,
			client: client,
			err:    errors.ErrAuthentication,
		},
		{
			desc:   "refresh token with access token for a non-existing client",
			token:  token.AccessToken,
			client: clients.Client{},
			err:    errors.ErrAuthentication,
		},
		{
			desc:   "refresh token with invalid token for an existing client",
			token:  generateValidToken(t, testsutil.GenerateUUID(t, idProvider), svc, cRepo),
			client: client,
			err:    errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall1 := cRepo.On("RetrieveByIdentity", context.Background(), tc.client.Credentials.Identity).Return(tc.client, nil)
		repoCall2 := cRepo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.client, tc.err)
		token, err := svc.RefreshToken(context.Background(), tc.token)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.NotEmpty(t, token.AccessToken, fmt.Sprintf("%s: expected %s not to be empty\n", tc.desc, token.AccessToken))
			assert.NotEmpty(t, token.RefreshToken, fmt.Sprintf("%s: expected %s not to be empty\n", tc.desc, token.RefreshToken))
			if !repoCall1.Parent.AssertCalled(t, "RetrieveByIdentity", context.Background(), tc.client.Credentials.Identity) {
				assert.Fail(t, fmt.Sprintf("RetrieveByIdentity was not called on %s", tc.desc))
			}
			if !repoCall2.Parent.AssertCalled(t, "RetrieveByID", context.Background(), tc.client.ID) {
				assert.Fail(t, fmt.Sprintf("RetrieveByID was not called on %s", tc.desc))
			}
		}
		repoCall1.Unset()
		repoCall2.Unset()
	}
}

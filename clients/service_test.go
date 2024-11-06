// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package clients_test

import (
	"context"
	"fmt"
	"testing"

	chmocks "github.com/absmach/magistrala/channels/mocks"
	"github.com/absmach/magistrala/clients"
	climocks "github.com/absmach/magistrala/clients/mocks"
	gpmocks "github.com/absmach/magistrala/groups/mocks"
	"github.com/absmach/magistrala/internal/testsutil"
	mgauthn "github.com/absmach/magistrala/pkg/authn"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	policysvc "github.com/absmach/magistrala/pkg/policies"
	policymocks "github.com/absmach/magistrala/pkg/policies/mocks"
	"github.com/absmach/magistrala/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	secret         = "strongsecret"
	validTMetadata = clients.Metadata{"role": "thing"}
	ID             = "6e5e10b3-d4df-4758-b426-4929d55ad740"
	thing          = clients.Client{
		ID:          ID,
		Name:        "thingname",
		Tags:        []string{"tag1", "tag2"},
		Credentials: clients.Credentials{Identity: "thingidentity", Secret: secret},
		Metadata:    validTMetadata,
		Status:      clients.EnabledStatus,
	}
	validToken        = "token"
	valid             = "valid"
	invalid           = "invalid"
	validID           = "d4ebb847-5d0e-4e46-bdd9-b6aceaaa3a22"
	wrongID           = testsutil.GenerateUUID(&testing.T{})
	errRemovePolicies = errors.New("failed to delete policies")
)

var (
	pService   *policymocks.Service
	pEvaluator *policymocks.Evaluator
	cache      *climocks.Cache
	repo       *climocks.Repository
)

func newService() clients.Service {
	pService = new(policymocks.Service)
	cache = new(climocks.Cache)
	idProvider := uuid.NewMock()
	sidProvider := uuid.NewMock()
	repo = new(climocks.Repository)
	chgRPCClient := new(chmocks.ChannelsServiceClient)
	gpgRPCClient := new(gpmocks.GroupsServiceClient)
	tsv, _ := clients.NewService(repo, pService, cache, chgRPCClient, gpgRPCClient, idProvider, sidProvider)
	return tsv
}

func TestCreateClients(t *testing.T) {
	svc := newService()

	cases := []struct {
		desc            string
		thing           clients.Client
		token           string
		addPolicyErr    error
		deletePolicyErr error
		saveErr         error
		err             error
	}{
		{
			desc:  "create a new thing successfully",
			thing: thing,
			token: validToken,
			err:   nil,
		},
		{
			desc:    "create an existing thing",
			thing:   thing,
			token:   validToken,
			saveErr: repoerr.ErrConflict,
			err:     repoerr.ErrConflict,
		},
		{
			desc: "create a new thing without secret",
			thing: clients.Client{
				Name: "thingWithoutSecret",
				Credentials: clients.Credentials{
					Identity: "newthingwithoutsecret@example.com",
				},
				Status: clients.EnabledStatus,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new thing without identity",
			thing: clients.Client{
				Name: "thingWithoutIdentity",
				Credentials: clients.Credentials{
					Identity: "newthingwithoutsecret@example.com",
				},
				Status: clients.EnabledStatus,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new enabled thing with name",
			thing: clients.Client{
				Name: "thingWithName",
				Credentials: clients.Credentials{
					Identity: "newthingwithname@example.com",
					Secret:   secret,
				},
				Status: clients.EnabledStatus,
			},
			token: validToken,
			err:   nil,
		},

		{
			desc: "create a new disabled thing with name",
			thing: clients.Client{
				Name: "thingWithName",
				Credentials: clients.Credentials{
					Identity: "newthingwithname@example.com",
					Secret:   secret,
				},
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new enabled thing with tags",
			thing: clients.Client{
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newthingwithtags@example.com",
					Secret:   secret,
				},
				Status: clients.EnabledStatus,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new disabled thing with tags",
			thing: clients.Client{
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newthingwithtags@example.com",
					Secret:   secret,
				},
				Status: clients.DisabledStatus,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new enabled thing with metadata",
			thing: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newthingwithmetadata@example.com",
					Secret:   secret,
				},
				Metadata: validTMetadata,
				Status:   clients.EnabledStatus,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new disabled thing with metadata",
			thing: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newthingwithmetadata@example.com",
					Secret:   secret,
				},
				Metadata: validTMetadata,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new disabled thing",
			thing: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newthingwithvalidstatus@example.com",
					Secret:   secret,
				},
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new thing with valid disabled status",
			thing: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newthingwithvalidstatus@example.com",
					Secret:   secret,
				},
				Status: clients.DisabledStatus,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new thing with all fields",
			thing: clients.Client{
				Name: "newthingwithallfields",
				Tags: []string{"tag1", "tag2"},
				Credentials: clients.Credentials{
					Identity: "newthingwithallfields@example.com",
					Secret:   secret,
				},
				Metadata: clients.Metadata{
					"name": "newthingwithallfields",
				},
				Status: clients.EnabledStatus,
			},
			token: validToken,
			err:   nil,
		},
		{
			desc: "create a new thing with invalid status",
			thing: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newthingwithinvalidstatus@example.com",
					Secret:   secret,
				},
				Status: clients.AllStatus,
			},
			token: validToken,
			err:   svcerr.ErrInvalidStatus,
		},
		{
			desc: "create a new thing with failed add policies response",
			thing: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newthingwithfailedpolicy@example.com",
					Secret:   secret,
				},
				Status: clients.EnabledStatus,
			},
			token:        validToken,
			addPolicyErr: svcerr.ErrInvalidPolicy,
			err:          svcerr.ErrInvalidPolicy,
		},
		{
			desc: "create a new thing with failed delete policies response",
			thing: clients.Client{
				Credentials: clients.Credentials{
					Identity: "newthingwithfailedpolicy@example.com",
					Secret:   secret,
				},
				Status: clients.EnabledStatus,
			},
			token:           validToken,
			saveErr:         repoerr.ErrConflict,
			deletePolicyErr: svcerr.ErrInvalidPolicy,
			err:             repoerr.ErrConflict,
		},
	}

	for _, tc := range cases {
		repoCall := repo.On("Save", context.Background(), mock.Anything).Return([]clients.Client{tc.thing}, tc.saveErr)
		policyCall := pService.On("AddPolicies", mock.Anything, mock.Anything).Return(tc.addPolicyErr)
		policyCall1 := pService.On("DeletePolicies", mock.Anything, mock.Anything).Return(tc.deletePolicyErr)
		expected, err := svc.CreateClients(context.Background(), mgauthn.Session{}, tc.thing)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			tc.thing.ID = expected[0].ID
			tc.thing.CreatedAt = expected[0].CreatedAt
			tc.thing.UpdatedAt = expected[0].UpdatedAt
			tc.thing.Credentials.Secret = expected[0].Credentials.Secret
			tc.thing.Domain = expected[0].Domain
			tc.thing.UpdatedBy = expected[0].UpdatedBy
			assert.Equal(t, tc.thing, expected[0], fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.thing, expected[0]))
		}
		repoCall.Unset()
		policyCall.Unset()
		policyCall1.Unset()
	}
}

func TestViewClient(t *testing.T) {
	svc := newService()

	cases := []struct {
		desc        string
		clientID    string
		response    clients.Client
		retrieveErr error
		err         error
	}{
		{
			desc:     "view thing successfully",
			response: thing,
			clientID: thing.ID,
			err:      nil,
		},
		{
			desc:     "view thing with an invalid token",
			response: clients.Client{},
			clientID: "",
			err:      svcerr.ErrAuthorization,
		},
		{
			desc:        "view thing with valid token and invalid thing id",
			response:    clients.Client{},
			clientID:    wrongID,
			retrieveErr: svcerr.ErrNotFound,
			err:         svcerr.ErrNotFound,
		},
		{
			desc:     "view thing with an invalid token and invalid thing id",
			response: clients.Client{},
			clientID: wrongID,
			err:      svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		repoCall1 := repo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.response, tc.err)
		rThing, err := svc.View(context.Background(), mgauthn.Session{}, tc.clientID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, rThing, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, rThing))
		repoCall1.Unset()
	}
}

func TestListClients(t *testing.T) {
	svc := newService()

	adminID := testsutil.GenerateUUID(t)
	domainID := testsutil.GenerateUUID(t)
	nonAdminID := testsutil.GenerateUUID(t)
	thing.Permissions = []string{"read", "write"}

	cases := []struct {
		desc                    string
		userKind                string
		session                 mgauthn.Session
		page                    clients.Page
		listObjectsResponse     policysvc.PolicyPage
		retrieveAllResponse     clients.ClientsPage
		listPermissionsResponse policysvc.Permissions
		response                clients.ClientsPage
		id                      string
		size                    uint64
		listObjectsErr          error
		retrieveAllErr          error
		listPermissionsErr      error
		err                     error
	}{
		{
			desc:     "list all things successfully as non admin",
			userKind: "non-admin",
			session:  mgauthn.Session{UserID: nonAdminID, DomainID: domainID, SuperAdmin: false},
			id:       nonAdminID,
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
			},
			listObjectsResponse: policysvc.PolicyPage{Policies: []string{thing.ID, thing.ID}},
			retrieveAllResponse: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{thing, thing},
			},
			listPermissionsResponse: []string{"read", "write"},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{thing, thing},
			},
			err: nil,
		},
		{
			desc:     "list all things as non admin with failed to retrieve all",
			userKind: "non-admin",
			session:  mgauthn.Session{UserID: nonAdminID, DomainID: domainID, SuperAdmin: false},
			id:       nonAdminID,
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
			},
			listObjectsResponse: policysvc.PolicyPage{Policies: []string{thing.ID, thing.ID}},
			retrieveAllResponse: clients.ClientsPage{},
			response:            clients.ClientsPage{},
			retrieveAllErr:      repoerr.ErrNotFound,
			err:                 svcerr.ErrNotFound,
		},
		{
			desc:     "list all things as non admin with failed to list permissions",
			userKind: "non-admin",
			session:  mgauthn.Session{UserID: nonAdminID, DomainID: domainID, SuperAdmin: false},
			id:       nonAdminID,
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
			},
			listObjectsResponse: policysvc.PolicyPage{Policies: []string{thing.ID, thing.ID}},
			retrieveAllResponse: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{thing, thing},
			},
			listPermissionsResponse: []string{},
			response:                clients.ClientsPage{},
			listPermissionsErr:      svcerr.ErrNotFound,
			err:                     svcerr.ErrNotFound,
		},
		{
			desc:     "list all things as non admin with failed super admin",
			userKind: "non-admin",
			session:  mgauthn.Session{UserID: nonAdminID, DomainID: domainID, SuperAdmin: false},
			id:       nonAdminID,
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
			},
			response:            clients.ClientsPage{},
			listObjectsResponse: policysvc.PolicyPage{},
			err:                 nil,
		},
		{
			desc:     "list all things as non admin with failed to list objects",
			userKind: "non-admin",
			id:       nonAdminID,
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
			},
			response:            clients.ClientsPage{},
			listObjectsResponse: policysvc.PolicyPage{},
			listObjectsErr:      svcerr.ErrNotFound,
			err:                 svcerr.ErrNotFound,
		},
	}

	for _, tc := range cases {
		listAllObjectsCall := pService.On("ListAllObjects", mock.Anything, mock.Anything).Return(tc.listObjectsResponse, tc.listObjectsErr)
		retrieveAllCall := repo.On("SearchClients", mock.Anything, mock.Anything).Return(tc.retrieveAllResponse, tc.retrieveAllErr)
		listPermissionsCall := pService.On("ListPermissions", mock.Anything, mock.Anything, mock.Anything).Return(tc.listPermissionsResponse, tc.listPermissionsErr)
		page, err := svc.ListClients(context.Background(), tc.session, tc.id, tc.page)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, page, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, page))
		listAllObjectsCall.Unset()
		retrieveAllCall.Unset()
		listPermissionsCall.Unset()
	}

	cases2 := []struct {
		desc                    string
		userKind                string
		session                 mgauthn.Session
		page                    clients.Page
		listObjectsResponse     policysvc.PolicyPage
		retrieveAllResponse     clients.ClientsPage
		listPermissionsResponse policysvc.Permissions
		response                clients.ClientsPage
		id                      string
		size                    uint64
		listObjectsErr          error
		retrieveAllErr          error
		listPermissionsErr      error
		err                     error
	}{
		{
			desc:     "list all things as admin successfully",
			userKind: "admin",
			id:       adminID,
			session:  mgauthn.Session{UserID: adminID, DomainID: domainID, SuperAdmin: true},
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
				Domain:    domainID,
			},
			listObjectsResponse: policysvc.PolicyPage{Policies: []string{thing.ID, thing.ID}},
			retrieveAllResponse: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{thing, thing},
			},
			listPermissionsResponse: []string{"read", "write"},
			response: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{thing, thing},
			},
			err: nil,
		},
		{
			desc:     "list all things as admin with failed to retrieve all",
			userKind: "admin",
			id:       adminID,
			session:  mgauthn.Session{UserID: adminID, DomainID: domainID, SuperAdmin: true},
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
				Domain:    domainID,
			},
			listObjectsResponse: policysvc.PolicyPage{},
			retrieveAllResponse: clients.ClientsPage{},
			retrieveAllErr:      repoerr.ErrNotFound,
			err:                 svcerr.ErrNotFound,
		},
		{
			desc:     "list all things as admin with failed to list permissions",
			userKind: "admin",
			id:       adminID,
			session:  mgauthn.Session{UserID: adminID, DomainID: domainID, SuperAdmin: true},
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
				Domain:    domainID,
			},
			listObjectsResponse: policysvc.PolicyPage{},
			retrieveAllResponse: clients.ClientsPage{
				Page: clients.Page{
					Total:  2,
					Offset: 0,
					Limit:  100,
				},
				Clients: []clients.Client{thing, thing},
			},
			listPermissionsResponse: []string{},
			listPermissionsErr:      svcerr.ErrNotFound,
			err:                     svcerr.ErrNotFound,
		},
		{
			desc:     "list all things as admin with failed to list things",
			userKind: "admin",
			id:       adminID,
			session:  mgauthn.Session{UserID: adminID, DomainID: domainID, SuperAdmin: true},
			page: clients.Page{
				Offset:    0,
				Limit:     100,
				ListPerms: true,
				Domain:    domainID,
			},
			retrieveAllResponse: clients.ClientsPage{},
			retrieveAllErr:      repoerr.ErrNotFound,
			err:                 svcerr.ErrNotFound,
		},
	}

	for _, tc := range cases2 {
		listAllObjectsCall := pService.On("ListAllObjects", context.Background(), policysvc.Policy{
			SubjectType: policysvc.UserType,
			Subject:     tc.session.DomainID + "_" + adminID,
			Permission:  "",
			ObjectType:  policysvc.ThingType,
		}).Return(tc.listObjectsResponse, tc.listObjectsErr)
		listAllObjectsCall2 := pService.On("ListAllObjects", context.Background(), policysvc.Policy{
			SubjectType: policysvc.UserType,
			Subject:     tc.session.UserID,
			Permission:  "",
			ObjectType:  policysvc.ThingType,
		}).Return(tc.listObjectsResponse, tc.listObjectsErr)
		retrieveAllCall := repo.On("SearchClients", mock.Anything, mock.Anything).Return(tc.retrieveAllResponse, tc.retrieveAllErr)
		listPermissionsCall := pService.On("ListPermissions", mock.Anything, mock.Anything, mock.Anything).Return(tc.listPermissionsResponse, tc.listPermissionsErr)
		page, err := svc.ListClients(context.Background(), tc.session, tc.id, tc.page)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, page, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.response, page))
		listAllObjectsCall.Unset()
		listAllObjectsCall2.Unset()
		retrieveAllCall.Unset()
		listPermissionsCall.Unset()
	}
}

func TestUpdateClient(t *testing.T) {
	svc := newService()

	thing1 := thing
	thing2 := thing
	thing1.Name = "Updated thing"
	thing2.Metadata = clients.Metadata{"role": "test"}

	cases := []struct {
		desc           string
		thing          clients.Client
		session        mgauthn.Session
		updateResponse clients.Client
		updateErr      error
		err            error
	}{
		{
			desc:           "update thing name successfully",
			thing:          thing1,
			session:        mgauthn.Session{UserID: validID},
			updateResponse: thing1,
			err:            nil,
		},
		{
			desc:           "update thing metadata with valid token",
			thing:          thing2,
			updateResponse: thing2,
			session:        mgauthn.Session{UserID: validID},
			err:            nil,
		},
		{
			desc:           "update thing with failed to update repo",
			thing:          thing1,
			updateResponse: clients.Client{},
			session:        mgauthn.Session{UserID: validID},
			updateErr:      repoerr.ErrMalformedEntity,
			err:            svcerr.ErrUpdateEntity,
		},
	}

	for _, tc := range cases {
		repoCall1 := repo.On("Update", context.Background(), mock.Anything).Return(tc.updateResponse, tc.updateErr)
		updatedThing, err := svc.Update(context.Background(), tc.session, tc.thing)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.updateResponse, updatedThing, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.updateResponse, updatedThing))
		repoCall1.Unset()
	}
}

func TestUpdateTags(t *testing.T) {
	svc := newService()

	thing.Tags = []string{"updated"}

	cases := []struct {
		desc           string
		thing          clients.Client
		session        mgauthn.Session
		updateResponse clients.Client
		updateErr      error
		err            error
	}{
		{
			desc:           "update thing tags successfully",
			thing:          thing,
			session:        mgauthn.Session{UserID: validID},
			updateResponse: thing,
			err:            nil,
		},
		{
			desc:           "update thing tags with failed to update repo",
			thing:          thing,
			updateResponse: clients.Client{},
			session:        mgauthn.Session{UserID: validID},
			updateErr:      repoerr.ErrMalformedEntity,
			err:            svcerr.ErrUpdateEntity,
		},
	}

	for _, tc := range cases {
		repoCall1 := repo.On("UpdateTags", context.Background(), mock.Anything).Return(tc.updateResponse, tc.updateErr)
		updatedThing, err := svc.UpdateTags(context.Background(), tc.session, tc.thing)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.updateResponse, updatedThing, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.updateResponse, updatedThing))
		repoCall1.Unset()
	}
}

func TestUpdateSecret(t *testing.T) {
	svc := newService()

	cases := []struct {
		desc                 string
		thing                clients.Client
		newSecret            string
		updateSecretResponse clients.Client
		session              mgauthn.Session
		updateErr            error
		err                  error
	}{
		{
			desc:      "update thing secret successfully",
			thing:     thing,
			newSecret: "newSecret",
			session:   mgauthn.Session{UserID: validID},
			updateSecretResponse: clients.Client{
				ID: thing.ID,
				Credentials: clients.Credentials{
					Identity: thing.Credentials.Identity,
					Secret:   "newSecret",
				},
			},
			err: nil,
		},
		{
			desc:                 "update thing secret with failed to update repo",
			thing:                thing,
			newSecret:            "newSecret",
			session:              mgauthn.Session{UserID: validID},
			updateSecretResponse: clients.Client{},
			updateErr:            repoerr.ErrMalformedEntity,
			err:                  svcerr.ErrUpdateEntity,
		},
	}

	for _, tc := range cases {
		repoCall := repo.On("UpdateSecret", context.Background(), mock.Anything).Return(tc.updateSecretResponse, tc.updateErr)
		updatedThing, err := svc.UpdateSecret(context.Background(), tc.session, tc.thing.ID, tc.newSecret)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.updateSecretResponse, updatedThing, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.updateSecretResponse, updatedThing))
		repoCall.Unset()
	}
}

func TestEnable(t *testing.T) {
	svc := newService()

	enabledThing1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "thing1@example.com", Secret: "password"}, Status: clients.EnabledStatus}
	disabledThing1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "thing3@example.com", Secret: "password"}, Status: clients.DisabledStatus}
	endisabledThing1 := disabledThing1
	endisabledThing1.Status = clients.EnabledStatus

	cases := []struct {
		desc                 string
		id                   string
		session              mgauthn.Session
		thing                clients.Client
		changeStatusResponse clients.Client
		retrieveByIDResponse clients.Client
		changeStatusErr      error
		retrieveIDErr        error
		err                  error
	}{
		{
			desc:                 "enable disabled thing",
			id:                   disabledThing1.ID,
			session:              mgauthn.Session{UserID: validID},
			thing:                disabledThing1,
			changeStatusResponse: endisabledThing1,
			retrieveByIDResponse: disabledThing1,
			err:                  nil,
		},
		{
			desc:                 "enable disabled thing with failed to update repo",
			id:                   disabledThing1.ID,
			session:              mgauthn.Session{UserID: validID},
			thing:                disabledThing1,
			changeStatusResponse: clients.Client{},
			retrieveByIDResponse: disabledThing1,
			changeStatusErr:      repoerr.ErrMalformedEntity,
			err:                  svcerr.ErrUpdateEntity,
		},
		{
			desc:                 "enable enabled thing",
			id:                   enabledThing1.ID,
			session:              mgauthn.Session{UserID: validID},
			thing:                enabledThing1,
			changeStatusResponse: enabledThing1,
			retrieveByIDResponse: enabledThing1,
			changeStatusErr:      errors.ErrStatusAlreadyAssigned,
			err:                  errors.ErrStatusAlreadyAssigned,
		},
		{
			desc:                 "enable non-existing thing",
			id:                   wrongID,
			session:              mgauthn.Session{UserID: validID},
			thing:                clients.Client{},
			changeStatusResponse: clients.Client{},
			retrieveByIDResponse: clients.Client{},
			retrieveIDErr:        repoerr.ErrNotFound,
			err:                  repoerr.ErrNotFound,
		},
	}

	for _, tc := range cases {
		repoCall := repo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.retrieveByIDResponse, tc.retrieveIDErr)
		repoCall1 := repo.On("ChangeStatus", context.Background(), mock.Anything).Return(tc.changeStatusResponse, tc.changeStatusErr)
		_, err := svc.Enable(context.Background(), tc.session, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestDisable(t *testing.T) {
	svc := newService()

	enabledThing1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "thing1@example.com", Secret: "password"}, Status: clients.EnabledStatus}
	disabledThing1 := clients.Client{ID: ID, Credentials: clients.Credentials{Identity: "thing3@example.com", Secret: "password"}, Status: clients.DisabledStatus}
	disenabledClient1 := enabledThing1
	disenabledClient1.Status = clients.DisabledStatus

	cases := []struct {
		desc                 string
		id                   string
		session              mgauthn.Session
		thing                clients.Client
		changeStatusResponse clients.Client
		retrieveByIDResponse clients.Client
		changeStatusErr      error
		retrieveIDErr        error
		removeErr            error
		err                  error
	}{
		{
			desc:                 "disable enabled thing",
			id:                   enabledThing1.ID,
			session:              mgauthn.Session{UserID: validID},
			thing:                enabledThing1,
			changeStatusResponse: disenabledClient1,
			retrieveByIDResponse: enabledThing1,
			err:                  nil,
		},
		{
			desc:                 "disable thing with failed to update repo",
			id:                   enabledThing1.ID,
			session:              mgauthn.Session{UserID: validID},
			thing:                enabledThing1,
			changeStatusResponse: clients.Client{},
			retrieveByIDResponse: enabledThing1,
			changeStatusErr:      repoerr.ErrMalformedEntity,
			err:                  svcerr.ErrUpdateEntity,
		},
		{
			desc:                 "disable disabled thing",
			id:                   disabledThing1.ID,
			session:              mgauthn.Session{UserID: validID},
			thing:                disabledThing1,
			changeStatusResponse: clients.Client{},
			retrieveByIDResponse: disabledThing1,
			changeStatusErr:      errors.ErrStatusAlreadyAssigned,
			err:                  errors.ErrStatusAlreadyAssigned,
		},
		{
			desc:                 "disable non-existing thing",
			id:                   wrongID,
			thing:                clients.Client{},
			session:              mgauthn.Session{UserID: validID},
			changeStatusResponse: clients.Client{},
			retrieveByIDResponse: clients.Client{},
			retrieveIDErr:        repoerr.ErrNotFound,
			err:                  repoerr.ErrNotFound,
		},
		{
			desc:                 "disable thing with failed to remove from cache",
			id:                   enabledThing1.ID,
			session:              mgauthn.Session{UserID: validID},
			thing:                disabledThing1,
			changeStatusResponse: disenabledClient1,
			retrieveByIDResponse: enabledThing1,
			removeErr:            svcerr.ErrRemoveEntity,
			err:                  svcerr.ErrRemoveEntity,
		},
	}

	for _, tc := range cases {
		repoCall := repo.On("RetrieveByID", context.Background(), mock.Anything).Return(tc.retrieveByIDResponse, tc.retrieveIDErr)
		repoCall1 := repo.On("ChangeStatus", context.Background(), mock.Anything).Return(tc.changeStatusResponse, tc.changeStatusErr)
		repoCall2 := cache.On("Remove", mock.Anything, mock.Anything).Return(tc.removeErr)
		_, err := svc.Disable(context.Background(), tc.session, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
	}
}

func TestDelete(t *testing.T) {
	svc := newService()

	client := clients.Client{
		ID: testsutil.GenerateUUID(t),
	}

	cases := []struct {
		desc            string
		clientID        string
		removeErr       error
		deleteErr       error
		deletePolicyErr error
		err             error
	}{
		{
			desc:     "Delete client successfully",
			clientID: client.ID,
			err:      nil,
		},
		{
			desc:      "Delete non-existing client",
			clientID:  wrongID,
			deleteErr: repoerr.ErrNotFound,
			err:       svcerr.ErrRemoveEntity,
		},
		{
			desc:      "Delete client with repo error ",
			clientID:  client.ID,
			deleteErr: repoerr.ErrRemoveEntity,
			err:       repoerr.ErrRemoveEntity,
		},
		{
			desc:      "Delete client with cache error ",
			clientID:  client.ID,
			removeErr: svcerr.ErrRemoveEntity,
			err:       repoerr.ErrRemoveEntity,
		},
		{
			desc:            "Delete client with failed to delete policies",
			clientID:        client.ID,
			deletePolicyErr: errRemovePolicies,
			err:             errRemovePolicies,
		},
	}

	for _, tc := range cases {
		repoCall := cache.On("Remove", mock.Anything, tc.clientID).Return(tc.removeErr)
		policyCall := pService.On("DeletePolicyFilter", context.Background(), mock.Anything).Return(tc.deletePolicyErr)
		repoCall1 := repo.On("Delete", context.Background(), tc.clientID).Return(tc.deleteErr)
		err := svc.Delete(context.Background(), mgauthn.Session{}, tc.clientID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
		policyCall.Unset()
		repoCall1.Unset()
	}
}

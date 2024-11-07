// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	api "github.com/absmach/magistrala/clients/api/http"
	"github.com/absmach/magistrala/clients/mocks"
	gmocks "github.com/absmach/magistrala/groups/mocks"
	"github.com/absmach/magistrala/internal/testsutil"
	mglog "github.com/absmach/magistrala/logger"
	"github.com/absmach/magistrala/pkg/apiutil"
	mgauthn "github.com/absmach/magistrala/pkg/authn"
	authnmocks "github.com/absmach/magistrala/pkg/authn/mocks"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	policies "github.com/absmach/magistrala/pkg/policies"
	sdk "github.com/absmach/magistrala/pkg/sdk/go"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupClients() (*httptest.Server, *mocks.Service, *authnmocks.Authentication) {
	tsvc := new(mocks.Service)
	gsvc := new(gmocks.Service)

	logger := mglog.NewMock()
	mux := chi.NewRouter()
	authn := new(authnmocks.Authentication)
	api.MakeHandler(tsvc, gsvc, authn, mux, logger, "")

	return httptest.NewServer(mux), tsvc, authn
}

func TestCreateClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)
	createClientReq := sdk.Client{
		Name:        client.Name,
		Tags:        client.Tags,
		Credentials: client.Credentials,
		Metadata:    client.Metadata,
		Status:      client.Status,
	}

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}

	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		createClientReq sdk.Client
		svcReq          mgthings.Client
		svcRes          []mgthings.Client
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:            "create new client successfully",
			domainID:        domainID,
			token:           validToken,
			createClientReq: createClientReq,
			svcReq:          convertClient(createClientReq),
			svcRes:          []mgthings.Client{convertClient(client)},
			svcErr:          nil,
			response:        client,
			err:             nil,
		},
		{
			desc:            "create new client with invalid token",
			domainID:        domainID,
			token:           invalidToken,
			createClientReq: createClientReq,
			svcReq:          convertClient(createClientReq),
			svcRes:          []mgthings.Client{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:            "create new client with empty token",
			domainID:        domainID,
			token:           "",
			createClientReq: createClientReq,
			svcReq:          convertClient(createClientReq),
			svcRes:          []mgthings.Client{},
			svcErr:          nil,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:            "create an existing thing",
			domainID:        domainID,
			token:           validToken,
			createClientReq: createClientReq,
			svcReq:          convertClient(createClientReq),
			svcRes:          []mgthings.Client{},
			svcErr:          svcerr.ErrCreateEntity,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrCreateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:     "create a client with name too long",
			domainID: domainID,
			token:    validToken,
			createClientReq: sdk.Client{
				Name:        strings.Repeat("a", 1025),
				Tags:        client.Tags,
				Credentials: client.Credentials,
				Metadata:    client.Metadata,
				Status:      client.Status,
			},
			svcReq:   mgthings.Client{},
			svcRes:   []mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrNameSize), http.StatusBadRequest),
		},
		{
			desc:     "create a client with invalid id",
			domainID: domainID,
			token:    validToken,
			createClientReq: sdk.Client{
				ID:          "123456789",
				Name:        client.Name,
				Tags:        client.Tags,
				Credentials: client.Credentials,
				Metadata:    client.Metadata,
				Status:      client.Status,
			},
			svcReq:   mgthings.Client{},
			svcRes:   []mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrInvalidIDFormat), http.StatusBadRequest),
		},
		{
			desc:     "create a client with a request that can't be marshalled",
			domainID: domainID,
			token:    validToken,
			createClientReq: sdk.Client{
				Name: "test",
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
			},
			svcReq:   mgthings.Client{},
			svcRes:   []mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:            "create a client with a response that can't be unmarshalled",
			domainID:        domainID,
			token:           validToken,
			createClientReq: createClientReq,
			svcReq:          convertClient(createClientReq),
			svcRes: []mgthings.Client{{
				Name:        client.Name,
				Tags:        client.Tags,
				Credentials: mgthings.Credentials(client.Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			}},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("CreateClients", mock.Anything, tc.session, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.CreateClient(tc.createClientReq, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "CreateClients", mock.Anything, tc.session, tc.svcReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestCreateClients(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	clients := []sdk.Client{}
	for i := 0; i < 3; i++ {
		client := generateTestClient(t)
		clients = append(clients, client)
	}

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc                 string
		domainID             string
		token                string
		session              mgauthn.Session
		createClientsRequest []sdk.Client
		svcReq               []mgthings.Client
		svcRes               []mgthings.Client
		svcErr               error
		authenticateErr      error
		response             []sdk.Client
		err                  errors.SDKError
	}{
		{
			desc:                 "create new clients successfully",
			domainID:             domainID,
			token:                validToken,
			createClientsRequest: clients,
			svcReq:               convertClients(clients...),
			svcRes:               convertClients(clients...),
			svcErr:               nil,
			response:             clients,
			err:                  nil,
		},
		{
			desc:                 "create new clients with invalid token",
			domainID:             domainID,
			token:                invalidToken,
			createClientsRequest: clients,
			svcReq:               convertClients(clients...),
			svcRes:               []mgthings.Client{},
			authenticateErr:      svcerr.ErrAuthentication,
			response:             []sdk.Client{},
			err:                  errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:                 "create new clients with empty token",
			domainID:             domainID,
			token:                "",
			createClientsRequest: clients,
			svcReq:               convertClients(clients...),
			svcRes:               []mgthings.Client{},
			svcErr:               nil,
			response:             []sdk.Client{},
			err:                  errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:                 "create new clients with a request that can't be marshalled",
			domainID:             domainID,
			token:                validToken,
			createClientsRequest: []sdk.Client{{Name: "test", Metadata: map[string]interface{}{"test": make(chan int)}}},
			svcReq:               convertClients(clients...),
			svcRes:               []mgthings.Client{},
			svcErr:               nil,
			response:             []sdk.Client{},
			err:                  errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:                 "create new clients with a response that can't be unmarshalled",
			domainID:             domainID,
			token:                validToken,
			createClientsRequest: clients,
			svcReq:               convertClients(clients...),
			svcRes: []mgthings.Client{{
				Name:        clients[0].Name,
				Tags:        clients[0].Tags,
				Credentials: mgthings.Credentials(clients[0].Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			}},
			svcErr:   nil,
			response: []sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("CreateClients", mock.Anything, tc.session, tc.svcReq[0], tc.svcReq[1], tc.svcReq[2]).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.CreateClients(tc.createClientsRequest, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "CreateClients", mock.Anything, tc.session, tc.svcReq[0], tc.svcReq[1], tc.svcReq[2])
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestListClients(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	var clients []sdk.Client
	for i := 10; i < 100; i++ {
		c := generateTestClient(t)
		if i == 50 {
			c.Status = mgthings.DisabledStatus.String()
			c.Tags = []string{"tag1", "tag2"}
		}
		clients = append(clients, c)
	}

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		token           string
		domainID        string
		session         mgauthn.Session
		pageMeta        sdk.PageMetadata
		svcReq          mgthings.Page
		svcRes          mgthings.ClientsPage
		svcErr          error
		authenticateErr error
		response        sdk.ClientsPage
		err             errors.SDKError
	}{
		{
			desc:     "list all clients successfully",
			domainID: domainID,
			token:    validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  uint64(len(clients)),
				},
				Clients: convertClients(clients...),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: uint64(len(clients)),
				},
				Clients: clients,
			},
		},
		{
			desc:     "list all clients with an invalid token",
			domainID: domainID,
			token:    invalidToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes:          mgthings.ClientsPage{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.ClientsPage{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "list all clients with limit greater than max",
			domainID: domainID,
			token:    validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  1000,
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.ClientsPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrLimitSize), http.StatusBadRequest),
		},
		{
			desc:     "list all clients with name size greater than max",
			domainID: domainID,
			token:    validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
				Name:   strings.Repeat("a", 1025),
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.ClientsPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrNameSize), http.StatusBadRequest),
		},
		{
			desc:     "list all clients with status",
			domainID: domainID,
			token:    validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
				Status: mgthings.DisabledStatus.String(),
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
				Status:     mgthings.DisabledStatus,
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Clients: convertClients(clients[50]),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: 1,
				},
				Clients: []sdk.Client{clients[50]},
			},
			err: nil,
		},
		{
			desc:     "list all clients with tags",
			domainID: domainID,
			token:    validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
				Tag:    "tag1",
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
				Tag:        "tag1",
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Clients: convertClients(clients[50]),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: 1,
				},
				Clients: []sdk.Client{clients[50]},
			},
			err: nil,
		},
		{
			desc:     "list all clients with invalid metadata",
			domainID: domainID,
			token:    validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.ClientsPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:     "list all clients with response that can't be unmarshalled",
			domainID: domainID,
			token:    validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Clients: []mgthings.Client{{
					Name:        clients[0].Name,
					Tags:        clients[0].Tags,
					Credentials: mgthings.Credentials(clients[0].Credentials),
					Metadata: mgthings.Metadata{
						"test": make(chan int),
					},
				}},
			},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("ListClients", mock.Anything, tc.session, mock.Anything, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.Clients(tc.pageMeta, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "ListClients", mock.Anything, tc.session, mock.Anything, tc.svcReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestListClientsByChannel(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	var clients []sdk.Client
	for i := 10; i < 100; i++ {
		c := generateTestClient(t)
		if i == 50 {
			c.Status = mgthings.DisabledStatus.String()
		}
		clients = append(clients, c)
	}

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		token           string
		domainID        string
		session         mgauthn.Session
		channelID       string
		pageMeta        sdk.PageMetadata
		svcReq          mgthings.Page
		svcRes          mgthings.MembersPage
		svcErr          error
		authenticateErr error
		response        sdk.ClientsPage
		err             errors.SDKError
	}{
		{
			desc:      "list clients successfully",
			domainID:  domainID,
			token:     validToken,
			channelID: validID,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes: mgthings.MembersPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  uint64(len(clients)),
				},
				Members: convertClients(clients...),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: uint64(len(clients)),
				},
				Clients: clients,
			},
		},
		{
			desc:      "list clients with an invalid token",
			domainID:  domainID,
			token:     invalidToken,
			channelID: validID,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes:          mgthings.MembersPage{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.ClientsPage{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:      "list clients with empty token",
			domainID:  domainID,
			token:     "",
			channelID: validID,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.MembersPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "list clients with status",
			domainID:  domainID,
			token:     validToken,
			channelID: validID,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
				Status: mgthings.DisabledStatus.String(),
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
				Status:     mgthings.DisabledStatus,
			},
			svcRes: mgthings.MembersPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Members: convertClients(clients[50]),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: 1,
				},
				Clients: []sdk.Client{clients[50]},
			},
			err: nil,
		},
		{
			desc:      "list clients with empty channel id",
			domainID:  domainID,
			token:     validToken,
			channelID: "",
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.MembersPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:      "list clients with invalid metadata",
			domainID:  domainID,
			token:     validToken,
			channelID: validID,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.MembersPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:      "list clients with response that can't be unmarshalled",
			domainID:  domainID,
			token:     validToken,
			channelID: validID,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes: mgthings.MembersPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Members: []mgthings.Client{{
					Name:        clients[0].Name,
					Tags:        clients[0].Tags,
					Credentials: mgthings.Credentials(clients[0].Credentials),
					Metadata: mgthings.Metadata{
						"test": make(chan int),
					},
				}},
			},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("ListClientsByGroup", mock.Anything, tc.session, tc.channelID, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.ClientsByChannel(tc.channelID, tc.pageMeta, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "ListClientsByGroup", mock.Anything, tc.session, tc.channelID, tc.svcReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestViewClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)
	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		clientID        string
		svcRes          mgthings.Client
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:     "view client successfully",
			domainID: domainID,
			token:    validToken,
			clientID: client.ID,
			svcRes:   convertClient(client),
			svcErr:   nil,
			response: client,
			err:      nil,
		},
		{
			desc:            "view client with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			clientID:        client.ID,
			svcRes:          mgthings.Client{},
			authenticateErr: svcerr.ErrAuthorization,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "view client with empty token",
			domainID: domainID,
			token:    "",
			clientID: client.ID,
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "view client with an invalid client id",
			domainID: domainID,
			token:    validToken,
			clientID: wrongID,
			svcRes:   mgthings.Client{},
			svcErr:   svcerr.ErrViewEntity,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrViewEntity, http.StatusBadRequest),
		},
		{
			desc:     "view client with empty client id",
			domainID: domainID,
			token:    validToken,
			clientID: "",
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(apiutil.ErrMissingID),
		},
		{
			desc:     "view client with response that can't be unmarshalled",
			domainID: domainID,
			token:    validToken,
			clientID: client.ID,
			svcRes: mgthings.Client{
				Name:        client.Name,
				Tags:        client.Tags,
				Credentials: mgthings.Credentials(client.Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("View", mock.Anything, tc.session, tc.clientID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.Client(tc.clientID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "View", mock.Anything, tc.session, tc.clientID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestViewClientPermissions(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := sdk.Client{
		Permissions: []string{policies.ViewPermission},
	}
	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		clientID        string
		svcRes          []string
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:     "view client permissions successfully",
			domainID: domainID,
			token:    validToken,
			clientID: validID,
			svcRes:   []string{policies.ViewPermission},
			svcErr:   nil,
			response: client,
			err:      nil,
		},
		{
			desc:            "view client permissions with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			clientID:        validID,
			svcRes:          []string{},
			authenticateErr: svcerr.ErrAuthorization,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "view client permissions with empty token",
			domainID: domainID,
			token:    "",
			clientID: client.ID,
			svcRes:   []string{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "view client permissions with an invalid client id",
			domainID: domainID,
			token:    validToken,
			clientID: wrongID,
			svcRes:   []string{},
			svcErr:   svcerr.ErrViewEntity,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrViewEntity, http.StatusBadRequest),
		},
		{
			desc:     "view client permissions with empty client id",
			domainID: domainID,
			token:    validToken,
			clientID: "",
			svcRes:   []string{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("ViewPerms", mock.Anything, tc.session, tc.clientID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.ClientPermissions(tc.clientID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "ViewPerms", mock.Anything, tc.session, tc.clientID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestUpdateClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)
	updatedClient := thing
	updatedClient.Name = "newName"
	updatedClient.Metadata = map[string]interface{}{
		"newKey": "newValue",
	}
	updateClientReq := sdk.Client{
		ID:       thing.ID,
		Name:     updatedClient.Name,
		Metadata: updatedClient.Metadata,
	}

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		updateClientReq sdk.Client
		svcReq          mgthings.Client
		svcRes          mgthings.Client
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:            "update client successfully",
			domainID:        domainID,
			token:           validToken,
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes:          convertClient(updatedClient),
			svcErr:          nil,
			response:        updatedClient,
			err:             nil,
		},
		{
			desc:            "update client with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes:          mgthings.Client{},
			authenticateErr: svcerr.ErrAuthorization,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:            "update client with empty token",
			domainID:        domainID,
			token:           "",
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes:          mgthings.Client{},
			svcErr:          nil,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "update client with an invalid client id",
			domainID: domainID,
			token:    validToken,
			updateClientReq: sdk.Client{
				ID:   wrongID,
				Name: updatedClient.Name,
			},
			svcReq: convertClient(sdk.Client{
				ID:   wrongID,
				Name: updatedClient.Name,
			}),
			svcRes:   mgthings.Client{},
			svcErr:   svcerr.ErrUpdateEntity,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:     "update client with empty client id",
			domainID: domainID,
			token:    validToken,

			updateClientReq: sdk.Client{
				ID:   "",
				Name: updatedClient.Name,
			},
			svcReq: convertClient(sdk.Client{
				ID:   "",
				Name: updatedClient.Name,
			}),
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(apiutil.ErrMissingID),
		},
		{
			desc:     "update client with a request that can't be marshalled",
			domainID: domainID,
			token:    validToken,

			updateClientReq: sdk.Client{
				ID: "test",
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
			},
			svcReq:   mgthings.Client{},
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:            "update client with a response that can't be unmarshalled",
			domainID:        domainID,
			token:           validToken,
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes: mgthings.Client{
				Name:        updatedClient.Name,
				Tags:        updatedClient.Tags,
				Credentials: mgthings.Credentials(updatedClient.Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("Update", mock.Anything, tc.session, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.UpdateClient(tc.updateClientReq, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Update", mock.Anything, tc.session, tc.svcReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestUpdateClientTags(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)
	updatedClient := thing
	updatedClient.Tags = []string{"newTag1", "newTag2"}
	updateClientReq := sdk.Client{
		ID:   thing.ID,
		Tags: updatedClient.Tags,
	}

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		updateClientReq sdk.Client
		svcReq          mgthings.Client
		svcRes          mgthings.Client
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:            "update client tags successfully",
			domainID:        domainID,
			token:           validToken,
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes:          convertClient(updatedClient),
			svcErr:          nil,
			response:        updatedClient,
			err:             nil,
		},
		{
			desc:            "update client tags with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes:          mgthings.Client{},
			authenticateErr: svcerr.ErrAuthorization,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:            "update client tags with empty token",
			domainID:        domainID,
			token:           "",
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes:          mgthings.Client{},
			svcErr:          nil,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "update client tags with an invalid client id",
			domainID: domainID,
			token:    validToken,
			updateClientReq: sdk.Client{
				ID:   wrongID,
				Tags: updatedClient.Tags,
			},
			svcReq: convertClient(sdk.Client{
				ID:   wrongID,
				Tags: updatedClient.Tags,
			}),
			svcRes:   mgthings.Client{},
			svcErr:   svcerr.ErrUpdateEntity,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:     "update client tags with empty client id",
			domainID: domainID,
			token:    validToken,
			updateClientReq: sdk.Client{
				ID:   "",
				Tags: updatedClient.Tags,
			},
			svcReq: convertClient(sdk.Client{
				ID:   "",
				Tags: updatedClient.Tags,
			}),
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "update client tags with a request that can't be marshalled",
			domainID: domainID,
			token:    validToken,
			updateClientReq: sdk.Client{
				ID: "test",
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
			},
			svcReq:   mgthings.Client{},
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:            "update client tags with a response that can't be unmarshalled",
			domainID:        domainID,
			token:           validToken,
			updateClientReq: updateClientReq,
			svcReq:          convertClient(updateClientReq),
			svcRes: mgthings.Client{
				Name:        updatedClient.Name,
				Tags:        updatedClient.Tags,
				Credentials: mgthings.Credentials(updatedClient.Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("UpdateTags", mock.Anything, tc.session, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.UpdateClientTags(tc.updateClientReq, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "UpdateTags", mock.Anything, tc.session, tc.svcReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestUpdateClientSecret(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)
	newSecret := generateUUID(t)
	updatedClient := thing
	updatedClient.Credentials.Secret = newSecret

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		clientID        string
		newSecret       string
		svcRes          mgthings.Client
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:      "update client secret successfully",
			domainID:  domainID,
			token:     validToken,
			clientID:  thing.ID,
			newSecret: newSecret,
			svcRes:    convertClient(updatedClient),
			svcErr:    nil,
			response:  updatedClient,
			err:       nil,
		},
		{
			desc:            "update client secret with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			clientID:        thing.ID,
			newSecret:       newSecret,
			svcRes:          mgthings.Client{},
			authenticateErr: svcerr.ErrAuthorization,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:      "update client secret with empty token",
			domainID:  domainID,
			token:     "",
			clientID:  thing.ID,
			newSecret: newSecret,
			svcRes:    mgthings.Client{},
			svcErr:    nil,
			response:  sdk.Client{},
			err:       errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:      "update client secret with an invalid client id",
			domainID:  domainID,
			token:     validToken,
			clientID:  wrongID,
			newSecret: newSecret,
			svcRes:    mgthings.Client{},
			svcErr:    svcerr.ErrUpdateEntity,
			response:  sdk.Client{},
			err:       errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:      "update client secret with empty client id",
			domainID:  domainID,
			token:     validToken,
			clientID:  "",
			newSecret: newSecret,
			svcRes:    mgthings.Client{},
			svcErr:    nil,
			response:  sdk.Client{},
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:      "update client with empty new secret",
			domainID:  domainID,
			token:     validToken,
			clientID:  thing.ID,
			newSecret: "",
			svcRes:    mgthings.Client{},
			svcErr:    nil,
			response:  sdk.Client{},
			err:       errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingSecret), http.StatusBadRequest),
		},
		{
			desc:      "update client secret with a response that can't be unmarshalled",
			domainID:  domainID,
			token:     validToken,
			clientID:  thing.ID,
			newSecret: newSecret,
			svcRes: mgthings.Client{
				Name:        updatedClient.Name,
				Tags:        updatedClient.Tags,
				Credentials: mgthings.Credentials(updatedClient.Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("UpdateSecret", mock.Anything, tc.session, tc.clientID, tc.newSecret).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.UpdateClientSecret(tc.clientID, tc.newSecret, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "UpdateSecret", mock.Anything, tc.session, tc.clientID, tc.newSecret)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestEnableClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)
	enabledClient := thing
	enabledClient.Status = mgthings.EnabledStatus.String()

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		thingID         string
		svcRes          mgthings.Client
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:     "enable client successfully",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			svcRes:   convertClient(enabledClient),
			svcErr:   nil,
			response: enabledClient,
			err:      nil,
		},
		{
			desc:            "enable client with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			thingID:         thing.ID,
			svcRes:          mgthings.Client{},
			authenticateErr: svcerr.ErrAuthorization,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "enable client with an invalid client id",
			domainID: domainID,
			token:    validToken,
			thingID:  wrongID,
			svcRes:   mgthings.Client{},
			svcErr:   svcerr.ErrEnableClient,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrEnableClient, http.StatusUnprocessableEntity),
		},
		{
			desc:     "enable client with empty client id",
			domainID: domainID,
			token:    validToken,
			thingID:  "",
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "enable client with a response that can't be unmarshalled",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			svcRes: mgthings.Client{
				Name:        enabledClient.Name,
				Tags:        enabledClient.Tags,
				Credentials: mgthings.Credentials(enabledClient.Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("Enable", mock.Anything, tc.session, tc.thingID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.EnableClient(tc.thingID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Enable", mock.Anything, tc.session, tc.thingID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestDisableClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)
	disabledClient := thing
	disabledClient.Status = mgthings.DisabledStatus.String()

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		thingID         string
		svcRes          mgthings.Client
		svcErr          error
		authenticateErr error
		response        sdk.Client
		err             errors.SDKError
	}{
		{
			desc:     "disable client successfully",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			svcRes:   convertClient(disabledClient),
			svcErr:   nil,
			response: disabledClient,
			err:      nil,
		},
		{
			desc:            "disable client with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			thingID:         thing.ID,
			svcRes:          mgthings.Client{},
			authenticateErr: svcerr.ErrAuthorization,
			response:        sdk.Client{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "disable client with an invalid client id",
			domainID: domainID,
			token:    validToken,
			thingID:  wrongID,
			svcRes:   mgthings.Client{},
			svcErr:   svcerr.ErrDisableClient,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrDisableClient, http.StatusInternalServerError),
		},
		{
			desc:     "disable client with empty client id",
			domainID: domainID,
			token:    validToken,
			thingID:  "",
			svcRes:   mgthings.Client{},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "disable client with a response that can't be unmarshalled",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			svcRes: mgthings.Client{
				Name:        disabledClient.Name,
				Tags:        disabledClient.Tags,
				Credentials: mgthings.Credentials(disabledClient.Credentials),
				Metadata: mgthings.Metadata{
					"test": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Client{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("Disable", mock.Anything, tc.session, tc.thingID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.DisableClient(tc.thingID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Disable", mock.Anything, tc.session, tc.thingID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestShareClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		thingID         string
		shareReq        sdk.UsersRelationRequest
		authenticateErr error
		svcErr          error
		err             errors.SDKError
	}{
		{
			desc:     "share client successfully",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "share client with an invalid token",
			domainID: domainID,
			token:    invalidToken,
			thingID:  thing.ID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			authenticateErr: svcerr.ErrAuthorization,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "share client with empty token",
			domainID: domainID,
			token:    "",
			thingID:  thing.ID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			svcErr: svcerr.ErrAuthentication,
			err:    errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "share client with an invalid client id",
			domainID: domainID,
			token:    validToken,
			thingID:  wrongID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			svcErr: svcerr.ErrUpdateEntity,
			err:    errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:     "share client with empty client id",
			domainID: domainID,
			token:    validToken,
			thingID:  "",
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			svcErr: nil,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
		{
			desc:     "share client with empty relation",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: "",
			},
			svcErr: nil,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMalformedPolicy), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("Share", mock.Anything, tc.session, tc.thingID, tc.shareReq.Relation, tc.shareReq.UserIDs[0]).Return(tc.svcErr)
			err := mgsdk.ShareClient(tc.thingID, tc.shareReq, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Share", mock.Anything, tc.session, tc.thingID, tc.shareReq.Relation, tc.shareReq.UserIDs[0])
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestUnshareClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		thingID         string
		shareReq        sdk.UsersRelationRequest
		authenticateErr error
		svcErr          error
		err             errors.SDKError
	}{
		{
			desc:     "unshare client successfully",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "unshare client with an invalid token",
			domainID: domainID,
			token:    invalidToken,
			thingID:  thing.ID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			authenticateErr: svcerr.ErrAuthorization,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "unshare client with empty token",
			domainID: domainID,
			token:    "",
			thingID:  thing.ID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			err: errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "unshare client with an invalid client id",
			domainID: domainID,
			token:    validToken,
			thingID:  wrongID,
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			svcErr: svcerr.ErrUpdateEntity,
			err:    errors.NewSDKErrorWithStatus(svcerr.ErrUpdateEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:     "unshare client with empty client id",
			domainID: domainID,
			token:    validToken,
			thingID:  "",
			shareReq: sdk.UsersRelationRequest{
				UserIDs:  []string{validID},
				Relation: policies.EditorRelation,
			},
			svcErr: nil,
			err:    errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrMissingID), http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("Unshare", mock.Anything, tc.session, tc.thingID, tc.shareReq.Relation, tc.shareReq.UserIDs[0]).Return(tc.svcErr)
			err := mgsdk.UnshareClient(tc.thingID, tc.shareReq, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Unshare", mock.Anything, tc.session, tc.thingID, tc.shareReq.Relation, tc.shareReq.UserIDs[0])
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestDeleteClient(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	client := generateTestClient(t)

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         mgauthn.Session
		thingID         string
		svcErr          error
		authenticateErr error
		err             errors.SDKError
	}{
		{
			desc:     "delete client successfully",
			domainID: domainID,
			token:    validToken,
			thingID:  thing.ID,
			svcErr:   nil,
			err:      nil,
		},
		{
			desc:            "delete client with an invalid token",
			domainID:        domainID,
			token:           invalidToken,
			thingID:         thing.ID,
			authenticateErr: svcerr.ErrAuthorization,
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "delete client with empty token",
			domainID: domainID,
			token:    "",
			thingID:  thing.ID,
			svcErr:   svcerr.ErrAuthentication,
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "delete client with an invalid client id",
			domainID: domainID,
			token:    validToken,
			thingID:  wrongID,
			svcErr:   svcerr.ErrRemoveEntity,
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrRemoveEntity, http.StatusUnprocessableEntity),
		},
		{
			desc:     "delete client with empty client id",
			domainID: domainID,
			token:    validToken,
			thingID:  "",
			svcErr:   nil,
			err:      errors.NewSDKError(apiutil.ErrMissingID),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("Delete", mock.Anything, tc.session, tc.thingID).Return(tc.svcErr)
			err := mgsdk.DeleteClient(tc.thingID, tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "Delete", mock.Anything, tc.session, tc.thingID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestListUserClients(t *testing.T) {
	ts, tsvc, auth := setupClients()
	defer ts.Close()

	var clients []sdk.Client
	for i := 10; i < 100; i++ {
		c := generateTestClient(t)
		if i == 50 {
			c.Status = mgthings.DisabledStatus.String()
			c.Tags = []string{"tag1", "tag2"}
		}
		clients = append(clients, c)
	}

	conf := sdk.Config{
		ClientsURL: ts.URL,
	}
	mgsdk := sdk.NewSDK(conf)

	cases := []struct {
		desc            string
		token           string
		session         mgauthn.Session
		userID          string
		pageMeta        sdk.PageMetadata
		svcReq          mgthings.Page
		svcRes          mgthings.ClientsPage
		svcErr          error
		authenticateErr error
		response        sdk.ClientsPage
		err             errors.SDKError
	}{
		{
			desc:   "list user clients successfully",
			token:  validToken,
			userID: validID,
			pageMeta: sdk.PageMetadata{
				Offset:   0,
				Limit:    100,
				DomainID: domainID,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  uint64(len(clients)),
				},
				Clients: convertClients(clients...),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: uint64(len(clients)),
				},
				Clients: clients,
			},
		},
		{
			desc:   "list user clients with an invalid token",
			token:  invalidToken,
			userID: validID,
			pageMeta: sdk.PageMetadata{
				Offset:   0,
				Limit:    100,
				DomainID: domainID,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes:          mgthings.ClientsPage{},
			authenticateErr: svcerr.ErrAuthentication,
			response:        sdk.ClientsPage{},
			err:             errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:   "list user clients with limit greater than max",
			token:  validToken,
			userID: validID,
			pageMeta: sdk.PageMetadata{
				Offset:   0,
				Limit:    1000,
				DomainID: domainID,
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.ClientsPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrLimitSize), http.StatusBadRequest),
		},
		{
			desc:   "list user clients with name size greater than max",
			token:  validToken,
			userID: validID,
			pageMeta: sdk.PageMetadata{
				Offset:   0,
				Limit:    100,
				Name:     strings.Repeat("a", 1025),
				DomainID: domainID,
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.ClientsPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKErrorWithStatus(errors.Wrap(apiutil.ErrValidation, apiutil.ErrNameSize), http.StatusBadRequest),
		},
		{
			desc:   "list user clients with status",
			token:  validToken,
			userID: validID,
			pageMeta: sdk.PageMetadata{
				Offset:   0,
				Limit:    100,
				Status:   mgthings.DisabledStatus.String(),
				DomainID: domainID,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
				Status:     mgthings.DisabledStatus,
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Clients: convertClients(clients[50]),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: 1,
				},
				Clients: []sdk.Client{clients[50]},
			},
			err: nil,
		},
		{
			desc:   "list user clients with tags",
			token:  validToken,
			userID: validID,
			pageMeta: sdk.PageMetadata{
				Offset:   0,
				Limit:    100,
				Tag:      "tag1",
				DomainID: domainID,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
				Tag:        "tag1",
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Clients: convertClients(clients[50]),
			},
			svcErr: nil,
			response: sdk.ClientsPage{
				PageRes: sdk.PageRes{
					Limit: 100,
					Total: 1,
				},
				Clients: []sdk.Client{clients[50]},
			},
			err: nil,
		},
		{
			desc:   "list user clients with invalid metadata",
			token:  validToken,
			userID: validID,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  100,
				Metadata: map[string]interface{}{
					"test": make(chan int),
				},
				DomainID: domainID,
			},
			svcReq:   mgthings.Page{},
			svcRes:   mgthings.ClientsPage{},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:  "list user clients with response that can't be unmarshalled",
			token: validToken,
			pageMeta: sdk.PageMetadata{
				Offset:   0,
				Limit:    100,
				DomainID: domainID,
			},
			svcReq: mgthings.Page{
				Offset:     0,
				Limit:      100,
				Permission: policies.ViewPermission,
			},
			svcRes: mgthings.ClientsPage{
				Page: mgthings.Page{
					Offset: 0,
					Limit:  100,
					Total:  1,
				},
				Clients: []mgthings.Client{{
					Name:        clients[0].Name,
					Tags:        clients[0].Tags,
					Credentials: mgthings.Credentials(clients[0].Credentials),
					Metadata: mgthings.Metadata{
						"test": make(chan int),
					},
				}},
			},
			svcErr:   nil,
			response: sdk.ClientsPage{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authenticateErr)
			svcCall := tsvc.On("ListClients", mock.Anything, tc.session, tc.userID, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.ListUserClients(tc.userID, tc.pageMeta, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "ListClients", mock.Anything, tc.session, tc.userID, tc.svcReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func generateTestClient(t *testing.T) sdk.Client {
	createdAt, err := time.Parse(time.RFC3339, "2023-03-03T00:00:00Z")
	assert.Nil(t, err, fmt.Sprintf("unexpected error %s", err))
	updatedAt := createdAt
	return sdk.Client{
		ID:   testsutil.GenerateUUID(t),
		Name: "clientname",
		Credentials: sdk.ClientCredentials{
			Identity: "thing@example.com",
			Secret:   generateUUID(t),
		},
		Tags:      []string{"tag1", "tag2"},
		Metadata:  validMetadata,
		Status:    mgthings.EnabledStatus.String(),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

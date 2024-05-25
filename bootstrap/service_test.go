// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package bootstrap_test

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"testing"

	"github.com/absmach/magistrala"
	authmocks "github.com/absmach/magistrala/auth/mocks"
	"github.com/absmach/magistrala/bootstrap"
	"github.com/absmach/magistrala/bootstrap/mocks"
	"github.com/absmach/magistrala/internal/testsutil"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	mgsdk "github.com/absmach/magistrala/pkg/sdk/go"
	sdkmocks "github.com/absmach/magistrala/pkg/sdk/mocks"
	"github.com/absmach/magistrala/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validToken      = "validToken"
	invalidToken    = "invalid"
	invalidDomainID = "invalid"
	email           = "test@example.com"
	unknown         = "unknown"
	channelsNum     = 3
	instanceID      = "5de9b29a-feb9-11ed-be56-0242ac120002"
	validID         = "d4ebb847-5d0e-4e46-bdd9-b6aceaaa3a22"
)

var (
	encKey   = []byte("1234567891011121")
	domainID = testsutil.GenerateUUID(&testing.T{})
	channel  = bootstrap.Channel{
		ID:       testsutil.GenerateUUID(&testing.T{}),
		Name:     "name",
		Metadata: map[string]interface{}{"name": "value"},
	}

	config = bootstrap.Config{
		ThingID:     testsutil.GenerateUUID(&testing.T{}),
		ThingKey:    testsutil.GenerateUUID(&testing.T{}),
		ExternalID:  testsutil.GenerateUUID(&testing.T{}),
		ExternalKey: testsutil.GenerateUUID(&testing.T{}),
		Channels:    []bootstrap.Channel{channel},
		Content:     "config",
	}
)

func newService() (bootstrap.Service, *mocks.ConfigRepository, *authmocks.AuthClient, *sdkmocks.SDK) {
	boot := new(mocks.ConfigRepository)
	auth := new(authmocks.AuthClient)
	sdk := new(sdkmocks.SDK)
	idp := uuid.NewMock()

	return bootstrap.New(auth, boot, sdk, encKey, idp), boot, auth, sdk
}

func enc(in []byte) ([]byte, error) {
	block, err := aes.NewCipher(encKey)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, aes.BlockSize+len(in))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], in)
	return ciphertext, nil
}

func TestAdd(t *testing.T) {
	svc, boot, auth, sdk := newService()
	neID := config
	neID.ThingID = "non-existent"

	wrongChannels := config
	ch := channel
	ch.ID = "invalid"
	wrongChannels.Channels = append(wrongChannels.Channels, ch)

	cases := []struct {
		desc            string
		config          bootstrap.Config
		token           string
		id              string
		domain          string
		authResponse    *magistrala.AuthorizeRes
		authorizeErr    error
		identifyErr     error
		thingErr        error
		createThingErr  error
		channelErr      error
		listExistingErr error
		saveErr         error
		deleteThingErr  error
		err             error
	}{
		{
			desc:         "add a new config",
			config:       config,
			token:        validToken,
			id:           validID,
			domain:       domainID,
			authResponse: &magistrala.AuthorizeRes{Authorized: true},
			err:          nil,
		},
		{
			desc:         "add a config with an invalid ID",
			config:       neID,
			token:        validToken,
			id:           validID,
			domain:       domainID,
			authResponse: &magistrala.AuthorizeRes{Authorized: true},
			thingErr:     errors.NewSDKError(svcerr.ErrNotFound),
			err:          svcerr.ErrNotFound,
		},
		{
			desc:   "add a config with invalid token",
			config: config,
			token:  invalidToken,
			domain: domainID,
			err:    svcerr.ErrAuthentication,
		},
		{
			desc:   "add a config with empty token",
			config: config,
			token:  "",
			domain: domainID,
			err:    svcerr.ErrAuthentication,
		},
		{
			desc:            "add a config with invalid list of channels",
			config:          wrongChannels,
			token:           validToken,
			id:              validID,
			domain:          domainID,
			authResponse:    &magistrala.AuthorizeRes{Authorized: true},
			listExistingErr: svcerr.ErrMalformedEntity,
			err:             svcerr.ErrMalformedEntity,
		},
		{
			desc:         "add ampty config",
			config:       bootstrap.Config{},
			token:        validToken,
			id:           validID,
			domain:       domainID,
			authResponse: &magistrala.AuthorizeRes{Authorized: true},
		},
		{
			desc:         "add a config without authorization",
			config:       config,
			token:        validToken,
			id:           validID,
			domain:       domainID,
			authResponse: &magistrala.AuthorizeRes{Authorized: false},
			authorizeErr: svcerr.ErrAuthorization,
			err:          svcerr.ErrAuthorization,
		},
		{
			desc:        "add a config with empty domain ID",
			config:      config,
			token:       validToken,
			id:          validID,
			domain:      "",
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc:        "add a config with invalid domain ID",
			config:      config,
			token:       validToken,
			id:          validID,
			domain:      invalidDomainID,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: tc.id, DomainId: tc.domain}, tc.identifyErr)
		authCall1 := auth.On("Authorize", context.Background(), mock.Anything).Return(tc.authResponse, tc.authorizeErr)
		sdkCall := sdk.On("Thing", tc.config.ThingID, tc.token).Return(mgsdk.Thing{ID: tc.config.ThingID, Credentials: mgsdk.Credentials{Secret: tc.config.ThingKey}}, tc.thingErr)
		sdkCall1 := sdk.On("Channel", ch.ID, tc.token).Return(mgsdk.Channel{}, tc.channelErr)
		sdkCall2 := sdk.On("CreateThing", mock.Anything, tc.token).Return(mgsdk.Thing{}, tc.createThingErr)
		sdkCall3 := sdk.On("DeleteThing", tc.config.ThingID, tc.token).Return(tc.deleteThingErr)
		svcCall := boot.On("ListExisting", context.Background(), tc.domain, mock.Anything).Return(tc.config.Channels, tc.listExistingErr)
		svcCall1 := boot.On("Save", context.Background(), mock.Anything, mock.Anything).Return(mock.Anything, tc.saveErr)

		_, err := svc.Add(context.Background(), tc.token, tc.config)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))

		authCall.Unset()
		authCall1.Unset()
		sdkCall.Unset()
		sdkCall1.Unset()
		sdkCall2.Unset()
		sdkCall3.Unset()
		svcCall.Unset()
		svcCall1.Unset()
	}
}

func TestView(t *testing.T) {
	svc, boot, auth, _ := newService()
	cases := []struct {
		desc         string
		id           string
		domain       string
		thingDomain  string
		authorizeRes *magistrala.AuthorizeRes
		token        string
		identifyErr  error
		authorizeErr error
		retrieveErr  error
		thingErr     error
		channelErr   error
		err          error
	}{
		{
			desc:         "view an existing config",
			id:           saved.ThingID,
			thingDomain:  domainID,
			domain:       domainID,
			token:        validToken,
			authorizeRes: &magistrala.AuthorizeRes{Authorized: true},
			err:          nil,
		},
		{
			desc:        "view a non-existing config",
			id:          unknown,
			thingDomain: domainID,
			domain:      domainID,
			token:       validToken,
			retrieveErr: errors.NewSDKError(svcerr.ErrNotFound),
			err:         svcerr.ErrNotFound,
		},
		{
			desc:        "view a config with wrong credentials",
			id:          config.ThingID,
			thingDomain: domainID,
			domain:      domainID,
			token:       invalidToken,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc:        "view a config with invalid domain",
			id:          config.ThingID,
			thingDomain: domainID,
			domain:      invalidDomainID,
			token:       validToken,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc:        "view a config with invalid thing domain",
			id:          config.ThingID,
			thingDomain: invalidDomainID,
			domain:      domainID,
			token:       validToken,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc:         "view a config with failed authorization",
			id:           saved.ThingID,
			thingDomain:  domainID,
			domain:       domainID,
			token:        validToken,
			authorizeRes: &magistrala.AuthorizeRes{Authorized: false},
			authorizeErr: svcerr.ErrAuthorization,
			err:          svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID, DomainId: tc.domain}, tc.identifyErr)
		authCall1 := auth.On("Authorize", context.Background(), mock.Anything).Return(tc.authorizeRes, tc.authorizeErr)
		sdkCall := sdk.On("Thing", tc.id, tc.token).Return(mgsdk.Thing{ID: config.ThingID, Credentials: mgsdk.Credentials{Secret: config.ThingKey}, DomainID: tc.domain}, tc.thingErr)
		sdkCall1 := sdk.On("Channel", mock.Anything, tc.token).Return(mgsdk.Channel{ID: channel.ID, DomainID: tc.domain}, tc.channelErr)
		svcCall := boot.On("RetrieveByID", context.Background(), mock.Anything, mock.Anything).Return(config, tc.retrieveErr)

		_, err := svc.View(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		authCall.Unset()
		authCall1.Unset()
		sdkCall.Unset()
		sdkCall1.Unset()
		svcCall.Unset()
	}
}

func TestUpdate(t *testing.T) {
	svc, boot, auth, _ := newService()
	c := config

	ch := channel
	ch.ID = "2"
	c.Channels = append(c.Channels, ch)

	modifiedCreated := c
	modifiedCreated.Content = "new-config"
	modifiedCreated.Name = "new name"

	nonExisting := config
	nonExisting.ThingID = unknown

	cases := []struct {
		desc         string
		config       bootstrap.Config
		token        string
		authorizeRes *magistrala.AuthorizeRes
		authorizeErr error
		identifyErr  error
		updateErr    error
		err          error
	}{
		{
			desc:         "update a config with state Created",
			config:       modifiedCreated,
			token:        validToken,
			authorizeRes: &magistrala.AuthorizeRes{Authorized: true},
			err:          nil,
		},
		{
			desc:         "update a non-existing config",
			config:       nonExisting,
			token:        validToken,
			authorizeRes: &magistrala.AuthorizeRes{Authorized: true},
			updateErr:    svcerr.ErrNotFound,
			err:          svcerr.ErrNotFound,
		},
		{
			desc:        "update a config with wrong credentials",
			config:      c,
			token:       invalidToken,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc:         "update a config with failed authorization",
			config:       saved,
			token:        validToken,
			authorizeRes: &magistrala.AuthorizeRes{Authorized: false},
			authorizeErr: svcerr.ErrAuthorization,
			err:          svcerr.ErrAuthorization,
		},
		{
			desc:         "update a config with update error",
			config:       saved,
			token:        validToken,
			authorizeRes: &magistrala.AuthorizeRes{Authorized: true},
			updateErr:    svcerr.ErrUpdateEntity,
			err:          svcerr.ErrUpdateEntity,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, tc.identifyErr)
		authCall1 := auth.On("Authorize", context.Background(), mock.Anything).Return(tc.authorizeRes, tc.authorizeErr)
		svcCall := boot.On("Update", context.Background(), mock.Anything).Return(tc.updateErr)
		err := svc.Update(context.Background(), tc.token, tc.config)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		authCall.Unset()
		authCall1.Unset()
		svcCall.Unset()
	}
}

func TestUpdateCert(t *testing.T) {
	svc, boot, auth, _ := newService()
	c := config

	cases := []struct {
		desc           string
		token          string
		thingID        string
		clientCert     string
		clientKey      string
		caCert         string
		expectedConfig bootstrap.Config
		identifyErr    error
		updateErr      error
		err            error
	}{
		{
			desc:       "update certs for the valid config",
			thingID:    c.ThingID,
			clientCert: "newCert",
			clientKey:  "newKey",
			caCert:     "newCert",
			token:      validToken,
			expectedConfig: bootstrap.Config{
				Name:        c.Name,
				ThingKey:    c.ThingKey,
				Channels:    c.Channels,
				ExternalID:  c.ExternalID,
				ExternalKey: c.ExternalKey,
				Content:     c.Content,
				State:       c.State,
				Owner:       c.Owner,
				ThingID:     c.ThingID,
				ClientCert:  "newCert",
				CACert:      "newCert",
				ClientKey:   "newKey",
			},
			err: nil,
		},
		{
			desc:           "update cert for a non-existing config",
			thingID:        "empty",
			clientCert:     "newCert",
			clientKey:      "newKey",
			caCert:         "newCert",
			token:          validToken,
			expectedConfig: bootstrap.Config{},
			updateErr:      svcerr.ErrNotFound,
			err:            svcerr.ErrNotFound,
		},
		{
			desc:           "update config cert with wrong credentials",
			thingID:        c.ThingID,
			clientCert:     "newCert",
			clientKey:      "newKey",
			caCert:         "newCert",
			token:          invalidToken,
			expectedConfig: bootstrap.Config{},
			identifyErr:    svcerr.ErrAuthentication,
			err:            svcerr.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, tc.identifyErr)
		svcCall := boot.On("UpdateCert", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.expectedConfig, tc.updateErr)

		cfg, err := svc.UpdateCert(context.Background(), tc.token, tc.thingID, tc.clientCert, tc.clientKey, tc.caCert)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		sort.Slice(cfg.Channels, func(i, j int) bool {
			return cfg.Channels[i].ID < cfg.Channels[j].ID
		})
		sort.Slice(tc.expectedConfig.Channels, func(i, j int) bool {
			return tc.expectedConfig.Channels[i].ID < tc.expectedConfig.Channels[j].ID
		})
		assert.Equal(t, tc.expectedConfig, cfg, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.expectedConfig, cfg))
		authCall.Unset()
		svcCall.Unset()
	}
}

func TestUpdateConnections(t *testing.T) {
	svc, boot, auth, sdk := newService()
	c := config

	ch := channel

	authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: validToken}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, nil)
	authCall1 := auth.On("Authorize", context.Background(), mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	sdkCall := sdk.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: config.ThingID, Credentials: mgsdk.Credentials{Secret: config.ThingKey}}, nil)
	sdkCall1 := sdk.On("Channel", mock.Anything, mock.Anything).Return(mgsdk.Channel{}, nil)
	svcCall := boot.On("ListExisting", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(config.Channels, nil)
	svcCall1 := boot.On("Save", context.Background(), mock.Anything, mock.Anything).Return(mock.Anything, nil)

	created, err := svc.Add(context.Background(), validToken, config)
	assert.Nil(t, err, fmt.Sprintf("Saving config expected to succeed: %s.\n", err))
	authCall.Unset()
	authCall1.Unset()
	sdkCall.Unset()
	sdkCall1.Unset()
	svcCall.Unset()
	svcCall1.Unset()

	c.ExternalID = testsutil.GenerateUUID(t)
	authCall = auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: validToken}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, nil)
	authCall1 = auth.On("Authorize", context.Background(), mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	sdkCall = sdk.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: config.ThingID, Credentials: mgsdk.Credentials{Secret: config.ThingKey}}, nil)
	sdkCall1 = sdk.On("Channel", mock.Anything, mock.Anything).Return(toGroup(c.Channels[0]), nil)
	svcCall = boot.On("ListExisting", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(config.Channels, nil)
	svcCall1 = boot.On("Save", context.Background(), mock.Anything, mock.Anything).Return(mock.Anything, nil)
	active, err := svc.Add(context.Background(), validToken, c)
	assert.Nil(t, err, fmt.Sprintf("Saving config expected to succeed: %s.\n", err))
	authCall.Unset()
	authCall1.Unset()
	sdkCall.Unset()
	sdkCall1.Unset()
	svcCall.Unset()
	svcCall1.Unset()

	authCall = auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: validToken}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, nil)
	authCall1 = auth.On("Authorize", context.Background(), mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	sdkCall = sdk.On("Connect", mock.Anything, mock.Anything).Return(nil)
	svcCall = boot.On("RetrieveByID", context.Background(), mock.Anything, mock.Anything).Return(config, nil)
	svcCall1 = boot.On("ChangeState", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = svc.ChangeState(context.Background(), validToken, active.ThingID, bootstrap.Active)
	assert.Nil(t, err, fmt.Sprintf("Changing state expected to succeed: %s.\n", err))
	authCall.Unset()
	authCall1.Unset()
	sdkCall.Unset()
	svcCall.Unset()
	svcCall1.Unset()

	nonExisting := config
	nonExisting.ThingID = unknown

	cases := []struct {
		desc        string
		token       string
		id          string
		connections []string
		identifyErr error
		updateErr   error
		thingErr    error
		channelErr  error
		retrieveErr error
		listErr     error
		err         error
	}{
		{
			desc:        "update connections for config with state Inactive",
			token:       validToken,
			id:          created.ThingID,
			connections: []string{ch.ID},
			err:         nil,
		},
		{
			desc:        "update connections for config with state Active",
			token:       validToken,
			id:          active.ThingID,
			connections: []string{ch.ID},
			err:         nil,
		},
		{
			desc:        "update connections for non-existing config",
			token:       validToken,
			id:          "",
			connections: []string{"3"},
			retrieveErr: errors.NewSDKError(svcerr.ErrNotFound),
			err:         svcerr.ErrNotFound,
		},
		{
			desc:        "update connections with invalid channels",
			token:       validToken,
			id:          created.ThingID,
			connections: []string{"wrong"},
			channelErr:  errors.NewSDKError(svcerr.ErrNotFound),
			err:         svcerr.ErrNotFound,
		},
		{
			desc:        "update connections a config with wrong credentials",
			token:       invalidToken,
			id:          created.ThingID,
			connections: []string{ch.ID, "3"},
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, tc.identifyErr)
		sdkCall := sdk.On("Thing", tc.id, tc.token).Return(mgsdk.Thing{ID: config.ThingID, Credentials: mgsdk.Credentials{Secret: config.ThingKey}}, tc.thingErr)
		sdkCall1 := sdk.On("Channel", mock.Anything, tc.token).Return(mgsdk.Channel{}, tc.channelErr)
		svcCall := boot.On("RetrieveByID", context.Background(), mock.Anything, mock.Anything).Return(config, tc.retrieveErr)
		svcCall1 := boot.On("ListExisting", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(config.Channels, tc.listErr)
		svcCall2 := boot.On("UpdateConnections", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.updateErr)
		err := svc.UpdateConnections(context.Background(), tc.token, tc.id, tc.connections)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		authCall.Unset()
		sdkCall.Unset()
		sdkCall1.Unset()
		svcCall.Unset()
		svcCall1.Unset()
		svcCall2.Unset()
	}
}

func TestList(t *testing.T) {
	svc, boot, auth, sdk := newService()
	numThings := 101
	var saved []bootstrap.Config
	for i := 0; i < numThings; i++ {
		c := config
		c.ExternalID = testsutil.GenerateUUID(t)
		c.ExternalKey = testsutil.GenerateUUID(t)
		c.Name = fmt.Sprintf("%s-%d", config.Name, i)

		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: validToken}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, nil)
		authCall1 := auth.On("Authorize", context.Background(), mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
		sdkCall := sdk.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: c.ThingID, Credentials: mgsdk.Credentials{Secret: c.ThingKey}}, nil)
		sdkCall1 := sdk.On("Channel", mock.Anything, mock.Anything).Return(toGroup(c.Channels[0]), nil)
		svcCall := boot.On("ListExisting", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(c.Channels, nil)
		svcCall1 := boot.On("Save", context.Background(), mock.Anything, mock.Anything).Return(mock.Anything, nil)
		s, err := svc.Add(context.Background(), validToken, c)
		assert.Nil(t, err, fmt.Sprintf("Saving config expected to succeed: %s.\n", err))
		authCall.Unset()
		authCall1.Unset()
		sdkCall.Unset()
		sdkCall1.Unset()
		svcCall.Unset()
		svcCall1.Unset()
		saved = append(saved, s)
	}
	// Set one Thing to the different state
	authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: validToken}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, nil)
	authCall1 := auth.On("Authorize", context.Background(), mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	sdkCall := sdk.On("Connect", mock.Anything, mock.Anything).Return(nil)
	svcCall := boot.On("RetrieveByID", context.Background(), mock.Anything, mock.Anything).Return(config, nil)
	svcCall1 := boot.On("ChangeState", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := svc.ChangeState(context.Background(), validToken, saved[41].ThingID, bootstrap.Active)
	assert.Nil(t, err, fmt.Sprintf("Changing config state expected to succeed: %s.\n", err))
	authCall.Unset()
	authCall1.Unset()
	sdkCall.Unset()
	svcCall.Unset()
	svcCall1.Unset()

	saved[41].State = bootstrap.Active

	cases := []struct {
		desc         string
		config       bootstrap.ConfigsPage
		filter       bootstrap.Filter
		offset       uint64
		limit        uint64
		token        string
		authorizeErr error
		identifyErr  error
		retrieveErr  error
		err          error
	}{
		{
			desc: "list configs",
			config: bootstrap.ConfigsPage{
				Total:   uint64(len(saved)),
				Offset:  0,
				Limit:   10,
				Configs: saved[0:10],
			},
			filter: bootstrap.Filter{},
			token:  validToken,
			offset: 0,
			limit:  10,
			err:    nil,
		},
		{
			desc: "list configs with specified name",
			config: bootstrap.ConfigsPage{
				Total:   1,
				Offset:  0,
				Limit:   100,
				Configs: saved[95:96],
			},
			filter: bootstrap.Filter{PartialMatch: map[string]string{"name": "95"}},
			token:  validToken,
			offset: 0,
			limit:  100,
			err:    nil,
		},
		{
			desc:        "list configs with invalid token",
			config:      bootstrap.ConfigsPage{},
			filter:      bootstrap.Filter{},
			token:       invalidToken,
			offset:      0,
			limit:       10,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc: "list last page",
			config: bootstrap.ConfigsPage{
				Total:   uint64(len(saved)),
				Offset:  95,
				Limit:   10,
				Configs: saved[95:],
			},
			filter: bootstrap.Filter{},
			token:  validToken,
			offset: 95,
			limit:  10,
			err:    nil,
		},
		{
			desc: "list configs with Active state",
			config: bootstrap.ConfigsPage{
				Total:   1,
				Offset:  35,
				Limit:   20,
				Configs: []bootstrap.Config{saved[41]},
			},
			filter: bootstrap.Filter{FullMatch: map[string]string{"state": bootstrap.Active.String()}},
			token:  validToken,
			offset: 35,
			limit:  20,
			err:    nil,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, tc.identifyErr)
		svcCall := boot.On("RetrieveAll", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.config, tc.retrieveErr)

		result, err := svc.List(context.Background(), tc.token, tc.filter, tc.offset, tc.limit)
		assert.ElementsMatch(t, tc.config.Configs, result.Configs, fmt.Sprintf("%s: expected %v got %v", tc.desc, tc.config.Configs, result.Configs))
		assert.Equal(t, tc.config.Total, result.Total, fmt.Sprintf("%s: expected %v got %v", tc.desc, tc.config.Total, result.Total))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		authCall.Unset()
		svcCall.Unset()
	}
}

func TestRemove(t *testing.T) {
	svc, boot, auth, _ := newService()
	cases := []struct {
		desc        string
		id          string
		token       string
		identifyErr error
		removeErr   error
		err         error
	}{
		{
			desc:        "view a config with wrong credentials",
			id:          config.ThingID,
			token:       invalidToken,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc:  "remove an existing config",
			id:    config.ThingID,
			token: validToken,
			err:   nil,
		},
		{
			desc:  "remove removed config",
			id:    config.ThingID,
			token: validToken,
			err:   nil,
		},
		{
			desc:  "remove non-existing config",
			id:    unknown,
			token: validToken,
			err:   nil,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, tc.identifyErr)
		svcCall := boot.On("Remove", context.Background(), mock.Anything, mock.Anything).Return(tc.err)
		err := svc.Remove(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		authCall.Unset()
		svcCall.Unset()
	}
}

func TestBootstrap(t *testing.T) {
	svc, boot, _, _ := newService()
	e, err := enc([]byte(config.ExternalKey))
	assert.Nil(t, err, fmt.Sprintf("Encrypting external key expected to succeed: %s.\n", err))

	cases := []struct {
		desc        string
		config      bootstrap.Config
		externalKey string
		externalID  string
		err         error
		encrypted   bool
	}{
		{
			desc:        "bootstrap using invalid external id",
			config:      bootstrap.Config{},
			externalID:  "invalid",
			externalKey: config.ExternalKey,
			err:         svcerr.ErrNotFound,
			encrypted:   false,
		},
		{
			desc:        "bootstrap using invalid external key",
			config:      bootstrap.Config{},
			externalID:  config.ExternalID,
			externalKey: "invalid",
			err:         bootstrap.ErrExternalKey,
			encrypted:   false,
		},
		{
			desc:        "bootstrap an existing config",
			config:      config,
			externalID:  config.ExternalID,
			externalKey: config.ExternalKey,
			err:         nil,
			encrypted:   false,
		},
		{
			desc:        "bootstrap encrypted",
			config:      config,
			externalID:  config.ExternalID,
			externalKey: hex.EncodeToString(e),
			err:         nil,
			encrypted:   true,
		},
	}

	for _, tc := range cases {
		svcCall := boot.On("RetrieveByExternalID", context.Background(), mock.Anything).Return(tc.config, tc.err)
		config, err := svc.Bootstrap(context.Background(), tc.externalKey, tc.externalID, tc.encrypted)
		assert.Equal(t, tc.config, config, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.config, config))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		svcCall.Unset()
	}
}

func TestChangeState(t *testing.T) {
	svc, boot, auth, sdk := newService()
	cases := []struct {
		desc          string
		state         bootstrap.State
		id            string
		token         string
		identifyErr   error
		retrieveErr   error
		connectErr    error
		disconenctErr error
		stateErr      error
		err           error
	}{
		{
			desc:        "change state with wrong credentials",
			state:       bootstrap.Active,
			id:          config.ThingID,
			token:       invalidToken,
			identifyErr: svcerr.ErrAuthentication,
			err:         svcerr.ErrAuthentication,
		},
		{
			desc:        "change state of non-existing config",
			state:       bootstrap.Active,
			id:          unknown,
			token:       validToken,
			retrieveErr: svcerr.ErrNotFound,
			err:         svcerr.ErrNotFound,
		},
		{
			desc:  "change state to Active",
			state: bootstrap.Active,
			id:    config.ThingID,
			token: validToken,
			err:   nil,
		},
		{
			desc:  "change state to current state",
			state: bootstrap.Active,
			id:    config.ThingID,
			token: validToken,
			err:   nil,
		},
		{
			desc:  "change state to Inactive",
			state: bootstrap.Inactive,
			id:    config.ThingID,
			token: validToken,
			err:   nil,
		},
		{
			desc:     "change state with invalid state",
			state:    bootstrap.State(2),
			id:       saved.ThingID,
			token:    validToken,
			stateErr: svcerr.ErrMalformedEntity,
			err:      svcerr.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		authCall := auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID, DomainId: domainID}, tc.identifyErr)
		sdkCall := sdk.On("Connect", mock.Anything, mock.Anything).Return(tc.connectErr)
		sdkCall1 := sdk.On("DisconnectThing", mock.Anything, mock.Anything, mock.Anything).Return(tc.disconenctErr)
		svcCall := boot.On("RetrieveByID", context.Background(), mock.Anything, mock.Anything).Return(config, tc.retrieveErr)
		svcCall1 := boot.On("ChangeState", context.Background(), mock.Anything, mock.Anything, mock.Anything).Return(tc.stateErr)

		err := svc.ChangeState(context.Background(), tc.token, tc.id, tc.state)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		authCall.Unset()
		sdkCall.Unset()
		sdkCall1.Unset()
		svcCall.Unset()
		svcCall1.Unset()
	}
}

func TestUpdateChannelHandler(t *testing.T) {
	svc, boot, _, _ := newService()
	ch := bootstrap.Channel{
		ID:       channel.ID,
		Name:     "new name",
		Metadata: map[string]interface{}{"meta": "new"},
	}

	cases := []struct {
		desc    string
		channel bootstrap.Channel
		err     error
	}{
		{
			desc:    "update an existing channel",
			channel: ch,
			err:     nil,
		},
		{
			desc:    "update a non-existing channel",
			channel: bootstrap.Channel{ID: ""},
			err:     nil,
		},
	}

	for _, tc := range cases {
		svcCall := boot.On("UpdateChannel", context.Background(), mock.Anything).Return(tc.err)
		err := svc.UpdateChannelHandler(context.Background(), tc.channel)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		svcCall.Unset()
	}
}

func TestRemoveChannelHandler(t *testing.T) {
	svc, boot, _, _ := newService()
	cases := []struct {
		desc string
		id   string
		err  error
	}{
		{
			desc: "remove an existing channel",
			id:   config.Channels[0].ID,
			err:  nil,
		},
		{
			desc: "remove a non-existing channel",
			id:   "unknown",
			err:  nil,
		},
	}

	for _, tc := range cases {
		svcCall := boot.On("RemoveChannel", context.Background(), mock.Anything).Return(tc.err)
		err := svc.RemoveChannelHandler(context.Background(), tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		svcCall.Unset()
	}
}

func TestRemoveConfigHandler(t *testing.T) {
	svc, boot, _, _ := newService()
	cases := []struct {
		desc string
		id   string
		err  error
	}{
		{
			desc: "remove an existing config",
			id:   config.ThingID,
			err:  nil,
		},
		{
			desc: "remove a non-existing channel",
			id:   "unknown",
			err:  nil,
		},
	}

	for _, tc := range cases {
		svcCall := boot.On("RemoveThing", context.Background(), mock.Anything).Return(tc.err)
		err := svc.RemoveConfigHandler(context.Background(), tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		svcCall.Unset()
	}
}

func TestConnectThingsHandler(t *testing.T) {
	svc, boot, _, _ := newService()
	cases := []struct {
		desc      string
		thingID   string
		channelID string
		err       error
	}{
		{
			desc:      "connect",
			channelID: channel.ID,
			thingID:   config.ThingID,
			err:       nil,
		},
		{
			desc:      "connect connected",
			channelID: channel.ID,
			thingID:   config.ThingID,
			err:       svcerr.ErrAddPolicies,
		},
	}

	for _, tc := range cases {
		repoCall := boot.On("ConnectThing", context.Background(), mock.Anything, mock.Anything).Return(tc.err)
		err := svc.ConnectThingHandler(context.Background(), tc.channelID, tc.thingID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
	}
}

func TestDisconnectThingsHandler(t *testing.T) {
	svc, boot, _, _ := newService()
	cases := []struct {
		desc      string
		thingID   string
		channelID string
		err       error
	}{
		{
			desc:      "disconnect",
			channelID: channel.ID,
			thingID:   config.ThingID,
			err:       nil,
		},
		{
			desc:      "disconnect disconnected",
			channelID: channel.ID,
			thingID:   config.ThingID,
			err:       nil,
		},
	}

	for _, tc := range cases {
		svcCall := boot.On("DisconnectThing", context.Background(), mock.Anything, mock.Anything).Return(tc.err)
		err := svc.DisconnectThingHandler(context.Background(), tc.channelID, tc.thingID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		svcCall.Unset()
	}
}

func toGroup(ch bootstrap.Channel) mgsdk.Channel {
	return mgsdk.Channel{
		ID:       ch.ID,
		Name:     ch.Name,
		Metadata: ch.Metadata,
	}
}

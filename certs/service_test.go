// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package certs_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"time"

	"github.com/absmach/magistrala"
	authmocks "github.com/absmach/magistrala/auth/mocks"
	"github.com/absmach/magistrala/certs"
	"github.com/absmach/magistrala/certs/mocks"
	"github.com/absmach/magistrala/certs/pki"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	mgsdk "github.com/absmach/magistrala/pkg/sdk/go"
	sdkmocks "github.com/absmach/magistrala/pkg/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	invalid   = "invalid"
	email     = "user@example.com"
	token     = "token"
	thingsNum = 1
	thingKey  = "thingKey"
	thingID   = "1"
	ttl       = "1h"
	certNum   = 10
	validID   = "d4ebb847-5d0e-4e46-bdd9-b6aceaaa3a22"
)

type Mocks struct {
	Repo  *mocks.Repository
	Agent *mocks.Agent
	Auth  *authmocks.AuthClient
	SDK   *sdkmocks.SDK
}

func newService(_ *testing.T) (certs.Service, Mocks) {
	testMocks := Mocks{
		Repo:  new(mocks.Repository),
		Agent: new(mocks.Agent),
		Auth:  new(authmocks.AuthClient),
		SDK:   new(sdkmocks.SDK),
	}

	return certs.New(testMocks.Auth, testMocks.Repo, testMocks.SDK, testMocks.Agent), testMocks
}

func TestIssueCert(t *testing.T) {
	svc, testMocks := newService(t)
	cases := []struct {
		token   string
		desc    string
		thingID string
		ttl     string
		key     string
		err     error
	}{
		{
			desc:    "issue new cert",
			token:   token,
			thingID: thingID,
			ttl:     ttl,
			err:     nil,
		},
		{
			desc:    "issue new cert for non existing thing id",
			token:   token,
			thingID: "2",
			ttl:     ttl,
			err:     certs.ErrFailedCertCreation,
		},
		{
			desc:    "issue new cert for non existing thing id",
			token:   invalid,
			thingID: thingID,
			ttl:     ttl,
			err:     svcerr.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := testMocks.Auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, tc.err)
		repoCall2 := testMocks.SDK.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: tc.thingID, Credentials: mgsdk.Credentials{Secret: thingKey}}, errors.NewSDKError(tc.err))
		repoCall3 := testMocks.Agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, tc.err)
		repoCall4 := testMocks.Repo.On("Save", mock.Anything, mock.Anything).Return("", tc.err)

		c, err := svc.IssueCert(context.Background(), tc.token, tc.thingID, tc.ttl)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		cert, _ := certs.ReadCert([]byte(c.ClientCert))
		if cert != nil {
			assert.True(t, strings.Contains(cert.Subject.CommonName, thingKey), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, thingKey, cert.Subject.CommonName))
		}
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
		repoCall3.Unset()
		repoCall4.Unset()
	}
}

func TestRevokeCert(t *testing.T) {
	svc, testMocks := newService(t)

	repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
	repoCall1 := testMocks.Auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	repoCall2 := testMocks.SDK.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: thingID, Credentials: mgsdk.Credentials{Secret: thingKey}}, nil)
	repoCall3 := testMocks.Agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, nil)
	repoCall4 := testMocks.Repo.On("Save", mock.Anything, mock.Anything).Return("", nil)

	_, err := svc.IssueCert(context.Background(), token, thingID, ttl)
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))
	repoCall.Unset()
	repoCall1.Unset()
	repoCall2.Unset()
	repoCall3.Unset()
	repoCall4.Unset()

	cases := []struct {
		token   string
		desc    string
		thingID string
		err     error
	}{
		{
			desc:    "revoke cert",
			token:   token,
			thingID: thingID,
			err:     nil,
		},
		{
			desc:    "revoke cert for invalid token",
			token:   invalid,
			thingID: thingID,
			err:     svcerr.ErrAuthentication,
		},
		{
			desc:    "revoke cert for invalid thing id",
			token:   token,
			thingID: "2",
			err:     certs.ErrFailedCertRevocation,
		},
	}

	for _, tc := range cases {
		repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := testMocks.Auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, tc.err)
		repoCall2 := testMocks.SDK.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: tc.thingID, Credentials: mgsdk.Credentials{Secret: thingKey}}, errors.NewSDKError(tc.err))
		repoCall3 := testMocks.Repo.On("RetrieveByThing", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(certs.Page{}, tc.err)

		_, err := svc.RevokeCert(context.Background(), tc.token, tc.thingID)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
		repoCall3.Unset()
	}
}

func TestListCerts(t *testing.T) {
	svc, testMocks := newService(t)
	var mycerts []certs.Cert
	for i := 0; i < certNum; i++ {
		newCert := certs.Cert{
			OwnerID: validID,
			ThingID: thingID,
			Serial:  fmt.Sprintf("%d", i),
			Expire:  time.Now().Add(time.Hour),
		}
		mycerts = append(mycerts, newCert)
	}

	cases := []struct {
		token   string
		desc    string
		thingID string
		page    certs.Page
		err     error
	}{
		{
			desc:    "list all certs with valid token",
			token:   token,
			thingID: thingID,
			page:    certs.Page{Limit: certNum, Offset: 0, Total: certNum, Certs: mycerts},
			err:     nil,
		},
		{
			desc:    "list all certs with invalid token",
			token:   invalid,
			thingID: thingID,
			page:    certs.Page{},
			err:     svcerr.ErrAuthentication,
		},
		{
			desc:    "list half certs with valid token",
			token:   token,
			thingID: thingID,
			page:    certs.Page{Limit: certNum, Offset: certNum / 2, Total: certNum / 2, Certs: mycerts[certNum/2:]},
			err:     nil,
		},
		{
			desc:    "list last cert with valid token",
			token:   token,
			thingID: thingID,
			page:    certs.Page{Limit: certNum, Offset: certNum - 1, Total: 1, Certs: []certs.Cert{mycerts[certNum-1]}},
			err:     nil,
		},
	}

	for _, tc := range cases {
		repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := testMocks.Repo.On("RetrieveByThing", context.Background(), validID, thingID, tc.page.Offset, tc.page.Limit).Return(tc.page, tc.err)
		repoCall2 := testMocks.Agent.On("Read", mock.Anything).Return(pki.Cert{}, tc.err)

		page, err := svc.ListCerts(context.Background(), tc.token, tc.thingID, tc.page.Offset, tc.page.Limit)
		size := uint64(len(page.Certs))
		assert.Equal(t, tc.page.Total, size, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.page.Total, size))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
	}
}

func TestListSerials(t *testing.T) {
	svc, testMocks := newService(t)

	var issuedCerts []certs.Cert
	for i := 0; i < certNum; i++ {
		repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := testMocks.Auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
		repoCall2 := testMocks.SDK.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: thingID, Credentials: mgsdk.Credentials{Secret: thingKey}}, nil)
		repoCall3 := testMocks.Agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, nil)
		repoCall4 := testMocks.Repo.On("Save", mock.Anything, mock.Anything).Return("", nil)

		cert, err := svc.IssueCert(context.Background(), token, thingID, ttl)
		assert.Nil(t, err, fmt.Sprintf("unexpected cert creation error: %s\n", err))
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
		repoCall3.Unset()
		repoCall4.Unset()

		crt := certs.Cert{
			OwnerID: cert.OwnerID,
			ThingID: cert.ThingID,
			Serial:  cert.Serial,
			Expire:  cert.Expire,
		}
		issuedCerts = append(issuedCerts, crt)
	}

	cases := []struct {
		token   string
		desc    string
		thingID string
		offset  uint64
		limit   uint64
		certs   []certs.Cert
		err     error
	}{
		{
			desc:    "list all certs with valid token",
			token:   token,
			thingID: thingID,
			offset:  0,
			limit:   certNum,
			certs:   issuedCerts,
			err:     nil,
		},
		{
			desc:    "list all certs with invalid token",
			token:   invalid,
			thingID: thingID,
			offset:  0,
			limit:   certNum,
			certs:   nil,
			err:     svcerr.ErrAuthentication,
		},
		{
			desc:    "list half certs with valid token",
			token:   token,
			thingID: thingID,
			offset:  certNum / 2,
			limit:   certNum,
			certs:   issuedCerts[certNum/2:],
			err:     nil,
		},
		{
			desc:    "list last cert with valid token",
			token:   token,
			thingID: thingID,
			offset:  certNum - 1,
			limit:   certNum,
			certs:   []certs.Cert{issuedCerts[certNum-1]},
			err:     nil,
		},
	}

	for _, tc := range cases {
		repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := testMocks.Repo.On("RetrieveByThing", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(certs.Page{Limit: tc.limit, Offset: tc.offset, Total: certNum, Certs: tc.certs}, tc.err)

		page, err := svc.ListSerials(context.Background(), tc.token, tc.thingID, tc.offset, tc.limit)
		assert.Equal(t, tc.certs, page.Certs, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.certs, page.Certs))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
		repoCall1.Unset()
	}
}

func TestViewCert(t *testing.T) {
	svc, testMocks := newService(t)

	repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
	repoCall1 := testMocks.Auth.On("Authorize", mock.Anything, mock.Anything).Return(&magistrala.AuthorizeRes{Authorized: true}, nil)
	repoCall2 := testMocks.SDK.On("Thing", mock.Anything, mock.Anything).Return(mgsdk.Thing{ID: thingID, Credentials: mgsdk.Credentials{Secret: thingKey}}, nil)
	repoCall3 := testMocks.Agent.On("IssueCert", mock.Anything, mock.Anything).Return(pki.Cert{}, nil)
	repoCall4 := testMocks.Repo.On("Save", mock.Anything, mock.Anything).Return("", nil)

	ic, err := svc.IssueCert(context.Background(), token, thingID, ttl)
	require.Nil(t, err, fmt.Sprintf("unexpected cert creation error: %s\n", err))
	repoCall.Unset()
	repoCall1.Unset()
	repoCall2.Unset()
	repoCall3.Unset()
	repoCall4.Unset()

	cert := certs.Cert{
		ThingID:    thingID,
		ClientCert: ic.ClientCert,
		Serial:     ic.Serial,
		Expire:     ic.Expire,
	}

	cases := []struct {
		token    string
		desc     string
		serialID string
		cert     certs.Cert
		err      error
	}{
		{
			desc:     "list cert with valid token and serial",
			token:    token,
			serialID: cert.Serial,
			cert:     cert,
			err:      nil,
		},
		{
			desc:     "list cert with invalid token",
			token:    invalid,
			serialID: cert.Serial,
			cert:     certs.Cert{},
			err:      svcerr.ErrAuthentication,
		},
		{
			desc:     "list cert with invalid serial",
			token:    token,
			serialID: invalid,
			cert:     certs.Cert{},
			err:      svcerr.ErrNotFound,
		},
	}

	for _, tc := range cases {
		repoCall := testMocks.Auth.On("Identify", mock.Anything, &magistrala.IdentityReq{Token: tc.token}).Return(&magistrala.IdentityRes{Id: validID}, nil)
		repoCall1 := testMocks.Repo.On("RetrieveBySerial", context.Background(), mock.Anything, mock.Anything).Return(tc.cert, tc.err)
		repoCall2 := testMocks.Agent.On("Read", mock.Anything).Return(pki.Cert{}, tc.err)

		cert, err := svc.ViewCert(context.Background(), tc.token, tc.serialID)
		assert.Equal(t, tc.cert, cert, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.cert, cert))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
	}
}

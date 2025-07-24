// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package certs_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/absmach/supermq/certs"
	"github.com/absmach/supermq/certs/mocks"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	mgsdk "github.com/absmach/supermq/pkg/sdk"
	sdkmocks "github.com/absmach/supermq/pkg/sdk/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	invalid    = "invalid"
	email      = "user@example.com"
	domain     = "domain"
	token      = "token"
	clientsNum = 1
	clientKey  = "clientKey"
	clientID   = "1"
	ttl        = "1h"
	certNum    = 10
	validID    = "d4ebb847-5d0e-4e46-bdd9-b6aceaaa3a22"
)

func newService(_ *testing.T) (certs.Service, *mocks.Agent, *sdkmocks.SDK, *mocks.Repository) {
	agent := new(mocks.Agent)
	repo := new(mocks.Repository)
	sdk := new(sdkmocks.SDK)

	return certs.New(sdk, repo, agent), agent, sdk, repo
}

var cert = certs.Cert{
	ClientID:     clientID,
	SerialNumber: "Serial",
	ExpiryTime:   time.Now().Add(time.Duration(1000)),
}

func TestIssueCert(t *testing.T) {
	svc, agent, sdk, repo := newService(t)
	cases := []struct {
		domainID     string
		token        string
		desc         string
		clientID     string
		ttl          string
		ipAddr       []string
		key          string
		cert         certs.Cert
		clientErr    errors.SDKError
		issueCertErr error
		saveErr      error
		err          error
	}{
		{
			desc:     "issue new cert",
			domainID: domain,
			token:    token,
			clientID: clientID,
			ttl:      ttl,
			ipAddr:   []string{},
			cert:     cert,
		},
		{
			desc:         "issue new for failed pki",
			domainID:     domain,
			token:        token,
			clientID:     clientID,
			ttl:          ttl,
			ipAddr:       []string{},
			clientErr:    nil,
			issueCertErr: certs.ErrFailedCertCreation,
			err:          certs.ErrFailedCertCreation,
		},
		{
			desc:      "issue new cert for non existing client id",
			domainID:  domain,
			token:     token,
			clientID:  "2",
			ttl:       ttl,
			ipAddr:    []string{},
			clientErr: errors.NewSDKError(errors.ErrMalformedEntity),
			err:       certs.ErrFailedCertCreation,
		},
		{
			desc:      "issue new cert for invalid token",
			domainID:  domain,
			token:     invalid,
			clientID:  clientID,
			ttl:       ttl,
			ipAddr:    []string{},
			clientErr: errors.NewSDKError(svcerr.ErrAuthentication),
			err:       svcerr.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sdkCall := sdk.On("Client", mock.Anything, tc.clientID, tc.domainID, tc.token).Return(mgsdk.Client{ID: tc.clientID, Credentials: mgsdk.ClientCredentials{Secret: clientKey}}, tc.clientErr)
			agentCall := agent.On("Issue", clientID, tc.ttl, tc.ipAddr).Return(tc.cert, tc.issueCertErr)
			repoCall := repo.On("Save", mock.Anything, tc.cert).Return("", tc.saveErr)
			resp, err := svc.IssueCert(context.Background(), tc.domainID, tc.token, tc.clientID, tc.ttl)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			assert.Equal(t, tc.cert.SerialNumber, resp.SerialNumber, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.cert.SerialNumber, resp.SerialNumber))
			sdkCall.Unset()
			agentCall.Unset()
			repoCall.Unset()
		})
	}
}

func TestRevokeCert(t *testing.T) {
	svc, agent, _, repo := newService(t)
	cases := []struct {
		domainID    string
		token       string
		desc        string
		clientID    string
		page        certs.CertPage
		authErr     error
		clientErr   errors.SDKError
		revokeErr   error
		listErr     error
		retrieveErr error
		updateErr   error
		err         error
	}{
		{
			desc:     "revoke cert",
			domainID: domain,
			token:    token,
			clientID: clientID,
			page:     certs.CertPage{Limit: 10000, Offset: 0, Total: 1, Certificates: []certs.Cert{cert}},
		},
		{
			desc:      "revoke cert for failed pki revoke",
			domainID:  domain,
			token:     token,
			clientID:  clientID,
			page:      certs.CertPage{Limit: 10000, Offset: 0, Total: 1, Certificates: []certs.Cert{cert}},
			revokeErr: certs.ErrFailedCertRevocation,
			err:       certs.ErrFailedCertRevocation,
		},
		{
			desc:        "revoke cert with failed to list certs",
			domainID:    domain,
			token:       token,
			clientID:    clientID,
			page:        certs.CertPage{},
			retrieveErr: certs.ErrFailedCertRevocation,
			listErr:     certs.ErrFailedCertRevocation,
			err:         certs.ErrFailedCertRevocation,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			repoCall := repo.On("RetrieveByClient", mock.Anything, tc.clientID, mock.Anything).Return(tc.page, tc.retrieveErr)
			repoCall1 := repo.On("Update", mock.Anything, mock.Anything).Return(tc.updateErr)
			agentCall := agent.On("Revoke", mock.Anything).Return(tc.revokeErr)
			agentCall1 := agent.On("ListCerts", mock.Anything).Return(tc.page, tc.listErr)
			_, err := svc.RevokeCert(context.Background(), tc.domainID, tc.token, tc.clientID)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			repoCall.Unset()
			repoCall1.Unset()
			agentCall.Unset()
			agentCall1.Unset()
		})
	}
}

func TestRevokeBySerial(t *testing.T) {
	svc, agent, _, repo := newService(t)
	cases := []struct {
		desc         string
		serialID     string
		revokeErr    error
		updateErr    error
		retrieveErr  error
		Cert         certs.Cert
		expectedTime time.Time
		err          error
	}{
		{
			desc:         "revoke cert by serial successfully",
			serialID:     cert.SerialNumber,
			expectedTime: time.Now(),
			Cert:         certs.Cert{SerialNumber: cert.SerialNumber, ClientID: cert.ClientID, ExpiryTime: cert.ExpiryTime, Revoked: false},
		},
		{
			desc:      "revoke cert by serial with PKI revoke failure",
			serialID:  cert.SerialNumber,
			revokeErr: certs.ErrFailedCertRevocation,
			err:       certs.ErrFailedCertRevocation,
		},
		{
			desc:      "revoke cert by serial with repository remove failure",
			serialID:  cert.SerialNumber,
			updateErr: certs.ErrFailedReadFromDB,
			err:       certs.ErrFailedReadFromDB,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			agentCall := agent.On("Revoke", tc.serialID).Return(tc.revokeErr)
			repoCall := repo.On("Update", mock.Anything, mock.Anything).Return(tc.updateErr)
			repoCall1 := repo.On("RetrieveBySerial", mock.Anything, mock.Anything).Return(tc.Cert, tc.retrieveErr)

			result, err := svc.RevokeBySerial(context.Background(), tc.serialID)

			if tc.err != nil {
				assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			} else {
				assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
				assert.False(t, result.RevocationTime.IsZero(), fmt.Sprintf("%s: revocation time should be set", tc.desc))
				assert.True(t, time.Since(result.RevocationTime) < time.Minute, fmt.Sprintf("%s: revocation time should be recent", tc.desc))
			}

			agentCall.Unset()
			repoCall.Unset()
			repoCall1.Unset()
		})
	}
}

func TestListCerts(t *testing.T) {
	svc, agent, _, repo := newService(t)
	var mycerts []certs.Cert
	for i := 0; i < certNum; i++ {
		c := certs.Cert{
			ClientID:     clientID,
			SerialNumber: fmt.Sprintf("%d", i),
			ExpiryTime:   time.Now().Add(time.Hour),
		}
		mycerts = append(mycerts, c)
	}

	cases := []struct {
		desc        string
		clientID    string
		page        certs.CertPage
		listErr     error
		retrieveErr error
		err         error
	}{
		{
			desc:     "list all certs successfully",
			clientID: clientID,
			page:     certs.CertPage{Limit: certNum, Offset: 0, Total: certNum, Certificates: mycerts},
		},
		{
			desc:        "list all certs with failed pki",
			clientID:    clientID,
			page:        certs.CertPage{},
			retrieveErr: svcerr.ErrViewEntity,
			err:         svcerr.ErrViewEntity,
		},
		{
			desc:     "list half certs successfully",
			clientID: clientID,
			page:     certs.CertPage{Limit: certNum, Offset: certNum / 2, Total: certNum / 2, Certificates: mycerts[certNum/2:]},
		},
		{
			desc:     "list last cert successfully",
			clientID: clientID,
			page:     certs.CertPage{Limit: certNum, Offset: certNum - 1, Total: 1, Certificates: []certs.Cert{mycerts[certNum-1]}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			repoCall := repo.On("RetrieveByClient", mock.Anything, tc.clientID, mock.Anything, mock.Anything).Return(tc.page, tc.retrieveErr)
			agentCall := agent.On("View", mock.Anything).Return(certs.Cert{}, tc.listErr)
			page, err := svc.ListCerts(context.Background(), tc.clientID, certs.PageMetadata{Offset: tc.page.Offset, Limit: tc.page.Limit})
			size := uint64(len(page.Certificates))
			assert.Equal(t, tc.page.Total, size, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.page.Total, size))
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			repoCall.Unset()
			agentCall.Unset()
		})
	}
}

func TestListSerials(t *testing.T) {
	svc, _, _, repo := newService(t)

	var issuedCerts []certs.Cert
	for i := 0; i < certNum; i++ {
		crt := certs.Cert{
			ClientID:     cert.ClientID,
			SerialNumber: cert.SerialNumber,
			ExpiryTime:   cert.ExpiryTime,
		}
		issuedCerts = append(issuedCerts, crt)
	}

	cases := []struct {
		desc        string
		clientID    string
		offset      uint64
		limit       uint64
		certs       []certs.Cert
		retrieveErr error
		err         error
	}{
		{
			desc:     "list all certs successfully",
			clientID: clientID,
			offset:   0,
			limit:    certNum,
			certs:    issuedCerts,
		},
		{
			desc:        "list all certs with failed pki",
			clientID:    clientID,
			offset:      0,
			limit:       certNum,
			certs:       nil,
			retrieveErr: svcerr.ErrViewEntity,
			err:         svcerr.ErrViewEntity,
		},
		{
			desc:     "list half certs successfully",
			clientID: clientID,
			offset:   certNum / 2,
			limit:    certNum,
			certs:    issuedCerts[certNum/2:],
		},
		{
			desc:     "list last cert successfully",
			clientID: clientID,
			offset:   certNum - 1,
			limit:    certNum,
			certs:    []certs.Cert{issuedCerts[certNum-1]},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			repoCall := repo.On("RetrieveByClient", mock.Anything, tc.clientID, certs.PageMetadata{Offset: tc.offset, Limit: tc.limit}).Return(certs.CertPage{Certificates: tc.certs}, tc.retrieveErr)
			page, err := svc.ListSerials(context.Background(), tc.clientID, certs.PageMetadata{Offset: tc.offset, Limit: tc.limit})
			assert.Equal(t, len(tc.certs), len(page.Certificates), fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.certs, page.Certificates))
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			repoCall.Unset()
		})
	}
}

func TestViewCert(t *testing.T) {
	svc, agent, _, repo := newService(t)

	cases := []struct {
		desc     string
		serialID string
		cert     certs.Cert
		repoErr  error
		agentErr error
		err      error
	}{
		{
			desc:     "view cert with valid serial",
			serialID: cert.SerialNumber,
			cert:     cert,
		},
		{
			desc:     "list cert with invalid serial",
			serialID: invalid,
			cert:     certs.Cert{},
			agentErr: svcerr.ErrNotFound,
			err:      svcerr.ErrNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			repoCall := repo.On("RetrieveBySerial", mock.Anything, tc.serialID).Return(tc.cert, tc.repoErr)
			agentCall := agent.On("View", tc.serialID).Return(tc.cert, tc.agentErr)
			res, err := svc.ViewCert(context.Background(), tc.serialID)
			assert.Equal(t, tc.cert.SerialNumber, res.SerialNumber, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.cert.SerialNumber, res.SerialNumber))
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			repoCall.Unset()
			agentCall.Unset()
		})
	}
}

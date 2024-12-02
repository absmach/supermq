// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/absmach/magistrala/domains"
	httpapi "github.com/absmach/magistrala/domains/api/http"
	"github.com/absmach/magistrala/domains/mocks"
	internalapi "github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/internal/testsutil"
	mglog "github.com/absmach/magistrala/logger"
	"github.com/absmach/magistrala/pkg/apiutil"
	mgauthn "github.com/absmach/magistrala/pkg/authn"
	authnmocks "github.com/absmach/magistrala/pkg/authn/mocks"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	sdk "github.com/absmach/magistrala/pkg/sdk/go"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	authDomain, sdkDomain = generateTestDomain(&testing.T{})
	authDomainReq         = domains.Domain{
		Name:     authDomain.Name,
		Metadata: authDomain.Metadata,
		Tags:     authDomain.Tags,
		Alias:    authDomain.Alias,
	}
	sdkDomainReq = sdk.Domain{
		Name:     sdkDomain.Name,
		Metadata: sdkDomain.Metadata,
		Tags:     sdkDomain.Tags,
		Alias:    sdkDomain.Alias,
	}
	updatedDomianName = "updated-domain"
)

func setupDomains() (*httptest.Server, *mocks.Service, *authnmocks.Authentication) {
	svc := new(mocks.Service)
	logger := mglog.NewMock()
	mux := chi.NewRouter()
	authn := new(authnmocks.Authentication)

	mux = httpapi.MakeHandler(svc, authn, mux, logger, "")
	return httptest.NewServer(mux), svc, authn
}

func TestCreateDomain(t *testing.T) {
	ds, svc, auth := setupDomains()
	defer ds.Close()

	sdkConf := sdk.Config{
		DomainsURL:     ds.URL,
		MsgContentType: contentType,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc     string
		token    string
		session  mgauthn.Session
		domain   sdk.Domain
		svcReq   domains.Domain
		svcRes   domains.Domain
		svcErr   error
		authnErr error
		response sdk.Domain
		err      error
	}{
		{
			desc:     "create domain successfully",
			token:    validToken,
			domain:   sdkDomainReq,
			svcReq:   authDomainReq,
			svcRes:   authDomain,
			svcErr:   nil,
			response: sdkDomain,
			err:      nil,
		},
		{
			desc:     "create domain with invalid token",
			token:    invalidToken,
			domain:   sdkDomainReq,
			svcReq:   authDomainReq,
			svcRes:   domains.Domain{},
			authnErr: svcerr.ErrAuthentication,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "create domain with empty token",
			token:    "",
			domain:   sdkDomainReq,
			svcReq:   authDomainReq,
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:  "create domain with empty name",
			token: validToken,
			domain: sdk.Domain{
				Name:     "",
				Metadata: sdkDomain.Metadata,
				Tags:     sdkDomain.Tags,
				Alias:    sdkDomain.Alias,
			},
			svcReq:   domains.Domain{},
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrMissingName, http.StatusBadRequest),
		},
		{
			desc:  "create domain with request that cannot be marshalled",
			token: validToken,
			domain: sdk.Domain{
				Name: sdkDomain.Name,
				Metadata: sdk.Metadata{
					"key": make(chan int),
				},
			},
			svcReq:   domains.Domain{},
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:   "create domain with response that cannot be unmarshalled",
			token:  validToken,
			domain: sdkDomainReq,
			svcReq: authDomainReq,
			svcRes: domains.Domain{
				ID:   authDomain.ID,
				Name: authDomain.Name,
				Metadata: domains.Metadata{
					"key": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := auth.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authnErr)
			svcCall := svc.On("CreateDomain", mock.Anything, tc.session, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.CreateDomain(tc.domain, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "CreateDomain", mock.Anything, tc.session, tc.svcReq)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestUpdateDomain(t *testing.T) {
	ds, svc, authn := setupDomains()
	defer ds.Close()

	sdkConf := sdk.Config{
		DomainsURL:     ds.URL,
		MsgContentType: contentType,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	upDomainSDK := sdkDomain
	upDomainSDK.Name = updatedDomianName
	upDomainAuth := authDomain
	upDomainAuth.Name = updatedDomianName

	cases := []struct {
		desc     string
		token    string
		session  mgauthn.Session
		domainID string
		domain   sdk.Domain
		svcRes   domains.Domain
		svcErr   error
		authnErr error
		response sdk.Domain
		err      error
	}{
		{
			desc:     "update domain successfully",
			token:    validToken,
			domainID: sdkDomain.ID,
			domain: sdk.Domain{
				ID:   sdkDomain.ID,
				Name: updatedDomianName,
			},
			svcRes:   upDomainAuth,
			svcErr:   nil,
			response: upDomainSDK,
			err:      nil,
		},
		{
			desc:     "update domain with invalid token",
			token:    invalidToken,
			domainID: sdkDomain.ID,
			domain: sdk.Domain{
				ID:   sdkDomain.ID,
				Name: updatedDomianName,
			},
			svcRes:   domains.Domain{},
			authnErr: svcerr.ErrAuthentication,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "update domain with empty token",
			token:    "",
			domainID: sdkDomain.ID,
			domain: sdk.Domain{
				ID:   sdkDomain.ID,
				Name: updatedDomianName,
			},
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "update domain with invalid domain ID",
			token:    validToken,
			domainID: wrongID,
			domain: sdk.Domain{
				ID:   wrongID,
				Name: updatedDomianName,
			},
			svcRes:   domains.Domain{},
			svcErr:   svcerr.ErrAuthorization,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "update domain with empty id",
			token:    validToken,
			domainID: "",
			domain: sdk.Domain{
				Name: sdkDomain.Name,
			},
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKError(apiutil.ErrMissingID),
		},
		{
			desc:     "update domain with request that cannot be marshalled",
			token:    validToken,
			domainID: sdkDomain.ID,
			domain: sdk.Domain{
				ID:   sdkDomain.ID,
				Name: sdkDomain.Name,
				Metadata: sdk.Metadata{
					"key": make(chan int),
				},
			},
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:     "update domain with response that cannot be unmarshalled",
			token:    validToken,
			domainID: sdkDomain.ID,
			domain: sdk.Domain{
				ID:   sdkDomain.ID,
				Name: sdkDomain.Name,
			},
			svcRes: domains.Domain{
				ID:   authDomain.ID,
				Name: authDomain.Name,
				Metadata: domains.Metadata{
					"key": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: tc.domainID + "_" + validID, UserID: validID, DomainID: tc.domainID}
			}
			authCall := authn.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authnErr)
			svcCall := svc.On("UpdateDomain", mock.Anything, tc.session, tc.domainID, mock.Anything).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.UpdateDomain(tc.domain, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "UpdateDomain", mock.Anything, tc.session, tc.domainID, mock.Anything)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestViewDomain(t *testing.T) {
	ds, svc, authn := setupDomains()
	defer ds.Close()

	sdkConf := sdk.Config{
		DomainsURL:     ds.URL,
		MsgContentType: contentType,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc     string
		token    string
		session  mgauthn.Session
		domainID string
		svcRes   domains.Domain
		svcErr   error
		authnErr error
		response sdk.Domain
		err      error
	}{
		{
			desc:     "view domain successfully",
			token:    validToken,
			domainID: sdkDomain.ID,
			svcRes:   authDomain,
			svcErr:   nil,
			response: sdkDomain,
			err:      nil,
		},
		{
			desc:     "view domain with invalid token",
			token:    invalidToken,
			domainID: sdkDomain.ID,
			svcRes:   domains.Domain{},
			authnErr: svcerr.ErrAuthentication,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "view domain with empty token",
			token:    "",
			domainID: sdkDomain.ID,
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "view domain with invalid domain ID",
			token:    validToken,
			domainID: wrongID,
			svcRes:   domains.Domain{},
			svcErr:   svcerr.ErrAuthorization,
			response: sdk.Domain{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, http.StatusForbidden),
		},
		{
			desc:     "view domain with empty id",
			token:    validToken,
			domainID: "",
			svcRes:   domains.Domain{},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKError(apiutil.ErrMissingID),
		},
		{
			desc:     "view domain with response that cannot be unmarshalled",
			token:    validToken,
			domainID: sdkDomain.ID,
			svcRes: domains.Domain{
				ID:   authDomain.ID,
				Name: authDomain.Name,
				Metadata: domains.Metadata{
					"key": make(chan int),
				},
			},
			svcErr:   nil,
			response: sdk.Domain{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: tc.domainID + "_" + validID, UserID: validID, DomainID: tc.domainID}
			}
			authCall := authn.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authnErr)
			svcCall := svc.On("RetrieveDomain", mock.Anything, tc.session, tc.domainID).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.Domain(tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "RetrieveDomain", mock.Anything, tc.session, tc.domainID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestListDomians(t *testing.T) {
	ds, svc, authn := setupDomains()
	defer ds.Close()

	sdkConf := sdk.Config{
		DomainsURL:     ds.URL,
		MsgContentType: contentType,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc     string
		token    string
		session  mgauthn.Session
		pageMeta sdk.PageMetadata
		svcReq   domains.Page
		svcRes   domains.DomainsPage
		svcErr   error
		authnErr error
		response sdk.DomainsPage
		err      error
	}{
		{
			desc:  "list domains successfully",
			token: validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  10,
			},
			svcReq: domains.Page{
				Offset: 0,
				Limit:  10,
				Order:  internalapi.DefOrder,
				Dir:    internalapi.DefDir,
			},
			svcRes: domains.DomainsPage{
				Total:   1,
				Domains: []domains.Domain{authDomain},
			},
			svcErr: nil,
			response: sdk.DomainsPage{
				PageRes: sdk.PageRes{
					Total: 1,
				},
				Domains: []sdk.Domain{sdkDomain},
			},
			err: nil,
		},
		{
			desc:  "list domains with invalid token",
			token: invalidToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  10,
			},
			svcReq: domains.Page{
				Offset: 0,
				Limit:  10,
				Order:  internalapi.DefOrder,
				Dir:    internalapi.DefDir,
			},
			svcRes:   domains.DomainsPage{},
			authnErr: svcerr.ErrAuthentication,
			response: sdk.DomainsPage{},
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:  "list domains with empty token",
			token: "",
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  10,
			},
			svcReq:   domains.Page{},
			svcRes:   domains.DomainsPage{},
			svcErr:   nil,
			response: sdk.DomainsPage{},
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:  "list domains with invalid page metadata",
			token: validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  10,
				Metadata: sdk.Metadata{
					"key": make(chan int),
				},
			},
			svcReq:   domains.Page{},
			svcRes:   domains.DomainsPage{},
			svcErr:   nil,
			response: sdk.DomainsPage{},
			err:      errors.NewSDKError(errors.New("json: unsupported type: chan int")),
		},
		{
			desc:  "list domains with request that cannot be marshalled",
			token: validToken,
			pageMeta: sdk.PageMetadata{
				Offset: 0,
				Limit:  10,
			},
			svcReq: domains.Page{
				Offset: 0,
				Limit:  10,
				Order:  internalapi.DefOrder,
				Dir:    internalapi.DefDir,
			},
			svcRes: domains.DomainsPage{
				Total: 1,
				Domains: []domains.Domain{{
					Name:     authDomain.Name,
					Metadata: domains.Metadata{"key": make(chan int)},
				}},
			},
			svcErr:   nil,
			response: sdk.DomainsPage{},
			err:      errors.NewSDKError(errors.New("unexpected end of JSON input")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: domainID + "_" + validID, UserID: validID, DomainID: domainID}
			}
			authCall := authn.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authnErr)
			svcCall := svc.On("ListDomains", mock.Anything, tc.session, tc.svcReq).Return(tc.svcRes, tc.svcErr)
			resp, err := mgsdk.Domains(tc.pageMeta, tc.token)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.response, resp)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "ListDomains", mock.Anything, tc.session, mock.Anything)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestEnableDomain(t *testing.T) {
	ds, svc, authn := setupDomains()
	defer ds.Close()

	sdkConf := sdk.Config{
		DomainsURL:     ds.URL,
		MsgContentType: contentType,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc     string
		token    string
		session  mgauthn.Session
		domainID string
		svcRes   domains.Domain
		svcErr   error
		authnErr error
		err      error
	}{
		{
			desc:     "enable domain successfully",
			token:    validToken,
			domainID: sdkDomain.ID,
			svcRes:   authDomain,
			svcErr:   nil,
			err:      nil,
		},
		{
			desc:     "enable domain with invalid token",
			token:    invalidToken,
			domainID: sdkDomain.ID,
			svcRes:   domains.Domain{},
			authnErr: svcerr.ErrAuthentication,
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "enable domain with empty token",
			token:    "",
			domainID: sdkDomain.ID,
			svcRes:   domains.Domain{},
			svcErr:   nil,
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "enable domain with empty domain id",
			token:    validToken,
			domainID: "",
			svcRes:   domains.Domain{},
			svcErr:   nil,
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrMissingDomainID, http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: tc.domainID + "_" + validID, UserID: validID, DomainID: tc.domainID}
			}
			authCall := authn.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authnErr)
			svcCall := svc.On("EnableDomain", mock.Anything, tc.session, tc.domainID).Return(tc.svcRes, tc.svcErr)
			err := mgsdk.EnableDomain(tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "EnableDomain", mock.Anything, tc.session, tc.domainID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestDisableDomain(t *testing.T) {
	ds, svc, authn := setupDomains()
	defer ds.Close()

	sdkConf := sdk.Config{
		DomainsURL:     ds.URL,
		MsgContentType: contentType,
	}

	mgsdk := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc     string
		token    string
		session  mgauthn.Session
		domainID string
		svcRes   domains.Domain
		svcErr   error
		authnErr error
		err      error
	}{
		{
			desc:     "disable domain successfully",
			token:    validToken,
			domainID: sdkDomain.ID,
			svcRes:   authDomain,
			svcErr:   nil,
			err:      nil,
		},
		{
			desc:     "disable domain with invalid token",
			token:    invalidToken,
			domainID: sdkDomain.ID,
			svcRes:   domains.Domain{},
			authnErr: svcerr.ErrAuthentication,
			err:      errors.NewSDKErrorWithStatus(svcerr.ErrAuthentication, http.StatusUnauthorized),
		},
		{
			desc:     "disable domain with empty token",
			token:    "",
			domainID: sdkDomain.ID,
			svcRes:   domains.Domain{},
			svcErr:   nil,
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrBearerToken, http.StatusUnauthorized),
		},
		{
			desc:     "disable domain with empty domain id",
			token:    validToken,
			domainID: "",
			svcRes:   domains.Domain{},
			svcErr:   nil,
			err:      errors.NewSDKErrorWithStatus(apiutil.ErrMissingDomainID, http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.token == validToken {
				tc.session = mgauthn.Session{DomainUserID: tc.domainID + "_" + validID, UserID: validID, DomainID: tc.domainID}
			}
			authCall := authn.On("Authenticate", mock.Anything, mock.Anything).Return(tc.session, tc.authnErr)
			svcCall := svc.On("DisableDomain", mock.Anything, tc.session, tc.domainID).Return(tc.svcRes, tc.svcErr)
			err := mgsdk.DisableDomain(tc.domainID, tc.token)
			assert.Equal(t, tc.err, err)
			if tc.err == nil {
				ok := svcCall.Parent.AssertCalled(t, "DisableDomain", mock.Anything, tc.session, tc.domainID)
				assert.True(t, ok)
			}
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func generateTestDomain(t *testing.T) (domains.Domain, sdk.Domain) {
	createdAt, err := time.Parse(time.RFC3339, "2024-04-01T00:00:00Z")
	assert.Nil(t, err, fmt.Sprintf("Unexpected error parsing time: %s", err))
	ownerID := testsutil.GenerateUUID(t)
	ad := domains.Domain{
		ID:        testsutil.GenerateUUID(t),
		Name:      "test-domain",
		Metadata:  domains.Metadata(validMetadata),
		Tags:      []string{"tag1", "tag2"},
		Alias:     "test-alias",
		Status:    domains.EnabledStatus,
		CreatedBy: ownerID,
		CreatedAt: createdAt,
		UpdatedBy: ownerID,
		UpdatedAt: createdAt,
	}

	sd := sdk.Domain{
		ID:        ad.ID,
		Name:      ad.Name,
		Metadata:  validMetadata,
		Tags:      ad.Tags,
		Alias:     ad.Alias,
		Status:    ad.Status.String(),
		CreatedBy: ad.CreatedBy,
		CreatedAt: ad.CreatedAt,
		UpdatedBy: ad.UpdatedBy,
		UpdatedAt: ad.UpdatedAt,
	}
	return ad, sd
}

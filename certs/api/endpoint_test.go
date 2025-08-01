// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/certs"
	"github.com/absmach/supermq/certs/api"
	"github.com/absmach/supermq/certs/mocks"
	"github.com/absmach/supermq/internal/testsutil"
	smqlog "github.com/absmach/supermq/logger"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	authnmocks "github.com/absmach/supermq/pkg/authn/mocks"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	contentType = "application/json"
	valid       = "valid"
	invalid     = "invalid"
	clientID    = testsutil.GenerateUUID(&testing.T{})
	serial      = testsutil.GenerateUUID(&testing.T{})
	ttl         = "1h"
	cert        = certs.Cert{
		ClientID:     clientID,
		SerialNumber: serial,
		ExpiryTime:   time.Now().Add(time.Hour),
	}
	validID = testsutil.GenerateUUID(&testing.T{})
)

type testRequest struct {
	client      *http.Client
	method      string
	url         string
	contentType string
	token       string
	body        io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, tr.body)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", apiutil.BearerPrefix+tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}

	return tr.client.Do(req)
}

func newCertServer() (*httptest.Server, *mocks.Service, *authnmocks.Authentication) {
	svc := new(mocks.Service)
	logger := smqlog.NewMock()
	idp := uuid.NewMock()
	authn := new(authnmocks.Authentication)
	mux := api.MakeHandler(svc, authn, logger, "", idp)

	return httptest.NewServer(mux), svc, authn
}

func TestIssueCert(t *testing.T) {
	cs, svc, auth := newCertServer()
	defer cs.Close()

	validReqString := `{"client_id": "%s","ttl": "%s"}`
	invalidReqString := `{"client_id": "%s","ttl": %s}`

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		contentType     string
		clientID        string
		ttl             string
		request         string
		status          int
		authenticateErr error
		svcRes          certs.Cert
		svcErr          error
		err             error
	}{
		{
			desc:        "issue cert successfully",
			token:       valid,
			domainID:    valid,
			contentType: contentType,
			clientID:    clientID,
			ttl:         ttl,
			request:     fmt.Sprintf(validReqString, clientID, ttl),
			status:      http.StatusCreated,
			svcRes:      certs.Cert{SerialNumber: serial},
			svcErr:      nil,
			err:         nil,
		},
		{
			desc:        "issue cert with failed service",
			token:       valid,
			domainID:    valid,
			contentType: contentType,
			clientID:    clientID,
			ttl:         ttl,
			request:     fmt.Sprintf(validReqString, clientID, ttl),
			status:      http.StatusUnprocessableEntity,
			svcRes:      certs.Cert{},
			svcErr:      svcerr.ErrCreateEntity,
			err:         svcerr.ErrCreateEntity,
		},
		{
			desc:            "issue with invalid token",
			token:           invalid,
			contentType:     contentType,
			clientID:        clientID,
			ttl:             ttl,
			request:         fmt.Sprintf(validReqString, clientID, ttl),
			status:          http.StatusUnauthorized,
			svcRes:          certs.Cert{},
			authenticateErr: svcerr.ErrAuthentication,
			err:             svcerr.ErrAuthentication,
		},
		{
			desc:        "issue with empty token",
			domainID:    valid,
			contentType: contentType,
			request:     fmt.Sprintf(validReqString, clientID, ttl),
			status:      http.StatusUnauthorized,
			svcRes:      certs.Cert{},
			svcErr:      nil,
			err:         apiutil.ErrBearerToken,
		},
		{
			desc:        "issue with empty domain id",
			token:       valid,
			domainID:    "",
			contentType: contentType,
			request:     fmt.Sprintf(validReqString, clientID, ttl),
			status:      http.StatusBadRequest,
			svcRes:      certs.Cert{},
			svcErr:      nil,
			err:         apiutil.ErrMissingDomainID,
		},
		{
			desc:        "issue with empty client id",
			token:       valid,
			domainID:    valid,
			contentType: contentType,
			request:     fmt.Sprintf(validReqString, "", ttl),
			status:      http.StatusBadRequest,
			svcRes:      certs.Cert{},
			svcErr:      nil,
			err:         apiutil.ErrMissingID,
		},
		{
			desc:        "issue with empty ttl",
			token:       valid,
			domainID:    valid,
			contentType: contentType,
			request:     fmt.Sprintf(validReqString, clientID, ""),
			status:      http.StatusBadRequest,
			svcRes:      certs.Cert{},
			svcErr:      nil,
			err:         apiutil.ErrMissingCertData,
		},
		{
			desc:        "issue with invalid ttl",
			token:       valid,
			domainID:    valid,
			contentType: contentType,
			request:     fmt.Sprintf(validReqString, clientID, invalid),
			status:      http.StatusBadRequest,
			svcRes:      certs.Cert{},
			svcErr:      nil,
			err:         apiutil.ErrInvalidCertData,
		},
		{
			desc:        "issue with invalid content type",
			token:       valid,
			domainID:    valid,
			contentType: "application/xml",
			request:     fmt.Sprintf(validReqString, clientID, ttl),
			status:      http.StatusUnsupportedMediaType,
			svcRes:      certs.Cert{},
			svcErr:      nil,
			err:         apiutil.ErrUnsupportedContentType,
		},
		{
			desc:        "issue with invalid request body",
			token:       valid,
			domainID:    valid,
			contentType: contentType,
			request:     fmt.Sprintf(invalidReqString, clientID, ttl),
			status:      http.StatusInternalServerError,
			svcRes:      certs.Cert{},
			svcErr:      nil,
			err:         apiutil.ErrValidation,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			req := testRequest{
				client:      cs.Client(),
				method:      http.MethodPost,
				url:         fmt.Sprintf("%s/%s/certs", cs.URL, tc.domainID),
				contentType: tc.contentType,
				token:       tc.token,
				body:        strings.NewReader(tc.request),
			}
			if tc.token == valid {
				tc.session = smqauthn.Session{DomainUserID: validID, UserID: validID, DomainID: validID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := svc.On("IssueCert", mock.Anything, tc.domainID, tc.token, tc.clientID, tc.ttl).Return(tc.svcRes, tc.svcErr)
			res, err := req.make()
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			var errRes respBody
			err = json.NewDecoder(res.Body).Decode(&errRes)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error while decoding response body: %s", tc.desc, err))
			if errRes.Err != "" || errRes.Message != "" {
				err = errors.Wrap(errors.New(errRes.Err), errors.New(errRes.Message))
			}
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestViewCert(t *testing.T) {
	cs, svc, auth := newCertServer()
	defer cs.Close()

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		serialID        string
		status          int
		authenticateRes smqauthn.Session
		authenticateErr error
		svcRes          certs.Cert
		svcErr          error
		err             error
	}{
		{
			desc:     "view cert successfully",
			token:    valid,
			domainID: valid,
			serialID: serial,
			status:   http.StatusOK,
			svcRes:   certs.Cert{SerialNumber: serial},
			svcErr:   nil,
			err:      nil,
		},
		{
			desc:            "view with invalid token",
			token:           invalid,
			serialID:        serial,
			status:          http.StatusUnauthorized,
			svcRes:          certs.Cert{},
			authenticateErr: svcerr.ErrAuthentication,
			err:             svcerr.ErrAuthentication,
		},
		{
			desc:     "view with empty token",
			token:    "",
			domainID: valid,
			serialID: serial,
			status:   http.StatusUnauthorized,
			svcRes:   certs.Cert{},
			svcErr:   nil,
			err:      apiutil.ErrBearerToken,
		},
		{
			desc:     "view non-existing cert",
			token:    valid,
			domainID: valid,
			serialID: invalid,
			status:   http.StatusNotFound,
			svcRes:   certs.Cert{},
			svcErr:   svcerr.ErrNotFound,
			err:      svcerr.ErrNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			req := testRequest{
				client: cs.Client(),
				method: http.MethodGet,
				url:    fmt.Sprintf("%s/%s/certs/%s", cs.URL, tc.domainID, tc.serialID),
				token:  tc.token,
			}
			if tc.token == valid {
				tc.session = smqauthn.Session{DomainUserID: validID, UserID: validID, DomainID: validID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := svc.On("ViewCert", mock.Anything, tc.serialID).Return(tc.svcRes, tc.svcErr)
			res, err := req.make()
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			var errRes respBody
			err = json.NewDecoder(res.Body).Decode(&errRes)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error while decoding response body: %s", tc.desc, err))
			if errRes.Err != "" || errRes.Message != "" {
				err = errors.Wrap(errors.New(errRes.Err), errors.New(errRes.Message))
			}
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestRevokeAllCerts(t *testing.T) {
	cs, svc, auth := newCertServer()
	defer cs.Close()

	cases := []struct {
		desc            string
		domainID        string
		token           string
		session         smqauthn.Session
		serialID        string
		status          int
		authenticateErr error
		svcRes          certs.Revoke
		svcErr          error
		err             error
	}{
		{
			desc:     "revoke cert successfully",
			token:    valid,
			domainID: valid,
			serialID: serial,
			status:   http.StatusOK,
			svcRes:   certs.Revoke{RevocationTime: time.Now()},
			svcErr:   nil,
			err:      nil,
		},
		{
			desc:            "revoke with invalid token",
			token:           invalid,
			serialID:        serial,
			status:          http.StatusUnauthorized,
			svcRes:          certs.Revoke{},
			authenticateErr: svcerr.ErrAuthentication,
			err:             svcerr.ErrAuthentication,
		},
		{
			desc:     "revoke with empty domain id",
			token:    valid,
			domainID: "",
			serialID: serial,
			status:   http.StatusBadRequest,
			svcErr:   nil,
			err:      apiutil.ErrMissingDomainID,
		},
		{
			desc:     "revoke with empty token",
			token:    "",
			domainID: valid,
			serialID: serial,
			status:   http.StatusUnauthorized,
			svcErr:   nil,
			err:      apiutil.ErrBearerToken,
		},
		{
			desc:     "revoke non-existing cert",
			token:    valid,
			domainID: valid,
			serialID: invalid,
			status:   http.StatusNotFound,
			svcRes:   certs.Revoke{},
			svcErr:   svcerr.ErrNotFound,
			err:      svcerr.ErrNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			req := testRequest{
				client: cs.Client(),
				method: http.MethodPost,
				url:    fmt.Sprintf("%s/%s/certs/%s/revoke-all", cs.URL, tc.domainID, tc.serialID),
				token:  tc.token,
			}
			if tc.token == valid {
				tc.session = smqauthn.Session{DomainUserID: validID, UserID: validID, DomainID: validID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := svc.On("RevokeCert", mock.Anything, tc.domainID, tc.token, tc.serialID).Return(tc.svcRes, tc.svcErr)
			res, err := req.make()
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			var errRes respBody
			err = json.NewDecoder(res.Body).Decode(&errRes)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error while decoding response body: %s", tc.desc, err))
			if errRes.Err != "" || errRes.Message != "" {
				err = errors.Wrap(errors.New(errRes.Err), errors.New(errRes.Message))
			}
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n ", tc.desc, tc.err, err))
			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestRevokeBySerial(t *testing.T) {
	cs, svc, auth := newCertServer()
	defer cs.Close()

	cases := []struct {
		desc            string
		token           string
		domainID        string
		session         smqauthn.Session
		serialID        string
		status          int
		authenticateErr error
		svcRes          certs.Revoke
		svcErr          error
		err             error
	}{
		{
			desc:     "revoke cert by serial successfully",
			token:    valid,
			domainID: valid,
			serialID: serial,
			status:   http.StatusOK,
			svcRes:   certs.Revoke{RevocationTime: time.Now()},
			svcErr:   nil,
			err:      nil,
		},
		{
			desc:            "revoke by serial with invalid token",
			token:           invalid,
			domainID:        valid,
			serialID:        serial,
			status:          http.StatusUnauthorized,
			svcRes:          certs.Revoke{},
			authenticateErr: svcerr.ErrAuthentication,
			err:             svcerr.ErrAuthentication,
		},
		{
			desc:     "revoke by serial with empty domain id",
			token:    valid,
			domainID: "",
			serialID: serial,
			status:   http.StatusBadRequest,
			svcErr:   nil,
			err:      apiutil.ErrMissingDomainID,
		},
		{
			desc:     "revoke by serial with empty token",
			token:    "",
			serialID: serial,
			domainID: valid,
			status:   http.StatusUnauthorized,
			svcErr:   nil,
			err:      apiutil.ErrBearerToken,
		},
		{
			desc:     "revoke by serial with empty serial ID",
			token:    valid,
			domainID: valid,
			serialID: "",
			status:   http.StatusBadRequest,
			svcErr:   nil,
			err:      apiutil.ErrMissingID,
		},
		{
			desc:     "revoke non-existing cert by serial",
			token:    valid,
			domainID: valid,
			serialID: invalid,
			status:   http.StatusNotFound,
			svcRes:   certs.Revoke{},
			svcErr:   svcerr.ErrNotFound,
			err:      svcerr.ErrNotFound,
		},
		{
			desc:     "revoke by serial with service error",
			token:    valid,
			domainID: valid,
			serialID: serial,
			status:   http.StatusUnprocessableEntity,
			svcRes:   certs.Revoke{},
			svcErr:   svcerr.ErrRemoveEntity,
			err:      svcerr.ErrRemoveEntity,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			req := testRequest{
				client: cs.Client(),
				method: http.MethodPost,
				url:    fmt.Sprintf("%s/%s/certs/%s/revoke", cs.URL, tc.domainID, tc.serialID),
				token:  tc.token,
			}
			if tc.token == valid {
				tc.session = smqauthn.Session{DomainUserID: validID, UserID: validID, DomainID: validID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := svc.On("RevokeBySerial", mock.Anything, tc.serialID).Return(tc.svcRes, tc.svcErr)
			res, err := req.make()
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			var errRes respBody
			err = json.NewDecoder(res.Body).Decode(&errRes)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error while decoding response body: %s", tc.desc, err))
			if errRes.Err != "" || errRes.Message != "" {
				err = errors.Wrap(errors.New(errRes.Err), errors.New(errRes.Message))
			}
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

func TestListSerials(t *testing.T) {
	cs, svc, auth := newCertServer()
	defer cs.Close()
	revoked := "false"

	cases := []struct {
		desc            string
		token           string
		domainID        string
		session         smqauthn.Session
		clientID        string
		revoked         string
		offset          uint64
		limit           uint64
		query           string
		status          int
		authenticateErr error
		svcRes          certs.CertPage
		svcErr          error
		err             error
	}{
		{
			desc:     "list certs successfully with default limit",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			offset:   0,
			limit:    10,
			query:    "",
			status:   http.StatusOK,
			svcRes: certs.CertPage{
				Total:        1,
				Offset:       0,
				Limit:        10,
				Certificates: []certs.Cert{cert},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "list certs successfully with default revoke",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			offset:   0,
			limit:    10,
			query:    "",
			status:   http.StatusOK,
			svcRes: certs.CertPage{
				Total:        1,
				Offset:       0,
				Limit:        10,
				Certificates: []certs.Cert{cert},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "list certs successfully with all certs",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  "all",
			offset:   0,
			limit:    10,
			query:    "?revoked=all",
			status:   http.StatusOK,
			svcRes: certs.CertPage{
				Total:        1,
				Offset:       0,
				Limit:        10,
				Certificates: []certs.Cert{cert},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "list certs successfully with limit",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			offset:   0,
			limit:    5,
			query:    "?limit=5",
			status:   http.StatusOK,
			svcRes: certs.CertPage{
				Total:        1,
				Offset:       0,
				Limit:        5,
				Certificates: []certs.Cert{cert},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "list certs successfully with offset",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			offset:   1,
			limit:    10,
			query:    "?offset=1",
			status:   http.StatusOK,
			svcRes: certs.CertPage{
				Total:        1,
				Offset:       1,
				Limit:        10,
				Certificates: []certs.Cert{},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:     "list certs successfully with offset and limit",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			offset:   1,
			limit:    5,
			query:    "?offset=1&limit=5",
			status:   http.StatusOK,
			svcRes: certs.CertPage{
				Total:        1,
				Offset:       1,
				Limit:        5,
				Certificates: []certs.Cert{},
			},
			svcErr: nil,
			err:    nil,
		},
		{
			desc:            "list with invalid token",
			domainID:        valid,
			token:           invalid,
			clientID:        clientID,
			revoked:         revoked,
			offset:          0,
			limit:           10,
			query:           "",
			status:          http.StatusUnauthorized,
			svcRes:          certs.CertPage{},
			authenticateErr: svcerr.ErrAuthentication,
			err:             svcerr.ErrAuthentication,
		},
		{
			desc:     "list with empty token",
			domainID: valid,
			token:    "",
			clientID: clientID,
			revoked:  revoked,
			offset:   0,
			limit:    10,
			query:    "",
			status:   http.StatusUnauthorized,
			svcRes:   certs.CertPage{},
			svcErr:   nil,
			err:      apiutil.ErrBearerToken,
		},
		{
			desc:     "list with limit exceeding max limit",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			query:    "?limit=1000",
			status:   http.StatusBadRequest,
			svcRes:   certs.CertPage{},
			svcErr:   nil,
			err:      apiutil.ErrLimitSize,
		},
		{
			desc:     "list with invalid offset",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			query:    "?offset=invalid",
			status:   http.StatusBadRequest,
			svcRes:   certs.CertPage{},
			svcErr:   nil,
			err:      apiutil.ErrValidation,
		},
		{
			desc:     "list with invalid limit",
			domainID: valid,
			token:    valid,
			clientID: clientID,
			revoked:  revoked,
			query:    "?limit=invalid",
			status:   http.StatusBadRequest,
			svcRes:   certs.CertPage{},
			svcErr:   nil,
			err:      apiutil.ErrValidation,
		},
		{
			desc:     "list with invalid client id",
			domainID: valid,
			token:    valid,
			clientID: invalid,
			revoked:  revoked,
			offset:   0,
			limit:    10,
			query:    "",
			status:   http.StatusNotFound,
			svcRes:   certs.CertPage{},
			svcErr:   svcerr.ErrNotFound,
			err:      svcerr.ErrNotFound,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			req := testRequest{
				client: cs.Client(),
				method: http.MethodGet,
				url:    fmt.Sprintf("%s/%s/serials/%s", cs.URL, tc.domainID, tc.clientID) + tc.query,
				token:  tc.token,
			}
			if tc.token == valid {
				tc.session = smqauthn.Session{DomainUserID: validID, UserID: validID, DomainID: validID}
			}
			authCall := auth.On("Authenticate", mock.Anything, tc.token).Return(tc.session, tc.authenticateErr)
			svcCall := svc.On("ListSerials", mock.Anything, tc.clientID, certs.PageMetadata{Offset: tc.offset, Limit: tc.limit, Revoked: "all"}).Return(tc.svcRes, tc.svcErr)
			res, err := req.make()
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
			var errRes respBody
			err = json.NewDecoder(res.Body).Decode(&errRes)
			assert.Nil(t, err, fmt.Sprintf("%s: unexpected error while decoding response body: %s", tc.desc, err))
			if errRes.Err != "" || errRes.Message != "" {
				err = errors.Wrap(errors.New(errRes.Err), errors.New(errRes.Message))
			}
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n ", tc.desc, tc.err, err))
			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
			svcCall.Unset()
			authCall.Unset()
		})
	}
}

type respBody struct {
	Err     string `json:"error"`
	Message string `json:"message"`
}

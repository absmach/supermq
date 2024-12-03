// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/absmach/supermq/internal/testsutil"
	"github.com/absmach/supermq/invitations"
	"github.com/absmach/supermq/invitations/api"
	"github.com/absmach/supermq/invitations/mocks"
	smqlog "github.com/absmach/supermq/logger"
	"github.com/absmach/supermq/pkg/apiutil"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	authnmocks "github.com/absmach/supermq/pkg/authn/mocks"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	validToken      = "valid"
	validContenType = "application/json"
	validID         = testsutil.GenerateUUID(&testing.T{})
	domainID        = testsutil.GenerateUUID(&testing.T{})
)

type testRequest struct {
	client      *http.Client
	method      string
	url         string
	token       string
	contentType string
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

func newIvitationsServer() (*httptest.Server, *mocks.Service, *authnmocks.Authentication) {
	svc := new(mocks.Service)
	logger := smqlog.NewMock()
	authn := new(authnmocks.Authentication)
	mux := api.MakeHandler(svc, logger, authn, "test")
	return httptest.NewServer(mux), svc, authn
}

func TestSendInvitation(t *testing.T) {
	is, svc, authn := newIvitationsServer()

	cases := []struct {
		desc        string
		token       string
		data        string
		contentType string
		status      int
		authnRes    smqauthn.Session
		authnErr    error
		svcErr      error
	}{
		{
			desc:        "valid request",
			token:       validToken,
			data:        fmt.Sprintf(`{"user_id": "%s","domain_id": "%s", "relation": "%s"}`, validID, domainID, "domain"),
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			status:      http.StatusCreated,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "invalid token",
			token:       "",
			data:        fmt.Sprintf(`{"user_id": "%s","domain_id": "%s",  "relation": "%s"}`, validID, validID, "domain"),
			status:      http.StatusUnauthorized,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "empty domain_id",
			token:       validToken,
			data:        fmt.Sprintf(`{"user_id": "%s","domain_id": "%s",  "relation": "%s"}`, validID, "", "domain"),
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "invalid content type",
			token:       validToken,
			data:        fmt.Sprintf(`{"user_id": "%s","domain_id": "%s",  "relation": "%s"}`, validID, validID, "domain"),
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			status:      http.StatusUnsupportedMediaType,
			contentType: "text/plain",
			svcErr:      nil,
		},
		{
			desc:        "invalid data",
			token:       validToken,
			data:        `data`,
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with service error",
			token:       validToken,
			data:        fmt.Sprintf(`{"user_id": "%s", "domain_id": "%s", "relation": "%s"}`, validID, domainID, "domain"),
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			status:      http.StatusForbidden,
			contentType: validContenType,
			svcErr:      svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authnCall := authn.On("Authenticate", mock.Anything, tc.token).Return(tc.authnRes, tc.authnErr)
			repoCall := svc.On("SendInvitation", mock.Anything, tc.authnRes, mock.Anything).Return(tc.svcErr)
			req := testRequest{
				client:      is.Client(),
				method:      http.MethodPost,
				url:         is.URL + "/invitations",
				token:       tc.token,
				contentType: tc.contentType,
				body:        strings.NewReader(tc.data),
			}

			res, err := req.make()
			assert.Nil(t, err, tc.desc)
			assert.Equal(t, tc.status, res.StatusCode, tc.desc)
			repoCall.Unset()
			authnCall.Unset()
		})
	}
}

func TestListInvitation(t *testing.T) {
	is, svc, authn := newIvitationsServer()

	cases := []struct {
		desc        string
		token       string
		query       string
		contentType string
		status      int
		svcErr      error
		authnRes    smqauthn.Session
		authnErr    error
	}{
		{
			desc:        "valid request",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "invalid token",
			token:       "",
			status:      http.StatusUnauthorized,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with offset",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       "offset=1",
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with invalid offset",
			token:       validToken,
			query:       "offset=invalid",
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with limit",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       "limit=1",
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with invalid limit",
			token:       validToken,
			query:       "limit=invalid",
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with user_id",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       fmt.Sprintf("user_id=%s", validID),
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with duplicate user_id",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       "user_id=1&user_id=2",
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with invited_by",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       fmt.Sprintf("invited_by=%s", validID),
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with duplicate invited_by",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       "invited_by=1&invited_by=2",
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with relation",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       fmt.Sprintf("relation=%s", "relation"),
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with duplicate relation",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       "relation=1&relation=2",
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with state",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			query:       "state=pending",
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with invalid state",
			token:       validToken,
			query:       "state=invalid",
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with duplicate state",
			token:       validToken,
			query:       "state=all&state=all",
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with service error",
			authnRes:    smqauthn.Session{UserID: validID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			status:      http.StatusForbidden,
			contentType: validContenType,
			svcErr:      svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authnCall := authn.On("Authenticate", mock.Anything, tc.token).Return(tc.authnRes, tc.authnErr)
			repoCall := svc.On("ListInvitations", mock.Anything, tc.authnRes, mock.Anything).Return(invitations.InvitationPage{}, tc.svcErr)
			req := testRequest{
				client:      is.Client(),
				method:      http.MethodGet,
				url:         is.URL + "/invitations?" + tc.query,
				token:       tc.token,
				contentType: tc.contentType,
			}
			res, err := req.make()
			assert.Nil(t, err, tc.desc)
			assert.Equal(t, tc.status, res.StatusCode, tc.desc)
			repoCall.Unset()
			authnCall.Unset()
		})
	}
}

func TestViewInvitation(t *testing.T) {
	is, svc, authn := newIvitationsServer()

	cases := []struct {
		desc        string
		token       string
		domainID    string
		userID      string
		contentType string
		status      int
		svcErr      error
		authnRes    smqauthn.Session
		authnErr    error
	}{
		{
			desc:        "valid request",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			userID:      validID,
			domainID:    domainID,
			status:      http.StatusOK,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "invalid token",
			token:       "",
			userID:      validID,
			domainID:    domainID,
			status:      http.StatusUnauthorized,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with service error",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			userID:      validID,
			domainID:    domainID,
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      svcerr.ErrViewEntity,
		},
		{
			desc:        "with empty user_id",
			token:       validToken,
			userID:      "",
			domainID:    domainID,
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with empty domain",
			token:       validToken,
			userID:      validID,
			domainID:    "",
			status:      http.StatusNotFound,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with empty user_id and domain_id",
			token:       validToken,
			userID:      "",
			domainID:    "",
			status:      http.StatusNotFound,
			contentType: validContenType,
			svcErr:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authnCall := authn.On("Authenticate", mock.Anything, tc.token).Return(tc.authnRes, tc.authnErr)
			repoCall := svc.On("ViewInvitation", mock.Anything, tc.authnRes, tc.userID, tc.domainID).Return(invitations.Invitation{}, tc.svcErr)
			req := testRequest{
				client:      is.Client(),
				method:      http.MethodGet,
				url:         is.URL + "/invitations/" + tc.userID + "/" + tc.domainID,
				token:       tc.token,
				contentType: tc.contentType,
			}

			res, err := req.make()
			assert.Nil(t, err, tc.desc)
			assert.Equal(t, tc.status, res.StatusCode, tc.desc)
			repoCall.Unset()
			authnCall.Unset()
		})
	}
}

func TestDeleteInvitation(t *testing.T) {
	is, svc, authn := newIvitationsServer()
	_ = authn

	cases := []struct {
		desc        string
		token       string
		domainID    string
		userID      string
		contentType string
		status      int
		svcErr      error
		authnRes    smqauthn.Session
		authnErr    error
	}{
		{
			desc:        "valid request",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			userID:      validID,
			domainID:    domainID,
			status:      http.StatusNoContent,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "invalid token",
			token:       "",
			userID:      validID,
			domainID:    domainID,
			status:      http.StatusUnauthorized,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with service error",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			userID:      validID,
			domainID:    domainID,
			status:      http.StatusForbidden,
			contentType: validContenType,
			svcErr:      svcerr.ErrAuthorization,
		},
		{
			desc:        "with empty user_id",
			token:       validToken,
			userID:      "",
			domainID:    domainID,
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with empty domain_id",
			token:       validToken,
			userID:      validID,
			domainID:    "",
			status:      http.StatusNotFound,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with empty user_id and domain_id",
			token:       validToken,
			userID:      "",
			domainID:    "",
			status:      http.StatusNotFound,
			contentType: validContenType,
			svcErr:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authnCall := authn.On("Authenticate", mock.Anything, tc.token).Return(tc.authnRes, tc.authnErr)
			repoCall := svc.On("DeleteInvitation", mock.Anything, tc.authnRes, tc.userID, tc.domainID).Return(tc.svcErr)
			req := testRequest{
				client:      is.Client(),
				method:      http.MethodDelete,
				url:         is.URL + "/invitations/" + tc.userID + "/" + tc.domainID,
				token:       tc.token,
				contentType: tc.contentType,
			}

			res, err := req.make()
			assert.Nil(t, err, tc.desc)
			assert.Equal(t, tc.status, res.StatusCode, tc.desc)
			repoCall.Unset()
			authnCall.Unset()
		})
	}
}

func TestAcceptInvitation(t *testing.T) {
	is, svc, authn := newIvitationsServer()
	_ = authn
	cases := []struct {
		desc        string
		token       string
		data        string
		contentType string
		status      int
		svcErr      error
		authnRes    smqauthn.Session
		authnErr    error
	}{
		{
			desc:        "valid request",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, validID),
			token:       validToken,
			status:      http.StatusNoContent,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "invalid token",
			token:       "",
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, validID),
			status:      http.StatusUnauthorized,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "with service error",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, validID),
			status:      http.StatusForbidden,
			contentType: validContenType,
			svcErr:      svcerr.ErrAuthorization,
		},
		{
			desc:        "invalid content type",
			token:       validToken,
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, validID),
			status:      http.StatusUnsupportedMediaType,
			contentType: "text/plain",
			svcErr:      nil,
		},
		{
			desc:        "invalid data",
			token:       validToken,
			data:        `data`,
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authnCall := authn.On("Authenticate", mock.Anything, tc.token).Return(tc.authnRes, tc.authnErr)
			repoCall := svc.On("AcceptInvitation", mock.Anything, tc.authnRes, mock.Anything).Return(tc.svcErr)
			req := testRequest{
				client:      is.Client(),
				method:      http.MethodPost,
				url:         is.URL + "/invitations/accept",
				token:       tc.token,
				contentType: tc.contentType,
				body:        strings.NewReader(tc.data),
			}

			res, err := req.make()
			assert.Nil(t, err, tc.desc)
			assert.Equal(t, tc.status, res.StatusCode, tc.desc)
			repoCall.Unset()
			authnCall.Unset()
		})
	}
}

func TestRejectInvitation(t *testing.T) {
	is, svc, authn := newIvitationsServer()
	_ = authn

	cases := []struct {
		desc        string
		token       string
		data        string
		contentType string
		status      int
		svcErr      error
		authnRes    smqauthn.Session
		authnErr    error
	}{
		{
			desc:        "valid request",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, validID),
			status:      http.StatusNoContent,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "invalid token",
			token:       "",
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, validID),
			status:      http.StatusUnauthorized,
			contentType: validContenType,
			svcErr:      nil,
		},
		{
			desc:        "unauthorized error",
			authnRes:    smqauthn.Session{UserID: validID, DomainID: domainID, DomainUserID: domainID + "_" + validID},
			token:       validToken,
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, "invalid"),
			status:      http.StatusForbidden,
			contentType: validContenType,
			svcErr:      svcerr.ErrAuthorization,
		},
		{
			desc:        "invalid content type",
			token:       validToken,
			data:        fmt.Sprintf(`{"domain_id": "%s"}`, validID),
			status:      http.StatusUnsupportedMediaType,
			contentType: "text/plain",
			svcErr:      nil,
		},
		{
			desc:        "invalid data",
			token:       validToken,
			data:        `data`,
			status:      http.StatusBadRequest,
			contentType: validContenType,
			svcErr:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authnCall := authn.On("Authenticate", mock.Anything, tc.token).Return(tc.authnRes, tc.authnErr)
			repoCall := svc.On("RejectInvitation", mock.Anything, tc.authnRes, mock.Anything).Return(tc.svcErr)
			req := testRequest{
				client:      is.Client(),
				method:      http.MethodPost,
				url:         is.URL + "/invitations/reject",
				token:       tc.token,
				contentType: tc.contentType,
				body:        strings.NewReader(tc.data),
			}

			res, err := req.make()
			assert.Nil(t, err, tc.desc)
			assert.Equal(t, tc.status, res.StatusCode, tc.desc)
			repoCall.Unset()
			authnCall.Unset()
		})
	}
}

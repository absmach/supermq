// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/activitylog/api"
	"github.com/absmach/magistrala/activitylog/mocks"
	"github.com/absmach/magistrala/internal/apiutil"
	mglog "github.com/absmach/magistrala/logger"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var validToken = "valid"

type testRequest struct {
	client *http.Client
	method string
	url    string
	token  string
	body   io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, tr.body)
	if err != nil {
		return nil, err
	}

	if tr.token != "" {
		req.Header.Set("Authorization", apiutil.BearerPrefix+tr.token)
	}

	return tr.client.Do(req)
}

func newActivityLogServer() (*httptest.Server, *mocks.Service) {
	svc := new(mocks.Service)

	logger := mglog.NewMock()
	mux := api.MakeHandler(svc, logger, "activity-log", "test")
	return httptest.NewServer(mux), svc
}

func TestListActivitiesEndpoint(t *testing.T) {
	es, svc := newActivityLogServer()

	cases := []struct {
		desc        string
		token       string
		url         string
		contentType string
		status      int
		svcErr      error
	}{
		{
			desc:   "successful",
			token:  validToken,
			url:    "",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "empty token",
			token:  "",
			url:    "",
			status: http.StatusUnauthorized,
			svcErr: nil,
		},
		{
			desc:   "with service error",
			token:  validToken,
			url:    "",
			status: http.StatusForbidden,
			svcErr: svcerr.ErrAuthorization,
		},
		{
			desc:   "with offset",
			token:  validToken,
			url:    "?offset=10",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid offset",
			token:  validToken,
			url:    "?offset=ten",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with limit",
			token:  validToken,
			url:    "?limit=10",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid limit",
			token:  validToken,
			url:    "?limit=ten",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with operation",
			token:  validToken,
			url:    "?operation=user.create",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with malformed operation",
			token:  validToken,
			url:    "?operation=user.create&operation=user.update",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with from",
			token:  validToken,
			url:    fmt.Sprintf("?from=%d", time.Now().Unix()),
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid from",
			token:  validToken,
			url:    "?from=ten",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with invalid from as UnixNano",
			token:  validToken,
			url:    fmt.Sprintf("?from=%d", time.Now().UnixNano()),
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with to",
			token:  validToken,
			url:    fmt.Sprintf("?to=%d", time.Now().Unix()),
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid to",
			token:  validToken,
			url:    "?to=ten",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with invalid to as UnixNano",
			token:  validToken,
			url:    fmt.Sprintf("?to=%d", time.Now().UnixNano()),
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with attributes",
			token:  validToken,
			url:    fmt.Sprintf("?with_attributes=%s", strconv.FormatBool(true)),
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid attributes",
			token:  validToken,
			url:    "?with_attributes=ten",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with metadata",
			token:  validToken,
			url:    fmt.Sprintf("?with_metadata=%s", strconv.FormatBool(true)),
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid metadata",
			token:  validToken,
			url:    "?with_metadata=ten",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with asc direction",
			token:  validToken,
			url:    "?dir=asc",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with desc direction",
			token:  validToken,
			url:    "?dir=desc",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid direction",
			token:  validToken,
			url:    "?dir=ten",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with malformed direction",
			token:  validToken,
			url:    "?dir=invalid&dir=invalid2",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with id",
			token:  validToken,
			url:    "?id=123&entity_type=user",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with malformed id",
			token:  validToken,
			url:    "?id=123&id=456",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with entity type",
			token:  validToken,
			url:    "?entity_type=user&id=123",
			status: http.StatusOK,
			svcErr: nil,
		},
		{
			desc:   "with invalid entity type",
			token:  validToken,
			url:    "?entity_type=invalid",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with malformed entity type",
			token:  validToken,
			url:    "?entity_type=user&entity_type=thing",
			status: http.StatusBadRequest,
			svcErr: nil,
		},
		{
			desc:   "with all query params",
			token:  validToken,
			url:    "?offset=10&limit=10&operation=user.create&from=0&to=10&with_attributes=true&with_metadata=true&dir=asc&id=123&entity_type=user",
			status: http.StatusOK,
			svcErr: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			repoCall := svc.On("RetrieveAll", mock.Anything, c.token, mock.Anything).Return(activitylog.ActivitiesPage{}, c.svcErr)
			req := testRequest{
				client: es.Client(),
				method: http.MethodGet,
				url:    es.URL + "/activities" + c.url,
				token:  c.token,
			}

			resp, err := req.make()
			assert.Nil(t, err, c.desc)
			defer resp.Body.Close()
			assert.Equal(t, c.status, resp.StatusCode, c.desc)
			repoCall.Unset()
		})
	}
}

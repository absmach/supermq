// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package policies_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mainflux/mainflux/auth"
	httpapi "github.com/mainflux/mainflux/auth/api/http"
	"github.com/mainflux/mainflux/auth/jwt"
	"github.com/mainflux/mainflux/auth/mocks"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/opentracing/opentracing-go/mocktracer"
)

const (
	secret       = "secret"
	contentType  = "application/json"
	id           = "123e4567-e89b-12d3-a456-000000000001"
	email        = "user@example.com"
	unauthzID    = "123e4567-e89b-12d3-a456-000000000002"
	unauthzEmail = "unauthz@example.com"
)

type issueRequest struct {
	Duration time.Duration `json:"duration,omitempty"`
	Type     uint32        `json:"type,omitempty"`
}

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
		req.Header.Set("Authorization", tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}

	req.Header.Set("Referer", "http://localhost")
	return tr.client.Do(req)
}

func newService() auth.Service {
	repo := mocks.NewKeyRepository()
	groupRepo := mocks.NewGroupRepository()
	idProvider := uuid.NewMock()
	t := jwt.New(secret)

	mockAuthzDB := map[string][]mocks.MockSubjectSet{}
	mockAuthzDB[id] = append(mockAuthzDB[id], mocks.MockSubjectSet{Object: "authorities", Relation: "member"})
	mockAuthzDB[unauthzID] = append(mockAuthzDB[unauthzID], mocks.MockSubjectSet{Object: "users", Relation: "member"})
	ketoMock := mocks.NewKetoMock(mockAuthzDB)

	return auth.New(repo, groupRepo, idProvider, t, ketoMock)
}

func newServer(svc auth.Service) *httptest.Server {
	mux := httpapi.MakeHandler(svc, mocktracer.New())
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

type addPolicyRequest struct {
	SubjectIDs []string `json:"subjects"`
	Policies   []string `json:"policies"`
	Object     string   `json:"object"`
}

func TestAddPolicies(t *testing.T) {
	svc := newService()
	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), IssuerID: id, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	_, userLoginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), IssuerID: unauthzID, Subject: unauthzEmail})
	assert.Nil(t, err, fmt.Sprintf("Issuing unauthorized user's key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	valid := addPolicyRequest{Object: "obj", Policies: []string{"read"}, SubjectIDs: []string{"user1", "user2"}}
	invalidObject := addPolicyRequest{Object: "", Policies: []string{"read"}, SubjectIDs: []string{"user1", "user2"}}
	invalidPolicies := addPolicyRequest{Object: "obj", Policies: []string{"read", "invalid"}, SubjectIDs: []string{"user1", "user2"}}
	invalidSubjects := addPolicyRequest{Object: "obj", Policies: []string{"read", "access"}, SubjectIDs: []string{"", "user2"}}

	cases := []struct {
		desc   string
		token  string
		ct     string
		status int
		req    string
	}{
		{
			desc:   "Add policies with authorized access",
			token:  loginSecret,
			ct:     contentType,
			status: http.StatusOK,
			req:    toJSON(valid),
		},
		{
			desc:   "Add policies with unauthorized access",
			token:  userLoginSecret,
			ct:     contentType,
			status: http.StatusForbidden,
			req:    toJSON(valid),
		},
		{
			desc:   "Add policies with invalid token",
			token:  "invalid",
			ct:     contentType,
			status: http.StatusForbidden,
			req:    toJSON(valid),
		},
		{
			desc:   "Add policies with empty token",
			token:  "",
			ct:     contentType,
			status: http.StatusForbidden,
			req:    toJSON(valid),
		},
		{
			desc:   "Add policies with invalid content type",
			token:  loginSecret,
			ct:     "text/html",
			status: http.StatusUnsupportedMediaType,
			req:    toJSON(valid),
		},
		{
			desc:   "Add policies with empty content type",
			token:  loginSecret,
			ct:     "",
			status: http.StatusUnsupportedMediaType,
			req:    toJSON(valid),
		},
		{
			desc:   "Add policies with invalid object field in request body",
			token:  loginSecret,
			ct:     contentType,
			status: http.StatusBadRequest,
			req:    toJSON(invalidObject),
		},
		{
			desc:   "Add policies with invalid policies field in request body",
			token:  loginSecret,
			ct:     contentType,
			status: http.StatusBadRequest,
			req:    toJSON(invalidPolicies),
		},
		{
			desc:   "Add policies with invalid subjects field in request body",
			token:  loginSecret,
			ct:     contentType,
			status: http.StatusBadRequest,
			req:    toJSON(invalidSubjects),
		},
		{
			desc:   "Add policies with empty request body",
			token:  loginSecret,
			ct:     contentType,
			status: http.StatusBadRequest,
			req:    "",
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/policy", ts.URL),
			contentType: tc.ct,
			token:       tc.token,
			body:        strings.NewReader(tc.req),
		}

		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

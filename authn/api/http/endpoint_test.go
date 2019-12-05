// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authn "github.com/mainflux/mainflux/authn"
	httpapi "github.com/mainflux/mainflux/authn/api/http"
	"github.com/mainflux/mainflux/authn/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

const (
	secret       = "secret"
	contentType  = "application/json"
	invalidEmail = "userexample.com"
	wrongID      = "123e4567-e89b-12d3-a456-000000000042"
	id           = "123e4567-e89b-12d3-a456-000000000001"
	email        = "user@example.com"
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
		req.Header.Set("Authorization", tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}

	req.Header.Set("Referer", "http://localhost")
	return tr.client.Do(req)
}

func newService() authn.Service {
	repo := mocks.NewKeyRepository()
	idp := mocks.NewIdentityProvider()
	return authn.New(repo, idp, secret)
}

func newServer(svc authn.Service) *httptest.Server {
	mux := httpapi.MakeHandler(svc, mocktracer.New())
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestIssue(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	lk := authn.Key{Type: authn.LoginKey}
	uk := authn.Key{Type: authn.UserKey}
	rk := authn.Key{Type: authn.ResetKey}

	cases := []struct {
		desc   string
		req    string
		ct     string
		token  string
		status int
	}{
		{"issue login key", toJSON(lk), contentType, "", http.StatusCreated},
		{"issue user key", toJSON(uk), contentType, loginKey.Secret, http.StatusCreated},
		{"issue reset key", toJSON(rk), contentType, loginKey.Secret, http.StatusBadRequest},
		{"issue login key wrong content type", toJSON(lk), "", loginKey.Secret, http.StatusUnsupportedMediaType},
		{"issue key wrong content type", toJSON(rk), "", loginKey.Secret, http.StatusUnsupportedMediaType},
		{"issue key unauthorized", toJSON(uk), contentType, "wrong", http.StatusForbidden},
		{"issue reset key with empty token", toJSON(rk), contentType, "", http.StatusBadRequest},
		{"issue key with invalid request", "{", contentType, "", http.StatusBadRequest},
		{"issue key with invalid JSON", "{invalid}", contentType, "", http.StatusBadRequest},
		{"issue key with invalid JSON content", `{"Type":{"key":"value"}}`, contentType, "", http.StatusBadRequest},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      client,
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/keys", ts.URL),
			contentType: tc.ct,
			token:       tc.token,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestRetrieve(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := authn.Key{Type: authn.UserKey, IssuedAt: time.Now()}

	k, err := svc.Issue(context.Background(), loginKey.Secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	cases := []struct {
		desc   string
		id     string
		token  string
		status int
	}{
		{"retrieve an existing key", k.ID, loginKey.Secret, http.StatusOK},
		{"retrieve a non-existing key", "non-existing", loginKey.Secret, http.StatusNotFound},
		{"retrieve a key unauthorized", k.ID, "wrong", http.StatusForbidden},
	}

	for _, tc := range cases {
		req := testRequest{
			client: client,
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/keys/%s", ts.URL, tc.id),
			token:  tc.token,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestRevoke(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), email, authn.Key{Type: authn.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := authn.Key{Type: authn.UserKey, IssuedAt: time.Now()}

	k, err := svc.Issue(context.Background(), loginKey.Secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()

	cases := []struct {
		desc   string
		id     string
		token  string
		status int
	}{
		{"revoke an existing key", k.ID, loginKey.Secret, http.StatusNoContent},
		{"revoke a non-existing key", "non-existing", loginKey.Secret, http.StatusNoContent},
		{"revoke a key unauthorized", k.ID, "wrong", http.StatusForbidden},
	}

	for _, tc := range cases {
		req := testRequest{
			client: client,
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/keys/%s", ts.URL, tc.id),
			token:  tc.token,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

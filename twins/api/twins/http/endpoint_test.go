//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/twins"
	httpapi "github.com/mainflux/mainflux/twins/api/twins/http"
	"github.com/mainflux/mainflux/twins/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
)

const (
	contentType = "application/json"
	email       = "user@example.com"
	token       = "token"
	wrongValue  = "wrong_value"
	wrongID     = 0
	maxNameSize = 1024
)

var (
	twin = twins.Twin{
		Name:     "test_app",
		Metadata: map[string]interface{}{"test": "data"},
	}
	invalidName = strings.Repeat("m", maxNameSize+1)
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
	return tr.client.Do(req)
}

func newService(tokens map[string]string) twins.Service {
	users := mocks.NewUsersService(tokens)
	twinsRepo := mocks.NewTwinRepository()
	idp := mocks.NewIdentityProvider()

	opts := mqtt.NewClientOptions()
	mc := mqtt.NewClient(opts)

	return twins.New("secret", mc, users, twinsRepo, idp)
}

func newServer(svc twins.Service) *httptest.Server {
	mux := httpapi.MakeHandler(mocktracer.New(), svc)
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestAddTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	tw := twin
	tw.Key = "key"
	data := toJSON(tw)

	tw.Name = invalidName
	invalidData := toJSON(tw)

	cases := []struct {
		desc        string
		req         string
		contentType string
		auth        string
		status      int
		location    string
	}{
		{
			desc:        "add valid twin",
			req:         data,
			contentType: contentType,
			auth:        token,
			status:      http.StatusCreated,
			location:    "/twins/1",
		},
		{
			desc:        "add twin with existing key",
			req:         data,
			contentType: contentType,
			auth:        token,
			status:      http.StatusUnprocessableEntity,
			location:    "",
		},
		{
			desc:        "add twin with empty JSON request",
			req:         "{}",
			contentType: contentType,
			auth:        token,
			status:      http.StatusCreated,
			location:    "/twins/2",
		},
		{
			desc:        "add twin with invalid auth token",
			req:         data,
			contentType: contentType,
			auth:        wrongValue,
			status:      http.StatusForbidden,
			location:    "",
		},
		{
			desc:        "add twin with empty auth token",
			req:         data,
			contentType: contentType,
			auth:        "",
			status:      http.StatusForbidden,
			location:    "",
		},
		{
			desc:        "add twin with invalid request format",
			req:         "}",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
			location:    "",
		},
		{
			desc:        "add twin with empty request",
			req:         "",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
			location:    "",
		},
		{
			desc:        "add twin without content type",
			req:         data,
			contentType: "",
			auth:        token,
			status:      http.StatusUnsupportedMediaType,
			location:    "",
		},
		{
			desc:        "add twin with invalid name",
			req:         invalidData,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
			location:    "",
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/twins", ts.URL),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))

		location := res.Header.Get("Location")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.location, location, fmt.Sprintf("%s: expected location %s got %s", tc.desc, tc.location, location))
	}
}

func TestUpdateTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	data := toJSON(twin)
	stw, _ := svc.AddTwin(context.Background(), token, twin)

	tw := twin
	tw.Name = invalidName
	invalidData := toJSON(tw)

	cases := []struct {
		desc        string
		req         string
		id          string
		contentType string
		auth        string
		status      int
	}{
		{
			desc:        "update existing twin",
			req:         data,
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusOK,
		},
		{
			desc:        "update twin with empty JSON request",
			req:         "{}",
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusOK,
		},
		{
			desc:        "update non-existent twin",
			req:         data,
			id:          strconv.FormatUint(wrongID, 10),
			contentType: contentType,
			auth:        token,
			status:      http.StatusNotFound,
		},
		{
			desc:        "update twin with invalid id",
			req:         data,
			id:          "invalid",
			contentType: contentType,
			auth:        token,
			status:      http.StatusNotFound,
		},
		{
			desc:        "update twin with invalid user token",
			req:         data,
			id:          stw.ID,
			contentType: contentType,
			auth:        wrongValue,
			status:      http.StatusForbidden,
		},
		{
			desc:        "update twin with empty user token",
			req:         data,
			id:          stw.ID,
			contentType: contentType,
			auth:        "",
			status:      http.StatusForbidden,
		},
		{
			desc:        "update twin with invalid data format",
			req:         "{",
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update twin with empty request",
			req:         "",
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update twin without content type",
			req:         data,
			id:          stw.ID,
			contentType: "",
			auth:        token,
			status:      http.StatusUnsupportedMediaType,
		},
		{
			desc:        "update twin with invalid name",
			req:         invalidData,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/twins/%s", ts.URL, tc.id),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestUpdateKey(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	tw := twin
	tw.Key = "key"
	stw, _ := svc.AddTwin(context.Background(), token, tw)

	stw.Key = "new-key"
	data := toJSON(stw)

	stw.Key = "key"
	dummyData := toJSON(stw)

	cases := []struct {
		desc        string
		req         string
		id          string
		contentType string
		auth        string
		status      int
	}{
		{
			desc:        "update key for an existing twin",
			req:         data,
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusOK,
		},
		{
			desc:        "update twin with conflicting key",
			req:         data,
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusUnprocessableEntity,
		},
		{
			desc:        "update key with empty JSON request",
			req:         "{}",
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update key of non-existent twin",
			req:         dummyData,
			id:          strconv.FormatUint(wrongID, 10),
			contentType: contentType,
			auth:        token,
			status:      http.StatusNotFound,
		},
		{
			desc:        "update twin with invalid id",
			req:         dummyData,
			id:          "invalid",
			contentType: contentType,
			auth:        token,
			status:      http.StatusNotFound,
		},
		{
			desc:        "update twin with invalid user token",
			req:         data,
			id:          stw.ID,
			contentType: contentType,
			auth:        wrongValue,
			status:      http.StatusForbidden,
		},
		{
			desc:        "update twin with empty user token",
			req:         data,
			id:          stw.ID,
			contentType: contentType,
			auth:        "",
			status:      http.StatusForbidden,
		},
		{
			desc:        "update twin with invalid data format",
			req:         "{",
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update twin with empty request",
			req:         "",
			id:          stw.ID,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update twin without content type",
			req:         data,
			id:          stw.ID,
			contentType: "",
			auth:        token,
			status:      http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodPatch,
			url:         fmt.Sprintf("%s/twins/%s/key", ts.URL, tc.id),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

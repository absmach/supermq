// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mainflux/mainflux/twins"
	httpapi "github.com/mainflux/mainflux/twins/api/http"
	"github.com/mainflux/mainflux/twins/mocks"

	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	twinName    = "name"
	contentType = "application/json"
	email       = "user@example.com"
	token       = "token"
	wrongValue  = "wrong_value"
	wrongID     = 0
	maxNameSize = 1024
	topic       = "topic"
)

var invalidName = strings.Repeat("m", maxNameSize+1)

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

func newServer(svc twins.Service) *httptest.Server {
	mux := httpapi.MakeHandler(mocktracer.New(), svc)
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestAddTwin(t *testing.T) {
	svc := mocks.New(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	tw := twinReq{}
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
			location:    "/twins/123e4567-e89b-12d3-a456-000000000001",
		},
		{
			desc:        "add twin with empty JSON request",
			req:         "{}",
			contentType: contentType,
			auth:        token,
			status:      http.StatusCreated,
			location:    "/twins/123e4567-e89b-12d3-a456-000000000002",
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
	svc := mocks.New(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	twin := twins.Twin{}
	def := twins.Definition{}
	stw, _ := svc.AddTwin(context.Background(), token, twin, def)

	twin.Name = twinName
	data := toJSON(twin)

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
			status:      http.StatusBadRequest,
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

func TestViewTwin(t *testing.T) {
	svc := mocks.New(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	def := twins.Definition{}
	twin := twins.Twin{}
	stw, err := svc.AddTwin(context.Background(), token, twin, def)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	twres := twinRes{
		Owner:       stw.Owner,
		Name:        stw.Name,
		ID:          stw.ID,
		Revision:    stw.Revision,
		Created:     stw.Created,
		Updated:     stw.Updated,
		Definitions: stw.Definitions,
		Metadata:    stw.Metadata,
	}
	data := toJSON(twres)

	cases := []struct {
		desc   string
		id     string
		auth   string
		status int
		res    string
	}{
		{
			desc:   "view existing twin",
			id:     stw.ID,
			auth:   token,
			status: http.StatusOK,
			res:    data,
		},
		{
			desc:   "view non-existent twin",
			id:     strconv.FormatUint(wrongID, 10),
			auth:   token,
			status: http.StatusNotFound,
			res:    "",
		},
		{
			desc:   "view twin by passing invalid token",
			id:     stw.ID,
			auth:   wrongValue,
			status: http.StatusForbidden,
			res:    "",
		},
		{
			desc:   "view twin by passing empty id",
			id:     "",
			auth:   token,
			status: http.StatusBadRequest,
			res:    "",
		},
		{
			desc:   "view twin by passing empty token",
			id:     stw.ID,
			auth:   "",
			status: http.StatusForbidden,
			res:    "",
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/twins/%s", ts.URL, tc.id),
			token:  tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		body, err := ioutil.ReadAll(res.Body)
		data := strings.Trim(string(body), "\n")
		assert.Equal(t, tc.res, data, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, data))
	}
}

func TestListTwins(t *testing.T) {
	svc := mocks.New(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	data := []twinRes{}
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("%s-%d", twinName, i)
		twin := twins.Twin{
			Owner: email,
			Name:  name,
		}
		tw, err := svc.AddTwin(context.Background(), token, twin, twins.Definition{})
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		twres := twinRes{
			Owner:       tw.Owner,
			ID:          tw.ID,
			Name:        tw.Name,
			Revision:    tw.Revision,
			Created:     tw.Created,
			Updated:     tw.Updated,
			Definitions: tw.Definitions,
			Metadata:    tw.Metadata,
		}
		data = append(data, twres)
	}

	baseURL := fmt.Sprintf("%s/twins", ts.URL)
	queryFmt := "%s?offset=%d&limit=%d"
	cases := []struct {
		desc   string
		auth   string
		status int
		url    string
		res    []twinRes
	}{
		{
			desc:   "get a list of twins",
			auth:   token,
			status: http.StatusOK,
			url:    baseURL,
			res:    data[0:10],
		},
		{
			desc:   "get a list of twins with invalid token",
			auth:   wrongValue,
			status: http.StatusForbidden,
			url:    fmt.Sprintf(queryFmt, baseURL, 0, 1),
			res:    nil,
		},
		{
			desc:   "get a list of twins with empty token",
			auth:   "",
			status: http.StatusForbidden,
			url:    fmt.Sprintf(queryFmt, baseURL, 0, 1),
			res:    nil,
		},
		{
			desc:   "get a list of twins with valid offset and limit",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf(queryFmt, baseURL, 25, 40),
			res:    data[25:65],
		},
		{
			desc:   "get a list of twins with offset + limit > total",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf(queryFmt, baseURL, 91, 20),
			res:    data[91:],
		},
		{
			desc:   "get a list of twins with negative offset",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf(queryFmt, baseURL, -1, 5),
			res:    nil,
		},
		{
			desc:   "get a list of twins with negative limit",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf(queryFmt, baseURL, 1, -5),
			res:    nil,
		},
		{
			desc:   "get a list of twins with zero limit",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf(queryFmt, baseURL, 1, 0),
			res:    nil,
		},
		{
			desc:   "get a list of twins with limit greater than max",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s?offset=%d&limit=%d", baseURL, 0, 110),
			res:    nil,
		},
		{
			desc:   "get a list of twins with invalid offset",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s%s", baseURL, "?offset=e&limit=5"),
			res:    nil,
		},
		{
			desc:   "get a list of twins with invalid limit",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s%s", baseURL, "?offset=5&limit=e"),
			res:    nil,
		},
		{
			desc:   "get a list of twins without offset",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf("%s?limit=%d", baseURL, 5),
			res:    data[0:5],
		},
		{
			desc:   "get a list of twins without limit",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf("%s?offset=%d", baseURL, 1),
			res:    data[1:11],
		},
		{
			desc:   "get a list of twins with invalid number of parameters",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s%s", baseURL, "?offset=4&limit=4&limit=5&offset=5"),
			res:    nil,
		},
		{
			desc:   "get a list of twins with redundant query parameters",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf("%s?offset=%d&limit=%d&value=something", baseURL, 0, 5),
			res:    data[0:5],
		},
		{
			desc:   "get a list of twins filtering with invalid name",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s?offset=%d&limit=%d&name=%s", baseURL, 0, 5, invalidName),
			res:    nil,
		},
		{
			desc:   "get a list of twins filtering with valid name",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf("%s?offset=%d&limit=%d&name=%s", baseURL, 2, 1, twinName+"-2"),
			res:    data[2:3],
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodGet,
			url:    tc.url,
			token:  tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		var resData twinsPageRes
		json.NewDecoder(res.Body).Decode(&resData)
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.ElementsMatch(t, tc.res, resData.Twins, fmt.Sprintf("%s: expected body %v got %v", tc.desc, tc.res, resData.Twins))
	}
}

func TestRemoveTwin(t *testing.T) {
	svc := mocks.New(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	def := twins.Definition{}
	twin := twins.Twin{}
	stw, _ := svc.AddTwin(context.Background(), token, twin, def)

	cases := []struct {
		desc   string
		id     string
		auth   string
		status int
	}{
		{
			desc:   "delete existing twin",
			id:     stw.ID,
			auth:   token,
			status: http.StatusNoContent,
		},
		{
			desc:   "delete non-existent twin",
			id:     strconv.FormatUint(wrongID, 10),
			auth:   token,
			status: http.StatusNoContent,
		},
		{
			desc:   "delete twin by passing empty id",
			id:     "",
			auth:   token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "delete twin with invalid token",
			id:     stw.ID,
			auth:   wrongValue,
			status: http.StatusForbidden,
		},
		{
			desc:   "delete twin with empty token",
			id:     stw.ID,
			auth:   "",
			status: http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/twins/%s", ts.URL, tc.id),
			token:  tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

type twinReq struct {
	token      string
	Name       string                 `json:"name,omitempty"`
	Definition twins.Definition       `json:"definition,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type twinRes struct {
	Owner       string                 `json:"owner"`
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	Revision    int                    `json:"revision"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Definitions []twins.Definition     `json:"definitions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
}

type twinsPageRes struct {
	pageRes
	Twins []twinRes `json:"twins"`
}

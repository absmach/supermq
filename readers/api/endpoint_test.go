// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	svcName       = "test-service"
	token         = "1"
	invalid       = "invalid"
	numOfMessages = 100
	valueFields   = 5
	subtopic      = "topic"
	mqttProt      = "mqtt"
	httpProt      = "http"
	msgName       = "temperature"
)

var (
	v   float64 = 5
	vs          = "value"
	vb          = true
	vd          = "dataValue"
	sum float64 = 42

	idProvider = uuid.New()
)

func newServer(repo readers.MessageRepository, tc mainflux.ThingsServiceClient) *httptest.Server {
	mux := api.MakeHandler(repo, tc, svcName)
	return httptest.NewServer(mux)
}

type testRequest struct {
	client *http.Client
	method string
	url    string
	token  string
	body   io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, nil)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", tr.token)
	}

	return tr.client.Do(req)
}

func TestReadAll(t *testing.T) {
	chanID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	pubID, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	pubID2, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	now := float64(time.Now().UTC().Second())

	var messages []senml.Message
	var queryMsgs []senml.Message
	var valueMsgs []senml.Message
	var boolMsgs []senml.Message
	var stringMsgs []senml.Message
	var dataMsgs []senml.Message

	for i := 0; i < numOfMessages; i++ {
		// Mix possible values as well as value sum.
		msg := senml.Message{
			Channel:   chanID,
			Publisher: pubID,
			Protocol:  mqttProt,
			Time:      now - float64(i),
			Name:      "name",
		}

		count := i % valueFields
		switch count {
		case 0:
			msg.Value = &v
			valueMsgs = append(valueMsgs, msg)
		case 1:
			msg.BoolValue = &vb
			boolMsgs = append(boolMsgs, msg)
		case 2:
			msg.StringValue = &vs
			stringMsgs = append(stringMsgs, msg)
		case 3:
			msg.DataValue = &vd
			dataMsgs = append(dataMsgs, msg)
		case 4:
			msg.Sum = &sum
			msg.Subtopic = subtopic
			msg.Protocol = httpProt
			msg.Publisher = pubID2
			msg.Name = msgName
			queryMsgs = append(queryMsgs, msg)
		}

		messages = append(messages, msg)
	}

	svc := mocks.NewThingsService()
	repo := mocks.NewMessageRepository(chanID, fromSenml(messages))
	ts := newServer(repo, svc)
	defer ts.Close()

	cases := []struct {
		desc   string
		req    string
		url    string
		token  string
		status int
		res    pageRes
	}{
		{
			desc:   "read page with valid offset and limit",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: messages[0:10],
			},
		},
		{
			desc:   "read page with negative offset",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=-1&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with negative limit",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=-10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with zero limit",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=0", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with non-integer offset",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=abc&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with non-integer limit",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=abc", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with invalid channel id",
			url:    fmt.Sprintf("%s/channels//messages?offset=0&limit=10", ts.URL),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with invalid token",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=10", ts.URL, chanID),
			token:  invalid,
			status: http.StatusForbidden,
		},
		{
			desc:   "read page with multiple offset",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&offset=1&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with multiple limit",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=20&limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with empty token",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0&limit=10", ts.URL, chanID),
			token:  "",
			status: http.StatusForbidden,
		},
		{
			desc:   "read page with default offset",
			url:    fmt.Sprintf("%s/channels/%s/messages?limit=10", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: messages[0:10],
			},
		},
		{
			desc:   "read page with default limit",
			url:    fmt.Sprintf("%s/channels/%s/messages?offset=0", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: messages[0:10],
			},
		},
		{
			desc:   "read page with senml fornat",
			url:    fmt.Sprintf("%s/channels/%s/messages?format=messages", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: messages[0:10],
			},
		},
		{
			desc:   "read page with subtopic",
			url:    fmt.Sprintf("%s/channels/%s/messages?subtopic=%s", ts.URL, chanID, subtopic),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: queryMsgs[0:10],
			},
		},
		{
			desc:   "read page with publisher",
			url:    fmt.Sprintf("%s/channels/%s/messages?publisher=%s", ts.URL, chanID, pubID2),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: queryMsgs[0:10],
			},
		},
		{
			desc:   "read page with protocol",
			url:    fmt.Sprintf("%s/channels/%s/messages?protocol=http", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: queryMsgs[0:10],
			},
		},
		{
			desc:   "read page with name",
			url:    fmt.Sprintf("%s/channels/%s/messages?name=%s", ts.URL, chanID, msgName),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: queryMsgs[0:10],
			},
		},
		{
			desc:   "read page with value",
			url:    fmt.Sprintf("%s/channels/%s/messages?v=%f", ts.URL, chanID, v),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: valueMsgs[0:10],
			},
		},
		{
			desc:   "read page with non-float value",
			url:    fmt.Sprintf("%s/channels/%s/messages?v=ab01", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with boolean value",
			url:    fmt.Sprintf("%s/channels/%s/messages?vb=true", ts.URL, chanID),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: boolMsgs[0:10],
			},
		},
		{
			desc:   "read page with non-boolean value",
			url:    fmt.Sprintf("%s/channels/%s/messages?vb=yes", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with string value",
			url:    fmt.Sprintf("%s/channels/%s/messages?vs=%s", ts.URL, chanID, vs),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: stringMsgs[0:10],
			},
		},
		{
			desc:   "read page with data value",
			url:    fmt.Sprintf("%s/channels/%s/messages?vd=%s", ts.URL, chanID, vd),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: dataMsgs[0:10],
			},
		},
		{
			desc:   "read page with from",
			url:    fmt.Sprintf("%s/channels/%s/messages?from=%f", ts.URL, chanID, messages[9].Time),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: messages[0:10],
			},
		},
		{
			desc:   "read page with non-float from",
			url:    fmt.Sprintf("%s/channels/%s/messages?from=ABCD", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
		{
			desc:   "read page with to",
			url:    fmt.Sprintf("%s/channels/%s/messages?to=%f", ts.URL, chanID, messages[9].Time),
			token:  token,
			status: http.StatusOK,
			res: pageRes{
				Messages: messages[10:20],
			},
		},
		{
			desc:   "read page with non-float to",
			url:    fmt.Sprintf("%s/channels/%s/messages?to=ABCD", ts.URL, chanID),
			token:  token,
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodGet,
			url:    tc.url,
			token:  tc.token,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))

		var page pageRes
		json.NewDecoder(res.Body).Decode(&page)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.ElementsMatch(t, tc.res.Messages, page.Messages, fmt.Sprintf("%s: expected body %v got %v", tc.desc, tc.res.Messages, page.Messages))
	}
}

type pageRes struct {
	readers.PageMetadata
	Messages []senml.Message `json:"messages,omitempty"`
}

func fromSenml(in []senml.Message) []readers.Message {
	var ret []readers.Message
	for _, m := range in {
		ret = append(ret, m)
	}
	return ret
}

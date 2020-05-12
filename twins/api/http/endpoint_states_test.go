// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/mainflux/mainflux/messaging"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/senml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mainflux/mainflux/twins/uuid"
)

const (
	nanosec = 1e9

	attrName1     = "temperature"
	attrSubtopic1 = "engine"
	attrName2     = "humidity"
	attrSubtopic2 = "chassis"
)

type stateRes struct {
	TwinID     string                 `json:"twin_id"`
	ID         int64                  `json:"id"`
	Definition int                    `json:"definition"`
	Created    time.Time              `json:"created"`
	Payload    map[string]interface{} `json:"payload"`
}

type statesPageRes struct {
	pageRes
	States []stateRes `json:"states"`
}

func TestListStates(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	twin := twins.Twin{
		Owner: email,
	}
	attr1 := createAttribute(attrName1, attrSubtopic1)
	attr2 := createAttribute(attrName2, attrSubtopic2)
	def := twins.Definition{
		Attributes: []twins.Attribute{
			attr1,
			attr2,
		},
	}
	tw, _ := svc.AddTwin(context.Background(), token, twin, def)

	recs := createSenML(100, attrName1)
	message := createMessage(attr1, recs)
	err := svc.SaveStates(message)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	data := []stateRes{}
	for i := 0; i < len(recs); i++ {
		res := createStateResponse(i, tw, recs[i])
		data = append(data, res)
	}

	baseURL := fmt.Sprintf("%s/states/%s", ts.URL, tw.ID)
	queryFmt := "%s?offset=%d&limit=%d"
	cases := []struct {
		desc   string
		auth   string
		status int
		url    string
		res    []stateRes
	}{
		{
			desc:   "get a list of states",
			auth:   token,
			status: http.StatusOK,
			url:    baseURL,
			res:    data[0:10],
		},
		{
			desc:   "get a list of states with with valid offset and limit",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf(queryFmt, baseURL, 20, 15),
			res:    data[20:35],
		},
		{
			desc:   "get a list of states with invalid token",
			auth:   wrongValue,
			status: http.StatusForbidden,
			url:    fmt.Sprintf(queryFmt, baseURL, 0, 5),
			res:    nil,
		},
		{
			desc:   "get a list of states with empty token",
			auth:   "",
			status: http.StatusForbidden,
			url:    fmt.Sprintf(queryFmt, baseURL, 0, 5),
			res:    nil,
		},
		{
			desc:   "get a list of states with with  + limit > total",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf(queryFmt, baseURL, 91, 20),
			res:    data[91:],
		},
		{
			desc:   "get a list of states with negative offset",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf(queryFmt, baseURL, -1, 5),
			res:    nil,
		},
		{
			desc:   "get a list of states with negative limit",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf(queryFmt, baseURL, 0, -5),
			res:    nil,
		},
		{
			desc:   "get a list of states with zero limit",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf(queryFmt, baseURL, 0, 0),
			res:    nil,
		},
		{
			desc:   "get a list of states with limit greater than max",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf(queryFmt, baseURL, 0, 110),
			res:    nil,
		},
		{
			desc:   "get a list of states with invalid offset",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s?offset=invalid&limit=%d", baseURL, 15),
			res:    nil,
		},
		{
			desc:   "get a list of states with invalid limit",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s?offset=%d&limit=invalid", baseURL, 0),
			res:    nil,
		},
		{
			desc:   "get a list of states without offset",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf("%s?limit=%d", baseURL, 15),
			res:    data[0:15],
		},
		{
			desc:   "get a list of states without limit",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf("%s?offset=%d", baseURL, 14),
			res:    data[14:24],
		},
		{
			desc:   "get a list of states with invalid number of params",
			auth:   token,
			status: http.StatusBadRequest,
			url:    fmt.Sprintf("%s%s", baseURL, "?offset=4&limit=4&limit=5&offset=5"),
			res:    nil,
		},
		{
			desc:   "get a list of states with redundant query params",
			auth:   token,
			status: http.StatusOK,
			url:    fmt.Sprintf("%s?offset=%d&limit=%d&value=something", baseURL, 0, 5),
			res:    data[0:5],
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
		var resData statesPageRes
		json.NewDecoder(res.Body).Decode(&resData)
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.ElementsMatch(t, tc.res, resData.States, fmt.Sprintf("%s: expected body %v got %v", tc.desc, tc.res, resData.States))
	}
}

func createAttribute(name, subtopic string) twins.Attribute {
	id, _ := uuid.New().ID()
	return twins.Attribute{
		Name:         name,
		Channel:      id,
		Subtopic:     subtopic,
		PersistState: true,
	}
}

func createSenML(n int, bn string) []senml.Record {
	var recs []senml.Record
	bt := time.Now().Unix()
	for i := 0; i < n; i++ {
		rec := senml.Record{
			BaseName: bn,
			BaseTime: float64(bt),
			Time:     float64(i),
			Value:    nil,
		}
		recs = append(recs, rec)
	}
	return recs
}

func createMessage(attr twins.Attribute, recs []senml.Record) *messaging.Message {
	mRecs, _ := json.Marshal(recs)
	return &messaging.Message{
		Channel:   attr.Channel,
		Subtopic:  attr.Subtopic,
		Payload:   mRecs,
		Publisher: "twins",
	}
}

func createStateResponse(id int, tw twins.Twin, rec senml.Record) stateRes {
	recSec := rec.BaseTime + rec.Time
	sec, dec := math.Modf(recSec)
	recTime := time.Unix(int64(sec), int64(dec*nanosec))

	return stateRes{
		TwinID:     tw.ID,
		ID:         int64(id),
		Definition: tw.Definitions[len(tw.Definitions)-1].ID,
		Created:    recTime,
		Payload:    map[string]interface{}{rec.BaseName: nil},
	}
}

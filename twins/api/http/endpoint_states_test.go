// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
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
	millisec = 1e6
	nanosec  = 1e9

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

	recs := createSenML(100, "temperature")
	message := createMessage(attr1, recs)
	err := svc.SaveStates(message)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	data := []stateRes{}
	twDef := tw.Definitions[len(tw.Definitions)-1].ID
	for i := 0; i < len(recs); i++ {
		rec := recs[i]

		recSec := rec.BaseTime + rec.Time
		sec, dec := math.Modf(recSec)
		recTime := time.Unix(int64(sec), int64(dec*nanosec)).Round(0)

		pl := map[string]interface{}{
			rec.BaseName: rec.Value,
		}

		stres := stateRes{
			TwinID:     tw.ID,
			ID:         int64(i),
			Definition: twDef,
			Created:    recTime,
			Payload:    pl,
		}
		data = append(data, stres)
	}

	statesURL := fmt.Sprintf("%s/states/%s", ts.URL, tw.ID)
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
			url:    fmt.Sprintf(queryFmt, statesURL, 0, 5),
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
	bt := time.Now().Round(0).UnixNano()
	for i := 0; i < n; i++ {
		t := 10e-6 * float64(i)
		v := rand.Float64()
		rec := senml.Record{
			BaseName: bn,
			BaseTime: float64(bt),
			Time:     t,
			Value:    &v,
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

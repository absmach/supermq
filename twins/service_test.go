// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package twins_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	twinName   = "name"
	wrongID    = ""
	token      = "token"
	wrongToken = "wrong-token"
	email      = "user@example.com"
	natsURL    = "nats://localhost:4222"

	attrName1     = "temperature"
	attrSubtopic1 = "engine"
	attrName2     = "humidity"
	attrSubtopic2 = "chassis"
	numRecs       = 100
)

func TestAddTwin(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email})
	twin := twins.Twin{}
	def := twins.Definition{}

	cases := []struct {
		desc  string
		twin  twins.Twin
		token string
		err   error
	}{
		{
			desc:  "add new twin",
			twin:  twin,
			token: token,
			err:   nil,
		},
		{
			desc:  "add twin with wrong credentials",
			twin:  twin,
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		_, err := svc.AddTwin(context.Background(), tc.token, tc.twin, def)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestUpdateTwin(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email})
	twin := twins.Twin{}
	other := twins.Twin{}
	def := twins.Definition{}

	other.ID = wrongID
	saved, err := svc.AddTwin(context.Background(), token, twin, def)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	saved.Name = twinName

	cases := []struct {
		desc  string
		twin  twins.Twin
		token string
		err   error
	}{
		{
			desc:  "update existing twin",
			twin:  saved,
			token: token,
			err:   nil,
		},
		{
			desc:  "update twin with wrong credentials",
			twin:  saved,
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
		{
			desc:  "update non-existing twin",
			twin:  other,
			token: token,
			err:   twins.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := svc.UpdateTwin(context.Background(), tc.token, tc.twin, def)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestViewTwin(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email})
	twin := twins.Twin{}
	def := twins.Definition{}
	saved, err := svc.AddTwin(context.Background(), token, twin, def)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	cases := map[string]struct {
		id    string
		token string
		err   error
	}{
		"view existing twin": {
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		"view twin with wrong credentials": {
			id:    saved.ID,
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
		"view non-existing twin": {
			id:    wrongID,
			token: token,
			err:   twins.ErrNotFound,
		},
	}

	for desc, tc := range cases {
		_, err := svc.ViewTwin(context.Background(), tc.token, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestListTwins(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email})
	twin := twins.Twin{Name: twinName, Owner: email}
	def := twins.Definition{}
	m := make(map[string]interface{})
	m["serial"] = "123456"
	twin.Metadata = m

	n := uint64(10)
	for i := uint64(0); i < n; i++ {
		svc.AddTwin(context.Background(), token, twin, def)
	}

	cases := map[string]struct {
		token    string
		offset   uint64
		limit    uint64
		size     uint64
		metadata map[string]interface{}
		err      error
	}{
		"list all twins": {
			token:  token,
			offset: 0,
			limit:  n,
			size:   n,
			err:    nil,
		},
		"list with zero limit": {
			token:  token,
			limit:  0,
			offset: 0,
			size:   0,
			err:    nil,
		},
		"list with offset and limit": {
			token:  token,
			offset: 8,
			limit:  5,
			size:   2,
			err:    nil,
		},
		"list with wrong credentials": {
			token:  wrongToken,
			limit:  0,
			offset: n,
			err:    twins.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		page, err := svc.ListTwins(context.Background(), tc.token, tc.offset, tc.limit, twinName, tc.metadata)
		size := uint64(len(page.Twins))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected %d got %d\n", desc, tc.size, size))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestRemoveTwin(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email})
	twin := twins.Twin{}
	def := twins.Definition{}
	saved, err := svc.AddTwin(context.Background(), token, twin, def)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	cases := []struct {
		desc  string
		id    string
		token string
		err   error
	}{
		{
			desc:  "remove twin with wrong credentials",
			id:    saved.ID,
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
		{
			desc:  "remove existing twin",
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove removed twin",
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove non-existing twin",
			id:    wrongID,
			token: token,
			err:   nil,
		},
	}

	for _, tc := range cases {
		err := svc.RemoveTwin(context.Background(), tc.token, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestSaveStates(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email})

	twin := twins.Twin{Owner: email}
	def := mocks.CreateDefinition([]string{attrName1, attrName2}, []string{attrSubtopic1, attrSubtopic2})
	attr := def.Attributes[0]
	tw, err := svc.AddTwin(context.Background(), token, twin, def)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	recs := mocks.CreateSenML(numRecs, attrName1)
	var ttlAdded uint64

	cases := []struct {
		desc   string
		offset uint64
		limit  uint64
		err    error
	}{
		{
			desc:   "add 10 states",
			offset: 0,
			limit:  numRecs,
			err:    nil,
		},
		{
			desc:   "add 20 states",
			offset: 10,
			limit:  10,
			err:    nil,
		},
	}

	for _, tc := range cases {

		from := clamp(tc.offset, 0, numRecs)
		to := clamp(tc.offset+tc.limit, 0, numRecs)
		message, err := mocks.CreateMessage(attr, recs[from:to])
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		err = svc.SaveStates(message)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		ttlAdded += to - from

		page, err := svc.ListStates(context.TODO(), token, tc.offset, tc.limit, tw.ID)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		assert.Equal(t, ttlAdded, page.Total, fmt.Sprintf("%s: expected %s total got %s total\n", tc.desc, tc.err, err))
	}
}

func clamp(n, min, max uint64) uint64 {
	return uint64(math.Min(math.Max(float64(n), float64(min)), float64(max)))
}

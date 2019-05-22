//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package http

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/things/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddThingReqValidation(t *testing.T) {
	token, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	invalidName := strings.Repeat("0123456789", 50)

	validReq := addThingReq{
		token: token,
	}

	cases := map[string]struct {
		req addThingReq
		err error
	}{
		"valid thing addition request": {
			req: validReq,
			err: nil,
		},
		"missing token": {
			req: addThingReq{
				token: "",
			},
			err: things.ErrUnauthorizedAccess,
		},
		"valid name": {
			req: addThingReq{
				token: token,
				Name:  "thing_name",
			},
			err: nil,
		},
		"invalid name": {
			req: addThingReq{
				token: token,
				Name:  invalidName,
			},
			err: things.ErrMalformedEntity,
		},
	}

	for desc, tc := range cases {
		err := tc.req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestUpdateThingReqValidation(t *testing.T) {
	token, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	thid, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	validReq := updateThingReq{
		token: token,
		id:    thid,
	}

	cases := map[string]struct {
		req updateThingReq
		err error
	}{
		"valid thing update request": {
			req: validReq,
			err: nil,
		},
		"missing token": {
			req: updateThingReq{
				token: "",
				id:    thid,
			},
			err: things.ErrUnauthorizedAccess,
		},
		"empty thing id": {
			req: updateThingReq{
				token: token,
				id:    "",
			},
			err: things.ErrMalformedEntity,
		},
	}

	for desc, tc := range cases {
		err := tc.req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestUpdateKeyReqValidation(t *testing.T) {
	token, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	thing := things.Thing{ID: "1", Key: "key"}

	cases := map[string]struct {
		token string
		id    string
		key   string
		err   error
	}{
		"valid key update request": {
			token: token,
			id:    thing.ID,
			key:   thing.Key,
			err:   nil,
		},
		"missing token": {
			token: "",
			id:    thing.ID,
			key:   thing.Key,
			err:   things.ErrUnauthorizedAccess,
		},
		"empty thing id": {
			token: token,
			id:    "",
			key:   thing.Key,
			err:   things.ErrMalformedEntity,
		},
		"empty key": {
			token: token,
			id:    thing.ID,
			key:   "",
			err:   things.ErrMalformedEntity,
		},
	}

	for desc, tc := range cases {
		req := updateKeyReq{
			token: tc.token,
			id:    tc.id,
			Key:   tc.key,
		}

		err := req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestCreateChannelReqValidation(t *testing.T) {
	channel := things.Channel{}
	token, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	cases := map[string]struct {
		channel things.Channel
		token   string
		err     error
	}{
		"valid channel creation request": {
			channel: channel,
			token:   token,
			err:     nil,
		},
		"missing token": {
			channel: channel,
			token:   "",
			err:     things.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		req := createChannelReq{
			token: tc.token,
			Name:  tc.channel.Name,
		}

		err := req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestUpdateChannelReqValidation(t *testing.T) {
	token, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	channel := things.Channel{ID: "1"}

	cases := map[string]struct {
		channel things.Channel
		id      string
		token   string
		err     error
	}{
		"valid channel update request": {
			channel: channel,
			id:      channel.ID,
			token:   token,
			err:     nil,
		},
		"missing token": {
			channel: channel,
			id:      channel.ID,
			token:   "",
			err:     things.ErrUnauthorizedAccess,
		},
		"empty channel id": {
			channel: channel,
			id:      "",
			token:   token,
			err:     things.ErrMalformedEntity,
		},
	}

	for desc, tc := range cases {
		req := updateChannelReq{
			token: tc.token,
			id:    tc.id,
			Name:  tc.channel.Name,
		}

		err := req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestViewResourceReqValidation(t *testing.T) {
	token, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	id := uint64(1)

	cases := map[string]struct {
		id    string
		token string
		err   error
	}{
		"valid resource viewing request": {
			id:    strconv.FormatUint(id, 10),
			token: token,
			err:   nil,
		},
		"missing token": {
			id:    strconv.FormatUint(id, 10),
			token: "",
			err:   things.ErrUnauthorizedAccess,
		},
		"empty resource id": {
			id:    "",
			token: token,
			err:   things.ErrMalformedEntity,
		},
	}

	for desc, tc := range cases {
		req := viewResourceReq{tc.token, tc.id}
		err := req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestListResourcesReqValidation(t *testing.T) {
	token, err := uuid.New().ID()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	invalidName := strings.Repeat("0123456789", 50)
	value := uint64(10)

	cases := map[string]struct {
		token  string
		offset uint64
		limit  uint64
		name   string
		err    error
	}{
		"valid listing request": {
			token:  token,
			offset: value,
			limit:  value,
			err:    nil,
		},
		"missing token": {
			token:  "",
			offset: value,
			limit:  value,
			err:    things.ErrUnauthorizedAccess,
		},
		"zero limit": {
			token:  token,
			offset: value,
			limit:  0,
			err:    things.ErrMalformedEntity,
		},
		"too big limit": {
			token:  token,
			offset: value,
			limit:  20 * value,
			err:    things.ErrMalformedEntity,
		},
		"too long name": {
			token:  token,
			offset: value,
			limit:  value,
			name:   invalidName,
			err:    nil,
		},
	}

	for desc, tc := range cases {
		req := listResourcesReq{
			token:  tc.token,
			offset: tc.offset,
			limit:  tc.limit,
		}

		err := req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestConnectionReqValidation(t *testing.T) {
	cases := map[string]struct {
		token   string
		chanID  string
		thingID string
		err     error
	}{
		"valid token": {
			token:   "valid-token",
			chanID:  "1",
			thingID: "1",
			err:     nil,
		},
		"empty token": {
			token:   "",
			chanID:  "1",
			thingID: "1",
			err:     things.ErrUnauthorizedAccess,
		},
		"empty channel id": {
			token:   "valid-token",
			chanID:  "",
			thingID: "1",
			err:     things.ErrMalformedEntity,
		},
		"empty thing id": {
			token:   "valid-token",
			chanID:  "1",
			thingID: "",
			err:     things.ErrMalformedEntity,
		},
	}

	for desc, tc := range cases {
		req := connectionReq{
			token:   tc.token,
			chanID:  tc.chanID,
			thingID: tc.thingID,
		}

		err := req.validate()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

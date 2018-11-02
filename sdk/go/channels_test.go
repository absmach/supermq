//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package sdk_test

import (
	"fmt"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mainflux/mainflux/sdk/go"
	"github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux/things"
	httpapi "github.com/mainflux/mainflux/things/api/http"
	"github.com/mainflux/mainflux/things/mocks"
)

const (
	contentType = "application/json"
	email       = "user@example.com"
	token       = "token"
	wrongValue  = "wrong_value"
	wrongID     = 0
)

func newService(tokens map[string]string) things.Service {
	users := mocks.NewUsersService(tokens)
	thingsRepo := mocks.NewThingRepository()
	channelsRepo := mocks.NewChannelRepository(thingsRepo)
	chanCache := mocks.NewChannelCache()
	thingCache := mocks.NewThingCache()
	idp := mocks.NewIdentityProvider()

	return things.New(users, thingsRepo, channelsRepo, chanCache, thingCache, idp)
}

func newServer(svc things.Service) *httptest.Server {
	mux := httpapi.MakeHandler(svc)
	return httptest.NewServer(mux)
}

func TestCreateChannel(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()

	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "http",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	channel := sdk.Channel{ID: "1", Name: "test"}
	emptyChannel := sdk.Channel{}

	cases := []struct {
		desc     string
		channel  sdk.Channel
		token    string
		err      error
		location string
	}{
		{
			desc:     "create new channel",
			channel:  channel,
			token:    token,
			err:      nil,
			location: "/channels/1",
		},
		{
			desc:     "create new channel with empty token",
			channel:  channel,
			token:    "",
			err:      sdk.ErrUnauthorized,
			location: "",
		},
		{
			desc:     "create new channel with invalid token",
			channel:  channel,
			token:    wrongValue,
			err:      sdk.ErrUnauthorized,
			location: "",
		},
		{
			desc:     "create new empty channel",
			channel:  emptyChannel,
			token:    token,
			err:      nil,
			location: "/channels/2",
		},
	}

	for _, tc := range cases {
		loc, err := mainfluxSDK.CreateChannel(tc.channel, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.location, loc, fmt.Sprintf("%s: expected location %s got %s", tc.desc, tc.location, loc))

	}
}
func TestChannel(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "http",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	channel := sdk.Channel{ID: "1", Name: "test"}
	mainfluxSDK.CreateChannel(channel, token)

	cases := []struct {
		desc     string
		chId     string
		token    string
		err      error
		response sdk.Channel
	}{
		{
			desc:     "Get existing channel",
			chId:     "1",
			token:    token,
			err:      nil,
			response: channel,
		},
		{
			desc:     "Get non-existent channel",
			chId:     "43",
			token:    token,
			err:      sdk.ErrNotFound,
			response: sdk.Channel{},
		},
		{
			desc:     "Get channel with invalid token",
			chId:     "1",
			token:    wrongValue,
			err:      sdk.ErrUnauthorized,
			response: sdk.Channel{},
		},
	}

	for _, tc := range cases {
		respCh, err := mainfluxSDK.Channel(tc.chId, tc.token)

		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.response, respCh, fmt.Sprintf("%s: expected response channel %s, got %s", tc.desc, tc.response, respCh))
	}

}

func TestCahnnels(t *testing.T) {

	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "http",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}
	var channels []sdk.Channel

	mainfluxSDK := sdk.NewSDK(sdkConf)
	for i := 1; i < 101; i++ {

		channel := sdk.Channel{ID: strconv.Itoa(i), Name: "test"}
		mainfluxSDK.CreateChannel(channel, token)
		channels = append(channels, channel)
	}

	cases := []struct {
		desc     string
		token    string
		offset   uint64
		limit    uint64
		err      error
		response []sdk.Channel
	}{
		{
			desc:     "get a list of channels",
			token:    token,
			offset:   0,
			limit:    5,
			err:      nil,
			response: channels[0:5],
		},
		{
			desc:     "get a list of channels with invalid token",
			token:    wrongValue,
			offset:   0,
			limit:    5,
			err:      sdk.ErrUnauthorized,
			response: nil,
		},
		{
			desc:     "get a list of channels with empty token",
			token:    "",
			offset:   0,
			limit:    5,
			err:      sdk.ErrUnauthorized,
			response: nil,
		},
		{
			desc:     "get a list of channels with zero limit",
			token:    token,
			offset:   0,
			limit:    0,
			err:      sdk.ErrInvalidArgs,
			response: nil,
		},
		{
			desc:     "get a list of channels with limit greater than max",
			token:    token,
			offset:   0,
			limit:    110,
			err:      sdk.ErrInvalidArgs,
			response: nil,
		},
		{
			desc:     "get a list of channels with offset greater than max",
			token:    token,
			offset:   110,
			limit:    5,
			err:      nil,
			response: nil,
		},
		{
			desc:     "get a list of channels with invalid args (zero limit) and invalid token",
			token:    wrongValue,
			offset:   0,
			limit:    0,
			err:      sdk.ErrInvalidArgs,
			response: nil,
		},
	}
	for _, tc := range cases {
		respChs, err := mainfluxSDK.Channels(tc.token, tc.offset, tc.limit)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.response, respChs, fmt.Sprintf("%s: expected response channel %s, got %s", tc.desc, tc.response, respChs))

	}
}

func TestUpdateChannel(t *testing.T) {
	svc := newService(map[string]string{token: email})
	ts := newServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "http",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	channel := sdk.Channel{ID: "1", Name: "test"}
	mainfluxSDK.CreateChannel(channel, token)
	cases := []struct {
		desc     string
		chId     string
		token    string
		err      error
		response sdk.Channel
	}{
		{
			desc:     "update existing thing",
			chId:     "1",
			token:    token,
			err:      nil,
			response: channel,
		},
	}

	for _, tc := range cases {
		respCh, err := mainfluxSDK.UpdateChannel(tc.chId, tc.token)

		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.response, respCh, fmt.Sprintf("%s: expected response channel %s, got %s", tc.desc, tc.response, respCh))
	}

}

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

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

	"github.com/mainflux/mainflux"
	adapter "github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/http/api"
	"github.com/mainflux/mainflux/http/mocks"
	sdk "github.com/mainflux/mainflux/sdk/go"
	"github.com/stretchr/testify/assert"
)

func newMessageService() mainflux.MessagePublisher {
	pub := mocks.NewPublisher()
	return adapter.New(pub)
}

func newMessageServer(pub mainflux.MessagePublisher, cc mainflux.ThingsServiceClient) *httptest.Server {
	mux := api.MakeHandler(pub, cc)
	return httptest.NewServer(mux)
}

func TestSendMessage(t *testing.T) {
	chanID := "1"
	//	invalidID := "wrong"
	token := "auth_token"
	//	invalidToken := "invalid_token"
	msg := `[{"n":"current","t":-1,"v":1.6}]`
	id, _ := strconv.ParseUint(chanID, 10, 64)
	thingsClient := mocks.NewThingsClient(map[string]uint64{token: id})
	pub := newMessageService()
	ts := newMessageServer(pub, thingsClient)
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

	cases := map[string]struct {
		chanID string
		msg    string
		auth   string
		err    error
	}{
		"publish message": {
			chanID: chanID,
			msg:    msg,
			auth:   token,
			err:    nil,
		},
	}
	for desc, tc := range cases {
		err := mainfluxSDK.SendMessage(tc.chanID, tc.msg, tc.auth)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", desc, tc.err, err))

	}
}

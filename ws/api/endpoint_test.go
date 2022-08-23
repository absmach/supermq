// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/ws"

	httpmock "github.com/mainflux/mainflux/http/mocks"
	"github.com/mainflux/mainflux/ws/api"
	"github.com/mainflux/mainflux/ws/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	chanID   = "1"
	id       = "1"
	thingKey = "thing_key"
	protocol = "ws"
)

var msg = []byte(`[{"n":"current","t":-1,"v":1.6}]`)

func newService(cc mainflux.ThingsServiceClient) ws.Service {
	pubsub := mocks.NewPubSub()
	return ws.New(cc, pubsub)
}

func newHTTPServer(svc ws.Service) *httptest.Server {
	logger := log.NewMock()
	mux := api.MakeHandler(svc, logger)
	return httptest.NewServer(mux)
}

func makeURL(tsURL, chanID, subtopic, thingKey string, header bool) (string, error) {
	u, _ := url.Parse(tsURL)
	u.Scheme = protocol

	if chanID == "0" || chanID == "" {
		if header {
			return fmt.Sprintf("%s/channels/%s/messages", u, chanID), fmt.Errorf("invalid channel id")
		}
		return fmt.Sprintf("%s/channels/%s/messages?authorization=%s", u, chanID, thingKey), fmt.Errorf("invalid channel id")
	}

	subtopicPart := ""
	if subtopic != "" {
		subtopicPart = fmt.Sprintf("/%s", subtopic)
	}
	if header {
		return fmt.Sprintf("%s/channels/%s/messages%s", u, chanID, subtopicPart), nil
	}

	return fmt.Sprintf("%s/channels/%s/messages%s?authorization=%s", u, chanID, subtopicPart, thingKey), nil
}

func handshake(tsURL, chanID, subtopic, thingKey string, addHeader bool) (*websocket.Conn, *http.Response, error) {
	header := http.Header{}
	if addHeader {
		header.Add("Authorization", thingKey)
	}

	url, _ := makeURL(tsURL, chanID, subtopic, thingKey, addHeader)
	conn, res, errRet := websocket.DefaultDialer.Dial(url, header)

	return conn, res, errRet
}

func TestHandshake(t *testing.T) {
	thingsClient := httpmock.NewThingsClient(map[string]string{thingKey: chanID})
	svc := newService(thingsClient)
	ts := newHTTPServer(svc)
	defer ts.Close()

	cases := []struct {
		name     string
		chanID   string
		subtopic string
		header   bool
		thingKey string
		status   int
		msg      []byte
	}{
		{
			name:     "connect and send message",
			chanID:   id,
			subtopic: "",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			name:     "connect to empty channel",
			chanID:   "",
			subtopic: "",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusBadRequest,
			msg:      []byte{},
		},
		{
			name:     "connect with empty thingKey",
			chanID:   id,
			subtopic: "",
			header:   true,
			thingKey: "",
			status:   http.StatusForbidden,
			msg:      []byte{},
		},
		{
			name:     "connect and send message with thingKey as query parameter",
			chanID:   id,
			subtopic: "",
			header:   false,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			name:     "connect and send message that cannot be published",
			chanID:   id,
			subtopic: "",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      []byte{},
		},
		{
			name:     "connect and send message to subtopic",
			chanID:   id,
			subtopic: "subtopic",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			name:     "connect and send message to subtopic with invalid name",
			chanID:   id,
			subtopic: "sub/a*b/topic",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusBadRequest,
			msg:      msg,
		},
		{
			name:     "connect and send message to nested subtopic",
			chanID:   id,
			subtopic: "subtopic/nested",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			name:     "connect and send message to all subtopics",
			chanID:   id,
			subtopic: ">",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			conn, res, err := handshake(ts.URL, tt.chanID, tt.subtopic, tt.thingKey, tt.header)
			assert.Equal(t, tt.status, res.StatusCode, fmt.Sprintf("expected status code '%d' got '%d'\n", tt.status, res.StatusCode))
			if err != nil {
				return
			}

			err = conn.WriteMessage(websocket.TextMessage, tt.msg)
			assert.Nil(t, err, fmt.Sprintf("unexpected error %s\n", err))
		})
	}
}

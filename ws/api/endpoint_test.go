// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

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
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"

	// "github.com/mainflux/mainflux/ws/api"
	"github.com/mainflux/mainflux/ws/api"
	"github.com/mainflux/mainflux/ws/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	chanID   = "1"
	id       = "1"
	thingKey = "thing_key"
	protocol = "ws"
	// invalidKey = "invalid_key"
	// stringMsg  = `[{"n":"current","t":-1,"v":1.6}]`
	// msgJSON    = `{"field1":"val1","field2","val2"}`
)

var mssg = []byte(`[{"n":"current","t":-1,"v":1.6}]`)

var msg = messaging.Message{
	Channel:   chanID,
	Publisher: id,
	Subtopic:  "",
	Protocol:  protocol,
	Payload:   []byte(`[{"n":"current","t":-5,"v":1.2}]`),
}

func newService(cc mainflux.ThingsServiceClient) ws.Service {
	pubsub := mocks.NewPubSub()
	return ws.New(cc, pubsub)
}

func newHTTPServer(svc ws.Service) *httptest.Server {
	logger := log.NewMock()
	mux := api.MakeHandler(svc, logger)
	return httptest.NewServer(mux)
}

func NewThingsClient() mainflux.ThingsServiceClient {
	return mocks.NewThingsClient(map[string]string{thingKey: chanID})
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

	// subtopic, err := parseSubTopic(subtopic)
	// if err != nil {
	// 	if header {
	// 		return fmt.Sprintf("%s/channels/%s/messages", u, chanID), err
	// 	}
	// 	return fmt.Sprintf("%s/channels/%s/messages?authorization=%s", u, chanID, thingKey), err
	// }

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

	// if thingKey == mocks.ServiceErrToken {
	// 	res.StatusCode = http.StatusServiceUnavailable
	// } else if thingKey == "invalid" {
	// 	res.StatusCode = http.StatusForbidden
	// }

	// if err != nil {
	// 	res.StatusCode = http.StatusBadRequest
	// }
	return conn, res, errRet
}

func TestHandshake(t *testing.T) {
	thingsClient := NewThingsClient()
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
		{"connect and send message", id, "", true, thingKey, http.StatusSwitchingProtocols, mssg},
		{"connect to non-existent channel", "0", "", true, thingKey, http.StatusBadRequest, []byte{}},
		{"connect to invalid channel id", "", "", true, thingKey, http.StatusBadRequest, []byte{}},
		{"connect with empty thingKey", id, "", true, "", http.StatusForbidden, []byte{}},
		{"connect with invalid thingKey", id, "", true, "invalid", http.StatusForbidden, []byte{}},
		{"connect unable to authorize", id, "", true, mocks.ServiceErrToken, http.StatusServiceUnavailable, []byte{}},
		{"connect and send message with thingKey as query parameter", id, "", false, thingKey, http.StatusSwitchingProtocols, mssg},
		{"connect and send message that cannot be published", id, "", true, thingKey, http.StatusSwitchingProtocols, []byte{}},
		{"connect and send message to subtopic", id, "subtopic", true, thingKey, http.StatusSwitchingProtocols, mssg},
		{"connect and send message to subtopic with invalid name", id, "sub/a*b/topic", true, thingKey, http.StatusBadRequest, mssg},
		{"connect and send message to nested subtopic", id, "subtopic/nested", true, thingKey, http.StatusSwitchingProtocols, mssg},
		{"connect and send message to all subtopics", id, ">", true, thingKey, http.StatusSwitchingProtocols, mssg},
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

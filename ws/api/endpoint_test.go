// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/ws"

	httpmock "github.com/mainflux/mainflux/http/mocks"
	"github.com/mainflux/mainflux/ws/api"
	"github.com/mainflux/mainflux/ws/mocks"
	"github.com/stretchr/testify/assert"
)

// const (
// 	chanID   = "1"
// 	id       = "1"
// 	thingKey = "thing_key"
// 	protocol = "ws"
// )

const (
	chanID    = "30315311-56ba-484d-b500-c1e08305511f"
	id        = "1"
	thingKey  = "c02ff576-ccd5-40f6-ba5f-c85377aad529"
	thingKey2 = "120ce059-0a8b-4fe7-8db9-85bdd1d3aece"
	subTopic  = "subtopic"
	protocol  = "ws"
	socketUrl = "ws://localhost:8190/channels/" + chanID + "/messages/?authorization=" + thingKey
)

var msg = []byte(`[{"n":"current","t":-1,"v":1.6}]`)

var (
	mssg = messaging.Message{
		Channel:   chanID,
		Publisher: "2",
		Subtopic:  "",
		Protocol:  protocol,
		Payload:   []byte(`[{"n":"current","t":-5,"v":1.2}]`),
	}
	msgChan = make(chan []byte)
	done    = make(chan interface{})
)

func newService(cc mainflux.ThingsServiceClient) (ws.Service, mocks.MockPubSub) {
	pubsub := mocks.NewPubSub()
	return ws.New(cc, pubsub), pubsub
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
	svc, _ := newService(thingsClient)
	ts := newHTTPServer(svc)
	defer ts.Close()

	cases := []struct {
		desc     string
		chanID   string
		subtopic string
		header   bool
		thingKey string
		status   int
		err      error
		msg      []byte
	}{
		{
			desc:     "connect and send message",
			chanID:   id,
			subtopic: "",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			desc:     "connect to empty channel",
			chanID:   "",
			subtopic: "",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusBadRequest,
			msg:      []byte{},
		},
		{
			desc:     "connect with empty thingKey",
			chanID:   id,
			subtopic: "",
			header:   true,
			thingKey: "",
			status:   http.StatusForbidden,
			msg:      []byte{},
		},
		{
			desc:     "connect and send message with thingKey as query parameter",
			chanID:   id,
			subtopic: "",
			header:   false,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			desc:     "connect and send message that cannot be published",
			chanID:   id,
			subtopic: "",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      []byte{},
		},
		{
			desc:     "connect and send message to subtopic",
			chanID:   id,
			subtopic: "subtopic",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			desc:     "connect and send message to subtopic with invalid name",
			chanID:   id,
			subtopic: "sub/a*b/topic",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusBadRequest,
			msg:      msg,
		},
		{
			desc:     "connect and send message to nested subtopic",
			chanID:   id,
			subtopic: "subtopic/nested",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
		{
			desc:     "connect and send message to all subtopics",
			chanID:   id,
			subtopic: ">",
			header:   true,
			thingKey: thingKey,
			status:   http.StatusSwitchingProtocols,
			msg:      msg,
		},
	}

	for _, tc := range cases {
		conn, res, err := handshake(ts.URL, tc.chanID, tc.subtopic, tc.thingKey, tc.header)
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code '%d' got '%d'\n", tc.desc, tc.status, res.StatusCode))

		if tc.status == http.StatusSwitchingProtocols {
			assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error %s\n", tc.desc, err))

			err = conn.WriteMessage(websocket.TextMessage, tc.msg)
			assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error %s\n", tc.desc, err))
		}
	}
}

func TestWebsocketConn(t *testing.T) {
	// server
	thingsClient := httpmock.NewThingsClient(map[string]string{thingKey: chanID, thingKey2: chanID})
	svc, pubsub := newService(thingsClient)
	logger := log.NewMock()

	errCh := make(chan error, 2)
	server := &http.Server{Addr: ":8190", Handler: api.MakeHandler(svc, logger)}
	go func() {
		errCh <- server.ListenAndServe()
	}()

	// client
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
	assert.Nil(t, err, fmt.Sprintln("expected nil error when creating new ws connection"))
	if err != nil {
		return
	}
	defer conn.Close()
	assert.Nil(t, err, fmt.Sprintln("expected no error while closing"))

	c := ws.NewClient(conn)
	//! c.id = "enter_id" (To enter id here, signature of NewClient, will need to changed back, to accept two arguments)
	fmt.Println("conn:", c)
	pubsub.SetConn(conn)

	// subscribe
	//? Useless to subscribe, as mockPubSub.Subscribe() just returns nil error
	// err = svc.Subscribe(context.Background(), thingKey, chanID, "", c)
	// assert.Nil(t, err, fmt.Sprintln("expected nil error when creating new ws connection. Got: "))
	// err = svc.Subscribe(context.Background(), thingKey, chanID, subTopic, c)
	// assert.Nil(t, err, fmt.Sprintln("expected nil error when creating new ws connection"))

	// listen for received messages, then push them to channel
	go receiveHandler(t, conn)

	err = svc.Publish(context.Background(), thingKey2, mssg)
	assert.Nil(t, err, fmt.Sprintln("expected nil error when publishing"))
	data := []byte{}
	select {
	case a := <-msgChan:
		fmt.Println("in select received: ", a)
		data = a
	case <-done:
		fmt.Println("inside select: exited with defer channel done")
	case <-time.After(time.Duration(20) * time.Second): // Did not receive anything from done channel
		fmt.Println("Timeout in closing receiving channel. Exiting...")
	}
	fmt.Println("data ->", data)
	receivedMsg := messaging.Message{}
	err = json.Unmarshal(data, &receivedMsg)
	assert.Nil(t, err, fmt.Sprintf("expected no error while unmarshalling: %s", err))

	assert.Equal(t, mssg, receivedMsg, fmt.Sprintf("expected %+v, got %+v", mssg, receivedMsg))

	// Write the piece of code below for 4 test cases
	// 1. Publishing for different clientID -> Get message response from receiveHandler
	// 2. Publishing for the same clientID -> No response from receiveHandler -> timeout
	// 3. Add test where websocket.ReadMessage fails
	// 4. Add test isunexpected close error -> close the websocket connection in between

}

//
func receiveHandler(t *testing.T, connection *websocket.Conn) {
	defer close(done)

	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			fmt.Println("readmessage error, exiting: ", err.Error(), " <- error")
			break
		}
		// assert.Nil(t, err, fmt.Sprintf("unexpected error while reading ws message: %s", err))
		fmt.Println("ReadMessage() error -> ", err)
		fmt.Println("ReadMessage() msg = -> ", string(msg))
		msgChan <- msg
	}
}

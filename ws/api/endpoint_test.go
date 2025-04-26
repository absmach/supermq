// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/absmach/mgate"
	mHttp "github.com/absmach/mgate/pkg/http"
	"github.com/absmach/mgate/pkg/session"
	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	chmocks "github.com/absmach/supermq/channels/mocks"
	climocks "github.com/absmach/supermq/clients/mocks"
	"github.com/absmach/supermq/internal/testsutil"
	smqlog "github.com/absmach/supermq/logger"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	authnMocks "github.com/absmach/supermq/pkg/authn/mocks"
	"github.com/absmach/supermq/pkg/messaging/mocks"
	"github.com/absmach/supermq/ws"
	"github.com/absmach/supermq/ws/api"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	id         = "1"
	clientKey  = "c02ff576-ccd5-40f6-ba5f-c85377aad529"
	protocol   = "ws"
	instanceID = "5de9b29a-feb9-11ed-be56-0242ac120002"
)

var (
	msg      = []byte(`[{"n":"current","t":-1,"v":1.6}]`)
	domainID = testsutil.GenerateUUID(&testing.T{})
)

func newService(clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient) (ws.Service, *mocks.PubSub) {
	pubsub := new(mocks.PubSub)
	return ws.New(clients, channels, pubsub), pubsub
}

func newHTTPServer(svc ws.Service) *httptest.Server {
	mux := api.MakeHandler(context.Background(), svc, smqlog.NewMock(), instanceID)
	return httptest.NewServer(mux)
}

func newProxyHTPPServer(svc session.Handler, targetServer *httptest.Server) (*httptest.Server, error) {
	turl := strings.ReplaceAll(targetServer.URL, "http", "ws")
	ptUrl, _ := url.Parse(turl)
	ptHost, ptPort, _ := net.SplitHostPort(ptUrl.Host)
	config := mgate.Config{
		Host:           "",
		Port:           "",
		PathPrefix:     "",
		TargetHost:     ptHost,
		TargetPort:     ptPort,
		TargetProtocol: ptUrl.Scheme,
		TargetPath:     ptUrl.Path,
	}
	mp, err := mHttp.NewProxy(config, svc, smqlog.NewMock(), []string{}, []string{})
	if err != nil {
		return nil, err
	}
	return httptest.NewServer(http.HandlerFunc(mp.ServeHTTP)), nil
}

func makeURL(tsURL, domainID, chanID, subtopic, clientKey string, header bool) (string, error) {
	u, _ := url.Parse(tsURL)
	u.Scheme = protocol

	if chanID == "0" || chanID == "" {
		if header {
			return fmt.Sprintf("%s/m/%s/c/%s", u, domainID, chanID), fmt.Errorf("invalid channel id")
		}
		return fmt.Sprintf("%s/m/%s/c/%s?authorization=%s", u, domainID, chanID, clientKey), fmt.Errorf("invalid channel id")
	}

	subtopicPart := ""
	if subtopic != "" {
		subtopicPart = fmt.Sprintf("/%s", subtopic)
	}
	if header {
		return fmt.Sprintf("%s/m/%s/c/%s%s", u, domainID, chanID, subtopicPart), nil
	}

	return fmt.Sprintf("%s/m/%s/c/%s%s?authorization=%s", u, domainID, chanID, subtopicPart, clientKey), nil
}

func handshake(tsURL, domainID, chanID, subtopic, clientKey string, addHeader bool) (*websocket.Conn, *http.Response, error) {
	header := http.Header{}
	if addHeader {
		header.Add("Authorization", clientKey)
	}

	turl, _ := makeURL(tsURL, domainID, chanID, subtopic, clientKey, addHeader)
	conn, res, errRet := websocket.DefaultDialer.Dial(turl, header)

	return conn, res, errRet
}

func TestHandshake(t *testing.T) {
	clients := new(climocks.ClientsServiceClient)
	channels := new(chmocks.ChannelsServiceClient)
	authn := new(authnMocks.Authentication)
	svc, pubsub := newService(clients, channels)
	target := newHTTPServer(svc)
	defer target.Close()
	handler := ws.NewHandler(pubsub, smqlog.NewMock(), authn, clients, channels)
	ts, err := newProxyHTPPServer(handler, target)
	require.Nil(t, err)
	defer ts.Close()
	pubsub.On("Subscribe", mock.Anything, mock.Anything).Return(nil)
	pubsub.On("Unsubscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	pubsub.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	clients.On("Authenticate", mock.Anything, mock.MatchedBy(func(req *grpcClientsV1.AuthnReq) bool {
		return req.ClientSecret == clientKey
	})).Return(&grpcClientsV1.AuthnRes{Authenticated: true}, nil)
	clients.On("Authenticate", mock.Anything, mock.Anything).Return(&grpcClientsV1.AuthnRes{Authenticated: false}, nil)
	authn.On("Authenticate", mock.Anything, mock.Anything).Return(smqauthn.Session{}, nil)
	channels.On("Authorize", mock.Anything, mock.Anything, mock.Anything).Return(&grpcChannelsV1.AuthzRes{Authorized: true}, nil)

	cases := []struct {
		desc      string
		domainID  string
		chanID    string
		subtopic  string
		header    bool
		clientKey string
		status    int
		err       error
		msg       []byte
	}{
		{
			desc:      "connect and send message",
			domainID:  domainID,
			chanID:    id,
			subtopic:  "",
			header:    true,
			clientKey: clientKey,
			status:    http.StatusSwitchingProtocols,
			msg:       msg,
		},
		{
			desc:      "connect and send message with clientKey as query parameter",
			domainID:  domainID,
			chanID:    id,
			subtopic:  "",
			header:    false,
			clientKey: clientKey,
			status:    http.StatusSwitchingProtocols,
			msg:       msg,
		},
		{
			desc:      "connect and send message that cannot be published",
			domainID:  domainID,
			chanID:    id,
			subtopic:  "",
			header:    true,
			clientKey: clientKey,
			status:    http.StatusSwitchingProtocols,
			msg:       []byte{},
		},
		{
			desc:      "connect and send message to subtopic",
			domainID:  domainID,
			chanID:    id,
			subtopic:  "subtopic",
			header:    true,
			clientKey: clientKey,
			status:    http.StatusSwitchingProtocols,
			msg:       msg,
		},
		{
			desc:      "connect and send message to nested subtopic",
			domainID:  domainID,
			chanID:    id,
			subtopic:  "subtopic/nested",
			header:    true,
			clientKey: clientKey,
			status:    http.StatusSwitchingProtocols,
			msg:       msg,
		},
		{
			desc:      "connect and send message to all subtopics",
			domainID:  domainID,
			chanID:    id,
			subtopic:  ">",
			header:    true,
			clientKey: clientKey,
			status:    http.StatusSwitchingProtocols,
			msg:       msg,
		},
		{
			desc:      "connect to empty channel",
			domainID:  domainID,
			chanID:    "",
			subtopic:  "",
			header:    true,
			clientKey: clientKey,
			status:    http.StatusBadRequest,
			msg:       []byte{},
		},
		{
			desc:      "connect with empty clientKey",
			domainID:  domainID,
			chanID:    id,
			subtopic:  "",
			header:    true,
			clientKey: "",
			status:    http.StatusUnauthorized,
			msg:       []byte{},
		},
		{
			desc:      "connect and send message to subtopic with invalid name",
			domainID:  domainID,
			chanID:    id,
			subtopic:  "sub/a*b/topic",
			header:    true,
			clientKey: clientKey,
			status:    http.StatusBadGateway,
			msg:       msg,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			conn, res, err := handshake(ts.URL, tc.domainID, tc.chanID, tc.subtopic, tc.clientKey, tc.header)
			assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code '%d' got '%d'\n", tc.desc, tc.status, res.StatusCode))

			if tc.status == http.StatusSwitchingProtocols {
				assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error %s\n", tc.desc, err))

				err = conn.WriteMessage(websocket.TextMessage, tc.msg)
				assert.Nil(t, err, fmt.Sprintf("%s: got unexpected error %s\n", tc.desc, err))
			}
		})
	}
}

// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	apiutil "github.com/absmach/supermq/api/http/util"
	chmocks "github.com/absmach/supermq/channels/mocks"
	climocks "github.com/absmach/supermq/clients/mocks"
	smqhttp "github.com/absmach/supermq/http"
	"github.com/absmach/supermq/internal/testsutil"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	authnmocks "github.com/absmach/supermq/pkg/authn/mocks"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/messaging/mocks"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	invalidID    = "invalidID"
	invalidKey   = "invalidKey"
	id           = "1"
	clientKey    = "client_key"
	subTopic     = "subtopic"
	protocol     = "ws"
	token        = "Bearer token"
	invalidToken = "Bearer invalid_token"
)

var (
	domainID = testsutil.GenerateUUID(&testing.T{})
	clientID = testsutil.GenerateUUID(&testing.T{})
	userID   = testsutil.GenerateUUID(&testing.T{})
	chanID   = testsutil.GenerateUUID(&testing.T{})
	msg      = messaging.Message{
		Channel:   chanID,
		Domain:    domainID,
		Publisher: id,
		Subtopic:  "",
		Protocol:  protocol,
		Payload:   []byte(`[{"n":"current","t":-5,"v":1.2}]`),
	}
	sessionID           = "sessionID"
	validEncodedCreds   = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientKey)))
	invalidEncodedCreds = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", invalidID, invalidKey)))
)

func newService() (smqhttp.Service, *mocks.PubSub, *climocks.ClientsServiceClient, *chmocks.ChannelsServiceClient, *authnmocks.Authentication) {
	pubsub := new(mocks.PubSub)
	clients := new(climocks.ClientsServiceClient)
	channels := new(chmocks.ChannelsServiceClient)
	authn := new(authnmocks.Authentication)

	return smqhttp.NewService(clients, channels, authn, pubsub), pubsub, clients, channels, authn
}

func TestSubscribe(t *testing.T) {
	svc, pubsub, clients, channels, auth := newService()

	c := smqhttp.NewClient(slog.Default(), nil, sessionID)

	cases := []struct {
		desc       string
		username   string
		password   string
		chanID     string
		domainID   string
		subtopic   string
		clientType string
		clientID   string
		topicType  messaging.TopicType
		authNToken string
		authNRes   *grpcClientsV1.AuthnRes
		authNErr   error
		authNRes1  smqauthn.Session
		authZRes   *grpcChannelsV1.AuthzRes
		authZErr   error
		subErr     error
		err        error
	}{
		{
			desc:       "subscribe to channel with valid clientKey, chanID, subtopic",
			password:   clientKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: true},
			err:        nil,
		},
		{
			desc:      "subscribe to channel with valid token, chanID, subtopic",
			password:  token,
			chanID:    chanID,
			domainID:  domainID,
			clientID:  userID,
			subtopic:  subTopic,
			topicType: messaging.MessageType,
			authNRes1: smqauthn.Session{UserID: userID},
			authZRes:  &grpcChannelsV1.AuthzRes{Authorized: true},
			err:       nil,
		},
		{
			desc:      "subscribe to channel with invalid token",
			password:  invalidToken,
			chanID:    chanID,
			domainID:  domainID,
			subtopic:  subTopic,
			topicType: messaging.MessageType,
			authNRes1: smqauthn.Session{},
			authNErr:  svcerr.ErrAuthentication,
			err:       svcerr.ErrAuthorization,
		},
		{
			desc:       "subscribe again to channel with valid clientKey, chanID, subtopic",
			password:   clientKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: true},
			err:        nil,
		},
		{
			desc:       "subscribe to channel with subscribe set to fail",
			password:   clientKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			subErr:     smqhttp.ErrFailedSubscription,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: true},
			err:        smqhttp.ErrFailedSubscription,
		},
		{
			desc:       "subscribe to channel with invalid clientKey",
			password:   invalidKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, invalidKey),
			authNRes:   &grpcClientsV1.AuthnRes{Authenticated: false},
			authNErr:   svcerr.ErrAuthentication,
			err:        svcerr.ErrAuthorization,
		},
		{
			desc:      "subscribe to channel with empty channel",
			password:  clientKey,
			chanID:    "",
			domainID:  domainID,
			clientID:  clientID,
			subtopic:  subTopic,
			topicType: messaging.MessageType,
			err:       svcerr.ErrAuthentication,
		},
		{
			desc:      "subscribe to channel with empty clientKey",
			password:  "",
			chanID:    chanID,
			domainID:  domainID,
			clientID:  clientID,
			subtopic:  subTopic,
			topicType: messaging.MessageType,
			err:       svcerr.ErrAuthentication,
		},
		{
			desc:      "subscribe to channel with empty clientKey and empty channel",
			password:  "",
			chanID:    "",
			domainID:  domainID,
			clientID:  clientID,
			subtopic:  subTopic,
			topicType: messaging.MessageType,
			err:       svcerr.ErrAuthentication,
		},
		{
			desc:       "subscribe to channel with invalid channel",
			password:   clientKey,
			chanID:     invalidID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: false},
			authZErr:   svcerr.ErrAuthorization,
			err:        svcerr.ErrAuthorization,
		},
		{
			desc:       "subscribe to channel with failed authentication",
			password:   clientKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Authenticated: false},
			err:        svcerr.ErrAuthorization,
		},
		{
			desc:       "subscribe to channel with failed authorization",
			password:   clientKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: false},
			err:        svcerr.ErrAuthorization,
		},
		{
			desc:       "subscribe to channel with valid clientKey prefixed with 'client_', chanID, subtopic",
			password:   "Client " + clientKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: true},
			err:        nil,
		},
		{
			desc:       "subscribe to channel with basic auth",
			username:   clientID,
			password:   clientKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.BasicAuth, clientID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: true},
			err:        nil,
		},
		{
			desc:       "subscribe to channel with basic auth and invalid credentials",
			username:   invalidID,
			password:   invalidKey,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   invalidID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.BasicAuth, invalidID, invalidKey),
			authNRes:   &grpcClientsV1.AuthnRes{Authenticated: false},
			authNErr:   svcerr.ErrAuthentication,
			err:        svcerr.ErrAuthorization,
		},
		{
			desc:       "subscribe to channel with b64 encoded credentials",
			password:   apiutil.BasicAuthPrefix + validEncodedCreds,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   clientID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.BasicAuth, clientID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			authZRes:   &grpcChannelsV1.AuthzRes{Authorized: true},
			err:        nil,
		},
		{
			desc:       "subscribe to channel with b64 encoded credentials and invalid credentials",
			password:   apiutil.BasicAuthPrefix + invalidEncodedCreds,
			chanID:     chanID,
			domainID:   domainID,
			clientID:   invalidID,
			subtopic:   subTopic,
			topicType:  messaging.MessageType,
			authNToken: smqauthn.AuthPack(smqauthn.BasicAuth, invalidID, invalidKey),
			authNRes:   &grpcClientsV1.AuthnRes{Authenticated: false},
			authNErr:   svcerr.ErrAuthentication,
			err:        svcerr.ErrAuthorization,
		},
		{
			desc:       "subscribe to health check topic with empty channel and valid clientKey",
			password:   clientKey,
			chanID:     "",
			domainID:   domainID,
			clientID:   clientID,
			topicType:  messaging.HealthType,
			authNToken: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, clientKey),
			authNRes:   &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			err:        nil,
		},
		{
			desc:      "subscribe to health check topic with empty channel and valid token",
			password:  token,
			chanID:    "",
			domainID:  domainID,
			clientID:  userID,
			topicType: messaging.HealthType,
			authNRes1: smqauthn.Session{UserID: userID},
			err:       nil,
		},
		{
			desc:      "subscribe to health check topic with empty domain and valid clientKey",
			password:  clientKey,
			chanID:    "",
			domainID:  "",
			clientID:  clientID,
			topicType: messaging.HealthType,
			authNRes:  &grpcClientsV1.AuthnRes{Id: clientID, Authenticated: true},
			err:       svcerr.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			subConfig := messaging.SubscriberConfig{
				ID:       sessionID,
				Topic:    "m." + tc.domainID + ".c." + tc.chanID + "." + subTopic,
				ClientID: tc.clientID,
				Handler:  c,
			}
			tc.clientType = policies.ClientType
			if strings.HasPrefix(tc.password, apiutil.BearerPrefix) {
				tc.clientType = policies.UserType
			}
			clientsCall := clients.On("Authenticate", mock.Anything, &grpcClientsV1.AuthnReq{Token: tc.authNToken}).Return(tc.authNRes, tc.authNErr)
			authCall := auth.On("Authenticate", mock.Anything, strings.TrimPrefix(tc.password, apiutil.BearerPrefix)).Return(tc.authNRes1, tc.authNErr)
			channelsCall := channels.On("Authorize", mock.Anything, &grpcChannelsV1.AuthzReq{
				ClientType: tc.clientType,
				ClientId:   tc.clientID,
				Type:       uint32(connections.Subscribe),
				ChannelId:  tc.chanID,
				DomainId:   tc.domainID,
			}).Return(tc.authZRes, tc.authZErr)
			repoCall := pubsub.On("Subscribe", mock.Anything, subConfig).Return(tc.subErr)
			err := svc.Subscribe(context.Background(), sessionID, tc.username, tc.password, tc.domainID, tc.chanID, tc.subtopic, tc.topicType, c)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			repoCall.Unset()
			clientsCall.Unset()
			authCall.Unset()
			channelsCall.Unset()
		})
	}
}

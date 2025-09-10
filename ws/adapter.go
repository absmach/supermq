// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"context"
	"strings"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	apiutil "github.com/absmach/supermq/api/http/util"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/policies"
)

var (
	// ErrFailedSubscription indicates that client couldn't subscribe to specified channel.
	ErrFailedSubscription = errors.New("failed to subscribe to a channel")
	// ErrFailedPublish indicates that client couldn't publish to specified channel.
	ErrFailedSubscribe = errors.New("failed to unsubscribe from topic")
	// ErrEmptyTopic indicate absence of clientKey in the request.
	ErrEmptyTopic = errors.New("empty topic")
)

// Service specifies web socket service API.
type Service interface {
	// Subscribe subscribes message from the broker using the clientKey for authorization,
	// the channelID for subscription and domainID specifies the domain for authorization.
	// Subtopic is optional.
	// If the subscription is successful, nil is returned otherwise error is returned.
	Subscribe(ctx context.Context, sessionID, username, password, domainID, chanID, subtopic string, client *Client) error

	Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	clients  grpcClientsV1.ClientsServiceClient
	channels grpcChannelsV1.ChannelsServiceClient
	authn    smqauthn.Authentication
	pubsub   messaging.PubSub
}

// New instantiates the WS adapter implementation.
func New(clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, authn smqauthn.Authentication, pubsub messaging.PubSub) Service {
	return &adapterService{
		clients:  clients,
		channels: channels,
		authn:    authn,
		pubsub:   pubsub,
	}
}

func (svc *adapterService) Subscribe(ctx context.Context, sessionID, username, password, domainID, channelID, subtopic string, c *Client) error {
	if channelID == "" || password == "" || domainID == "" {
		return svcerr.ErrAuthentication
	}

	clientID, err := svc.authorize(ctx, username, password, domainID, channelID, connections.Subscribe)
	if err != nil {
		return svcerr.ErrAuthorization
	}

	c.id = clientID

	subject := messaging.EncodeTopic(domainID, channelID, subtopic)
	subCfg := messaging.SubscriberConfig{
		ID:       sessionID,
		ClientID: clientID,
		Topic:    subject,
		Handler:  c,
	}
	if err := svc.pubsub.Subscribe(ctx, subCfg); err != nil {
		return errors.Wrap(ErrFailedSubscription, err)
	}

	return nil
}

func (svc *adapterService) Unsubscribe(ctx context.Context, sessionID, domainID, channelID, subtopic string) error {
	topic := messaging.EncodeTopic(domainID, channelID, subtopic)

	if err := svc.pubsub.Unsubscribe(ctx, sessionID, topic); err != nil {
		return errors.Wrap(ErrFailedSubscribe, err)
	}
	return nil
}

// authorize checks if the authKey is authorized to access the channel
// and returns the clientID or userID if it is.
func (svc *adapterService) authorize(ctx context.Context, username, password, domainID, chanID string, msgType connections.ConnType) (string, error) {
	var clientID, clientType string
	var err error
	switch {
	case strings.HasPrefix(password, apiutil.BearerPrefix):
		token := strings.TrimPrefix(password, apiutil.BearerPrefix)
		authnSession, err := svc.authn.Authenticate(ctx, token)
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		clientType = policies.UserType
		clientID = authnSession.UserID
	case username != "" && password != "":
		clientID, err = svc.clientAuthenticate(ctx, smqauthn.AuthPack(smqauthn.BasicAuth, username, password))
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		clientType = policies.ClientType
	case strings.HasPrefix(password, apiutil.BasicAuthPrefix):
		username, password, err := decodeAuth(strings.TrimPrefix(password, apiutil.BasicAuthPrefix))
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		clientID, err = svc.clientAuthenticate(ctx, smqauthn.AuthPack(smqauthn.BasicAuth, username, password))
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		clientType = policies.ClientType
	default:
		secret := strings.TrimPrefix(password, apiutil.ClientPrefix)
		clientID, err = svc.clientAuthenticate(ctx, smqauthn.AuthPack(smqauthn.DomainAuth, domainID, secret))
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		clientType = policies.ClientType
	}

	authzReq := &grpcChannelsV1.AuthzReq{
		ClientType: clientType,
		ClientId:   clientID,
		Type:       uint32(msgType),
		ChannelId:  chanID,
		DomainId:   domainID,
	}
	authzRes, err := svc.channels.Authorize(ctx, authzReq)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !authzRes.GetAuthorized() {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}

	return clientID, nil
}

func (svc *adapterService) clientAuthenticate(ctx context.Context, token string) (string, error) {
	authnRes, err := svc.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{Token: token})
	if err != nil {
		return "", svcerr.ErrAuthentication
	}
	if !authnRes.Authenticated {
		return "", svcerr.ErrAuthentication
	}

	return authnRes.GetId(), nil
}

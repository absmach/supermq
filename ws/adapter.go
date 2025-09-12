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
	Subscribe(ctx context.Context, sessionID, clientKey, domainID, chanID, subtopic string, topicType messaging.TopicType, client *Client) error

	Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string, topicType messaging.TopicType) error
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

func (svc *adapterService) Subscribe(ctx context.Context, sessionID, clientKey, domainID, channelID, subtopic string, topicType messaging.TopicType, c *Client) error {
	if (channelID == "" && topicType != messaging.HealthType) || clientKey == "" || domainID == "" {
		return svcerr.ErrAuthentication
	}

	clientID, err := svc.authorize(ctx, clientKey, domainID, channelID, connections.Subscribe, topicType)
	if err != nil {
		return svcerr.ErrAuthorization
	}

	c.id = clientID

	// Health check topics do not subscribe to the message broker.
	if topicType == messaging.HealthType {
		return nil
	}

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

func (svc *adapterService) Unsubscribe(ctx context.Context, sessionID, domainID, channelID, subtopic string, topicType messaging.TopicType) error {
	topic := messaging.EncodeTopic(domainID, channelID, subtopic)

	// Health check topics do not subscribe to the message broker.
	if topicType == messaging.MessageType {
		if err := svc.pubsub.Unsubscribe(ctx, sessionID, topic); err != nil {
			return errors.Wrap(ErrFailedSubscribe, err)
		}
	}
	return nil
}

// authorize checks if the authKey is authorized to access the channel
// and returns the clientID or userID if it is.
func (svc *adapterService) authorize(ctx context.Context, authKey, domainID, chanID string, msgType connections.ConnType, topicType messaging.TopicType) (string, error) {
	var clientID, clientType string
	switch {
	case strings.HasPrefix(authKey, apiutil.BearerPrefix):
		token := strings.TrimPrefix(authKey, apiutil.BearerPrefix)
		authnSession, err := svc.authn.Authenticate(ctx, token)
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		clientType = policies.UserType
		clientID = authnSession.UserID
	default:
		secret := strings.TrimPrefix(authKey, apiutil.ClientPrefix)
		authnRes, err := svc.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{Token: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, secret)})
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		if !authnRes.Authenticated {
			return "", svcerr.ErrAuthentication
		}
		clientType = policies.ClientType
		clientID = authnRes.GetId()
	}

	// Health check topics do not require channel authorization.
	if topicType == messaging.HealthType {
		return clientID, nil
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

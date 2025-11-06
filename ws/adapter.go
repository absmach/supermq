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
	Subscribe(ctx context.Context, sessionID, username, password, domainID, chanID, subtopic string, topicType messaging.TopicType, client *Client) error

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

func (svc *adapterService) Subscribe(ctx context.Context, sessionID, username, password, domainID, channelID, subtopic string, topicType messaging.TopicType, c *Client) error {
	if (channelID == "" && topicType != messaging.HealthType) || password == "" || domainID == "" {
		return svcerr.ErrAuthentication
	}

	clientID, err := svc.authorize(ctx, username, password, domainID, channelID, connections.Subscribe, topicType)
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
func (svc *adapterService) authorize(ctx context.Context, username, password, domainID, chanID string, msgType connections.ConnType, topicType messaging.TopicType) (string, error) {
	var token, clientType string
	var err error
	switch {
	case strings.HasPrefix(password, apiutil.BearerPrefix):
		token = strings.TrimPrefix(password, apiutil.BearerPrefix)
		clientType = policies.UserType
	case username != "" && password != "":
		token = smqauthn.AuthPack(smqauthn.BasicAuth, username, password)
		clientType = policies.ClientType
	case strings.HasPrefix(password, apiutil.BasicAuthPrefix):
		username, password, err := decodeAuth(strings.TrimPrefix(password, apiutil.BasicAuthPrefix))
		if err != nil {
			return "", errors.Wrap(svcerr.ErrAuthentication, err)
		}
		token = smqauthn.AuthPack(smqauthn.BasicAuth, username, password)
		clientType = policies.ClientType
	default:
		token = smqauthn.AuthPack(smqauthn.DomainAuth, domainID, strings.TrimPrefix(password, apiutil.ClientPrefix))
		clientType = policies.ClientType
	}

	id, err := svc.authenticate(ctx, clientType, token)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthentication, err)
	}

	// Health check topics do not require channel authorization.
	if topicType == messaging.HealthType {
		return id, nil
	}

	authzReq := &grpcChannelsV1.AuthzReq{
		ClientType: clientType,
		ClientId:   id,
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

	return id, nil
}

func (svc *adapterService) authenticate(ctx context.Context, authType, token string) (string, error) {
	switch authType {
	case policies.UserType:
		authnSession, err := svc.authn.Authenticate(ctx, token)
		if err != nil {
			return "", err
		}
		return authnSession.UserID, nil
	case policies.ClientType:
		authnRes, err := svc.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{Token: token})
		if err != nil {
			return "", err
		}
		if !authnRes.Authenticated {
			return "", svcerr.ErrAuthentication
		}

		return authnRes.GetId(), nil
	default:
		return "", errInvalidClientType
	}
}

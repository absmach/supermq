// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"context"
	"fmt"
	"strings"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
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
	Subscribe(ctx context.Context, sessionID, clientUsername, clientPassword, domainID, chanID, subtopic string, client *Client) error

	Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	clients  grpcClientsV1.ClientsServiceClient
	channels grpcChannelsV1.ChannelsServiceClient
	pubsub   messaging.PubSub
}

// New instantiates the WS adapter implementation.
func New(clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, pubsub messaging.PubSub) Service {
	return &adapterService{
		clients:  clients,
		channels: channels,
		pubsub:   pubsub,
	}
}

func (svc *adapterService) Subscribe(ctx context.Context, sessionID, clientUsername, clientPassword, domainID, channelID, subtopic string, c *Client) error {
	if channelID == "" || clientPassword == "" || domainID == "" {
		return svcerr.ErrAuthentication
	}

	clientID, err := svc.authorize(ctx, clientUsername, clientPassword, channelID, domainID, connections.Subscribe)
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

// authorize checks if the clientKey is authorized to access the channel
// and returns the clientID if it is.
func (svc *adapterService) authorize(ctx context.Context, clientUsername, clientPassword, channelID, domainID string, msgType connections.ConnType) (string, error) {
	var clientID, clientType string
	var err error
	switch {
	case strings.HasPrefix(clientPassword, "Client"):
		secret := extractClientSecret(clientPassword)
		fmt.Println("secret", secret)
		clientID, err = svc.clientAuthenticate(ctx, smqauthn.AuthPack(smqauthn.DomainAuth, domainID, secret))
		if err != nil {
			return "", err
		}
		clientType = policies.ClientType
	case clientUsername != "" && clientPassword != "":
		clientID, err = svc.clientAuthenticate(ctx, smqauthn.AuthPack(smqauthn.BasicAuth, clientUsername, clientPassword))
		if err != nil {
			return "", err
		}
		clientType = policies.ClientType
	case clientUsername == "" && clientPassword != "":
		cid, secret, err := decodeAuth(clientPassword)
		if err != nil {
			return "", err
		}
		clientID, err = svc.clientAuthenticate(ctx, smqauthn.AuthPack(smqauthn.BasicAuth, cid, secret))
		if err != nil {
			return "", err
		}
		clientType = policies.ClientType
	default:
		return "", svcerr.ErrAuthentication
	}

	ar := &grpcChannelsV1.AuthzReq{
		DomainId:   domainID,
		ClientId:   clientID,
		ClientType: clientType,
		ChannelId:  channelID,
		Type:       uint32(msgType),
	}
	res, err := svc.channels.Authorize(ctx, ar)
	if err != nil {
		return "", err
	}
	if !res.GetAuthorized() {
		return "", svcerr.ErrAuthorization
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

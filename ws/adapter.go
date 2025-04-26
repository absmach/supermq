// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"context"
	"fmt"
	"strings"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/policies"
)

const chansPrefix = "channels"

var (
	// ErrFailedSubscription indicates that client couldn't subscribe to specified channel.
	ErrFailedSubscription = errors.New("failed to subscribe to a channel")

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
	Subscribe(ctx context.Context, sessionID, clientKey, domainID, chanID, subtopic string, client *Client) error

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

func (svc *adapterService) Subscribe(ctx context.Context, sessionID, clientKey, domainID, chanID, subtopic string, c *Client) error {
	if chanID == "" || clientKey == "" || domainID == "" {
		return svcerr.ErrAuthentication
	}

	clientID, err := svc.authorize(ctx, clientKey, domainID, chanID, connections.Subscribe)
	if err != nil {
		return svcerr.ErrAuthorization
	}

	subject := fmt.Sprintf("%s.%s", chansPrefix, chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}

	subCfg := messaging.SubscriberConfig{
		ID:       sessionID,
		ClientID: clientID,
		Topic:    subject,
		Handler:  c,
	}
	if err := svc.pubsub.Subscribe(ctx, subCfg); err != nil {
		return ErrFailedSubscription
	}

	return nil
}

func (svc *adapterService) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string) error {
	topic := fmt.Sprintf("%s.%s", chansPrefix, chanID)
	if subtopic != "" {
		topic = fmt.Sprintf("%s.%s", topic, subtopic)
	}

	if err := svc.pubsub.Unsubscribe(ctx, sessionID, topic); err != nil {
		return errors.Wrap(ErrFailedSubscribe, err)
	}
	return nil
}

// authorize checks if the clientKey is authorized to access the channel
// and returns the clientID if it is.
func (svc *adapterService) authorize(ctx context.Context, clientKey, domainID, chanID string, msgType connections.ConnType) (string, error) {
	authnReq := &grpcClientsV1.AuthnReq{
		ClientSecret: clientKey,
	}
	if strings.HasPrefix(clientKey, "Client") {
		authnReq.ClientSecret = extractClientSecret(clientKey)
	}
	authnRes, err := svc.clients.Authenticate(ctx, authnReq)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if !authnRes.GetAuthenticated() {
		return "", errors.Wrap(svcerr.ErrAuthentication, err)
	}

	authzReq := &grpcChannelsV1.AuthzReq{
		ClientType: policies.ClientType,
		ClientId:   authnRes.GetId(),
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

	return authnRes.GetId(), nil
}

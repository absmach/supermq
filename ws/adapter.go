// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"context"
	"fmt"
	"strings"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	api "github.com/absmach/supermq/api/http"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/policies"
)

const chansPrefix = "channels"

var (
	// errFailedMessagePublish indicates that message publishing failed.
	errFailedMessagePublish = errors.New("failed to publish message")

	// ErrFailedSubscription indicates that client couldn't subscribe to specified channel.
	ErrFailedSubscription = errors.New("failed to subscribe to a channel")

	// errFailedUnsubscribe indicates that client couldn't unsubscribe from specified channel.
	errFailedUnsubscribe = errors.New("failed to unsubscribe from a channel")

	// ErrEmptyTopic indicate absence of clientKey in the request.
	ErrEmptyTopic = errors.New("empty topic")
)

// Service specifies web socket service API.
type Service interface {
	// Subscribe subscribes message from the broker using the clientKey for authorization,
	// the channelID for subscription and domainID specifies the domain for authorization.
	// Subtopic is optional.
	// If the subscription is successful, nil is returned otherwise error is returned.
	Subscribe(ctx context.Context, clientKey, domainID, chanID, subtopic string, client *Client) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	clients  grpcClientsV1.ClientsServiceClient
	channels grpcChannelsV1.ChannelsServiceClient
	domains  grpcDomainsV1.DomainsServiceClient
	pubsub   messaging.PubSub
}

// New instantiates the WS adapter implementation.
func New(clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, domains grpcDomainsV1.DomainsServiceClient, pubsub messaging.PubSub) Service {
	return &adapterService{
		clients:  clients,
		channels: channels,
		domains:  domains,
		pubsub:   pubsub,
	}
}

func (svc *adapterService) Subscribe(ctx context.Context, clientKey, domainID, chanID, subtopic string, c *Client) error {
	if chanID == "" || clientKey == "" || domainID == "" {
		return svcerr.ErrAuthentication
	}

	domainID, err := svc.resolveDomain(domainID)
	if err != nil {
		return err
	}
	chanID, err = svc.resolveChannel(chanID, domainID)
	if err != nil {
		return err
	}

	clientID, err := svc.authorize(ctx, clientKey, domainID, chanID, connections.Subscribe)
	if err != nil {
		return svcerr.ErrAuthorization
	}

	c.id = clientID

	subject := fmt.Sprintf("%s.%s", chansPrefix, chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}

	subCfg := messaging.SubscriberConfig{
		ID:       clientID,
		ClientID: clientID,
		Topic:    subject,
		Handler:  c,
	}
	if err := svc.pubsub.Subscribe(ctx, subCfg); err != nil {
		return ErrFailedSubscription
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

func (svc *adapterService) resolveDomain(domain string) (string, error) {
	if api.ValidateUUID(domain) == nil {
		return domain, nil
	}

	d, err := svc.domains.RetrieveByRoute(context.Background(), &grpcCommonV1.RetrieveByRouteReq{
		Route: domain,
	})
	if err != nil {
		return "", err
	}

	return d.Entity.Id, nil
}

func (svc *adapterService) resolveChannel(channel, domainID string) (string, error) {
	if api.ValidateUUID(channel) == nil {
		return channel, nil
	}

	c, err := svc.channels.RetrieveByRoute(context.Background(), &grpcCommonV1.RetrieveByRouteReq{
		Route:    channel,
		DomainId: domainID,
	})
	if err != nil {
		return "", err
	}

	return c.Entity.Id, nil
}

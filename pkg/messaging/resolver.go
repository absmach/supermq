// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/gofrs/uuid/v5"
)

var (
	ErrEmptyRouteID         = errors.New("empty route or id")
	ErrFailedResolveDomain  = errors.New("failed to resolve domain route")
	ErrFailedResolveChannel = errors.New("failed to resolve channel route")
)

// TopicResolver contains definitions for resolving domain and channel IDs
// from their respective routes from the message topic.
type TopicResolver interface {
	Resolve(ctx context.Context, domain, channel string) (domainID string, channelID string, err error)
	ResolveTopic(ctx context.Context, topic string) (rtopic string, err error)
}

type resolver struct {
	channels grpcChannelsV1.ChannelsServiceClient
	domains  grpcDomainsV1.DomainsServiceClient
}

// NewTopicResolver creates a new instance of TopicResolver.
func NewTopicResolver(channelsClient grpcChannelsV1.ChannelsServiceClient, domainsClient grpcDomainsV1.DomainsServiceClient) TopicResolver {
	return &resolver{
		channels: channelsClient,
		domains:  domainsClient,
	}
}

func (r *resolver) Resolve(ctx context.Context, domain, channel string) (string, string, error) {
	if domain == "" || channel == "" {
		return "", "", ErrEmptyRouteID
	}

	domainID, err := r.resolveDomain(ctx, domain)
	if err != nil {
		return "", "", errors.Wrap(ErrFailedResolveDomain, err)
	}
	channelID, err := r.resolveChannel(ctx, channel, domainID)
	if err != nil {
		return "", "", errors.Wrap(ErrFailedResolveChannel, err)
	}

	return domainID, channelID, nil
}

func (r *resolver) ResolveTopic(ctx context.Context, topic string) (string, error) {
	domain, channel, subtopic, err := ParseTopic(topic)
	if err != nil {
		return "", errors.Wrap(ErrMalformedTopic, err)
	}

	domainID, channelID, err := r.Resolve(ctx, domain, channel)
	if err != nil {
		return "", err
	}
	rtopic := EncodeAdapterTopic(domainID, channelID, subtopic)

	return rtopic, nil
}

func (r *resolver) resolveDomain(ctx context.Context, domain string) (string, error) {
	if validateUUID(domain) == nil {
		return domain, nil
	}
	d, err := r.domains.RetrieveByRoute(ctx, &grpcCommonV1.RetrieveByRouteReq{
		Route: domain,
	})
	if err != nil {
		return "", err
	}

	return d.Entity.Id, nil
}

func (r *resolver) resolveChannel(ctx context.Context, channel, domainID string) (string, error) {
	if validateUUID(channel) == nil {
		return channel, nil
	}
	c, err := r.channels.RetrieveByRoute(ctx, &grpcCommonV1.RetrieveByRouteReq{
		Route:    channel,
		DomainId: domainID,
	})
	if err != nil {
		return "", err
	}

	return c.Entity.Id, nil
}

func validateUUID(extID string) (err error) {
	id, err := uuid.FromString(extID)
	if id.String() != extID || err != nil {
		return err
	}

	return nil
}

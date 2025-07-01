// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging_test

import (
	"context"
	"fmt"
	"testing"

	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	chmocks "github.com/absmach/supermq/channels/mocks"
	dmocks "github.com/absmach/supermq/domains/mocks"
	"github.com/absmach/supermq/internal/testsutil"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	validRoute   = "valid-route"
	invalidRoute = "invalid-route"
	channelID    = testsutil.GenerateUUID(&testing.T{})
	domainID     = testsutil.GenerateUUID(&testing.T{})
	topicFmt     = "m/%s/c/%s"
)

func setupResolver() (messaging.TopicResolver, *dmocks.DomainsServiceClient, *chmocks.ChannelsServiceClient) {
	channels := new(chmocks.ChannelsServiceClient)
	domains := new(dmocks.DomainsServiceClient)
	resolver := messaging.NewTopicResolver(channels, domains)

	return resolver, domains, channels
}

func TestResolve(t *testing.T) {
	resolver, domains, channels := setupResolver()

	cases := []struct {
		desc        string
		domain      string
		channel     string
		domainID    string
		channelID   string
		domainsErr  error
		channelsErr error
		err         error
	}{
		{
			desc:      "valid domainID and channelID",
			domain:    domainID,
			channel:   channelID,
			domainID:  domainID,
			channelID: channelID,
			err:       nil,
		},
		{
			desc:      "valid domain route and channel ID",
			domain:    validRoute,
			channel:   channelID,
			domainID:  domainID,
			channelID: channelID,
			err:       nil,
		},
		{
			desc:      "valid domain ID and channel route",
			domain:    domainID,
			channel:   validRoute,
			domainID:  domainID,
			channelID: channelID,
			err:       nil,
		},
		{
			desc:      "valid domain route and channel route",
			domain:    validRoute,
			channel:   validRoute,
			domainID:  domainID,
			channelID: channelID,
			err:       nil,
		},
		{
			desc:       "invalid domain route  and valid channel",
			domain:     invalidRoute,
			channel:    channelID,
			domainID:   "",
			channelID:  "",
			domainsErr: svcerr.ErrNotFound,
			err:        messaging.ErrFailedResolveDomain,
		},
		{
			desc:        "valid domain and invalid channel",
			domain:      domainID,
			channel:     invalidRoute,
			domainID:    domainID,
			channelID:   "",
			channelsErr: svcerr.ErrNotFound,
			err:         messaging.ErrFailedResolveChannel,
		},
		{
			desc:      "empty domain",
			domain:    "",
			channel:   channelID,
			domainID:  "",
			channelID: "",
			err:       messaging.ErrEmptyRouteID,
		},
		{
			desc:      "empty channel",
			domain:    domainID,
			channel:   "",
			domainID:  domainID,
			channelID: "",
			err:       messaging.ErrEmptyRouteID,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			domainsCall := domains.On("RetrieveByRoute", mock.Anything, &grpcCommonV1.RetrieveByRouteReq{Route: tc.domain}).Return(&grpcCommonV1.RetrieveEntityRes{
				Entity: &grpcCommonV1.EntityBasic{
					Id: tc.domainID,
				},
			}, tc.domainsErr)
			channelsCall := channels.On("RetrieveByRoute", mock.Anything, &grpcCommonV1.RetrieveByRouteReq{Route: tc.channel, DomainId: tc.domainID}).Return(&grpcCommonV1.RetrieveEntityRes{
				Entity: &grpcCommonV1.EntityBasic{
					Id: tc.channelID,
				},
			}, tc.channelsErr)
			domainID, channelID, err := resolver.Resolve(context.Background(), tc.domain, tc.channel)
			assert.True(t, errors.Contains(err, tc.err), "expected error %v, got %v", tc.err, err)
			if err == nil {
				assert.Equal(t, tc.domainID, domainID, "expected domain ID %s, got %s", tc.domainID, domainID)
				assert.Equal(t, tc.channelID, channelID, "expected channel ID %s, got %s", tc.channelID, channelID)
			}
			domainsCall.Unset()
			channelsCall.Unset()
		})
	}
}

func TestResolveTopic(t *testing.T) {
	resolver, domains, channels := setupResolver()

	cases := []struct {
		desc        string
		topic       string
		domain      string
		channel     string
		domainID    string
		channelID   string
		domainsErr  error
		channelsErr error
		response    string
		err         error
	}{
		{
			desc:      "valid topic with domainID and channelID",
			topic:     fmt.Sprintf(topicFmt, domainID, channelID),
			domain:    domainID,
			channel:   channelID,
			domainID:  domainID,
			channelID: channelID,
			response:  fmt.Sprintf(topicFmt, domainID, channelID),
			err:       nil,
		},
		{
			desc:      "valid topic with domain route and channel ID",
			topic:     fmt.Sprintf(topicFmt, validRoute, channelID),
			domain:    validRoute,
			channel:   channelID,
			domainID:  domainID,
			channelID: channelID,
			response:  fmt.Sprintf(topicFmt, domainID, channelID),
			err:       nil,
		},
		{
			desc:      "valid topic with domain ID and channel route",
			topic:     fmt.Sprintf(topicFmt, domainID, validRoute),
			domain:    domainID,
			channel:   validRoute,
			domainID:  domainID,
			channelID: channelID,
			response:  fmt.Sprintf(topicFmt, domainID, channelID),
			err:       nil,
		},
		{
			desc:      "valid topic with domain route and channel route",
			topic:     fmt.Sprintf(topicFmt, validRoute, validRoute),
			domain:    validRoute,
			channel:   validRoute,
			domainID:  domainID,
			channelID: channelID,
			response:  fmt.Sprintf(topicFmt, domainID, channelID),
			err:       nil,
		},
		{
			desc:       "invalid topic with invalid domain route and valid channel",
			topic:      fmt.Sprintf(topicFmt, invalidRoute, channelID),
			domain:     invalidRoute,
			channel:    channelID,
			domainID:   "",
			channelID:  "",
			domainsErr: svcerr.ErrNotFound,
			err:        messaging.ErrFailedResolveDomain,
		},
		{
			desc:      "valid topic with valid topic with domainID and channelID and subtopic",
			topic:     fmt.Sprintf(topicFmt, domainID, channelID) + "/subtopic",
			domain:    domainID,
			channel:   channelID,
			domainID:  domainID,
			channelID: channelID,
			response:  fmt.Sprintf(topicFmt, domainID, channelID) + "/subtopic",
			err:       nil,
		},
		{
			desc:      "invalid topic with empty domain",
			topic:     fmt.Sprintf(topicFmt, "", channelID),
			domain:    "",
			channel:   channelID,
			domainID:  "",
			channelID: "",
			err:       messaging.ErrMalformedTopic,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			domainsCall := domains.On("RetrieveByRoute", mock.Anything, &grpcCommonV1.RetrieveByRouteReq{Route: tc.domain}).Return(&grpcCommonV1.RetrieveEntityRes{
				Entity: &grpcCommonV1.EntityBasic{
					Id: tc.domainID,
				},
			}, tc.domainsErr)
			channelsCall := channels.On("RetrieveByRoute", mock.Anything, &grpcCommonV1.RetrieveByRouteReq{Route: tc.channel, DomainId: tc.domainID}).Return(&grpcCommonV1.RetrieveEntityRes{
				Entity: &grpcCommonV1.EntityBasic{
					Id: tc.channelID,
				},
			}, tc.channelsErr)
			rtopic, err := resolver.ResolveTopic(context.Background(), tc.topic)
			assert.True(t, errors.Contains(err, tc.err), "expected error %v, got %v", tc.err, err)
			if err == nil {
				assert.Equal(t, tc.response, rtopic, "expected topic %s, got %s", tc.response, rtopic)
			}
			domainsCall.Unset()
			channelsCall.Unset()
		})
	}
}

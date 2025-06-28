// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging_test

import (
	"context"
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

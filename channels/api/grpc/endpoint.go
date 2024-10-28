// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	channels "github.com/absmach/magistrala/channels/private"
	"github.com/go-kit/kit/endpoint"
)

func removeThingConnectionsEndpoint(svc channels.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeThingConnectionsReq)

		if err := svc.RemoveThingConnections(ctx, req.thingID); err != nil {
			return removeThingConnectionsRes{}, err
		}

		return removeThingConnectionsRes{}, nil
	}
}

func unsetParentGroupFromChannelsEndpoint(svc channels.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(unsetParentGroupFromChannelsReq)

		if err := svc.UnsetParentGroupFromChannels(ctx, req.parentGroupID); err != nil {
			return unsetParentGroupFromChannelsRes{}, err
		}

		return unsetParentGroupFromChannelsRes{}, nil
	}
}

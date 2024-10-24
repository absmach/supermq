// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/absmach/magistrala/channels"
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

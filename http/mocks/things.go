// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ policies.ThingsServiceClient = (*thingsClient)(nil)

// ServiceErrToken is used to simulate internal server error.
const ServiceErrToken = "unavailable"

type thingsClient struct {
	things map[string]string
}

// NewThingsClient returns mock implementation of things service client.
func NewThingsClient(data map[string]string) policies.ThingsServiceClient {
	return &thingsClient{data}
}

func (tc thingsClient) AuthorizeByKey(ctx context.Context, req *policies.TAuthorizeReq, opts ...grpc.CallOption) (*policies.ThingID, error) {
	key := req.GetSub()

	// Since there is no appropriate way to simulate internal server error,
	// we had to use this obscure approach. ErrorToken simulates gRPC
	// call which returns internal server error.
	if key == ServiceErrToken {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	if key == "" {
		return nil, errors.ErrAuthentication
	}

	id, ok := tc.things[key]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials provided")
	}

	return &policies.ThingID{Value: id}, nil
}

func (tc thingsClient) Authorize(context.Context, *policies.TAuthorizeReq, ...grpc.CallOption) (*policies.TAuthorizeRes, error) {
	panic("not implemented")
}

func (tc thingsClient) Identify(ctx context.Context, req *policies.Key, opts ...grpc.CallOption) (*policies.ThingID, error) {
	panic("not implemented")
}

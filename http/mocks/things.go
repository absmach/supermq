// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ mainflux.ThingsServiceClient = (*thingsClient)(nil)

type thingsClient struct {
	things map[string]string
}

// NewThingsClient returns mock implementation of things service client.
func NewThingsClient(data map[string]string) mainflux.ThingsServiceClient {
	return &thingsClient{data}
}

func (tc thingsClient) CanAccessByKey(ctx context.Context, req *mainflux.AccessByKeyReq, opts ...grpc.CallOption) (*mainflux.ThingID, error) {
	key := req.GetToken()

	if key == "" {
		return nil, errors.ErrAuthentication
	}

	id, ok := tc.things[key]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials provided")
	}

	return &mainflux.ThingID{Value: id}, nil
}

func (tc thingsClient) CanAccessByID(context.Context, *mainflux.AccessByIDReq, ...grpc.CallOption) (*empty.Empty, error) {
	panic("not implemented")
}

func (tc thingsClient) IsChannelOwner(context.Context, *mainflux.ChannelOwnerReq, ...grpc.CallOption) (*empty.Empty, error) {
	panic("not implemented")
}

func (tc thingsClient) Identify(ctx context.Context, req *mainflux.Token, opts ...grpc.CallOption) (*mainflux.ThingID, error) {
	panic("not implemented")
}

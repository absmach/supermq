// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
	"google.golang.org/grpc"
)

var _ policies.ThingsServiceClient = (*thingsServiceMock)(nil)

type thingsServiceMock struct {
	channels map[string]string
}

// NewThingsService returns mock implementation of things service
func NewThingsService(channels map[string]string) policies.ThingsServiceClient {
	return &thingsServiceMock{channels}
}

func (svc thingsServiceMock) CanAccessByKey(ctx context.Context, in *policies.AccessByKeyReq, opts ...grpc.CallOption) (*policies.ThingID, error) {
	token := in.GetToken()
	if token == "invalid" {
		return nil, errors.ErrAuthentication
	}

	if token == "" {
		return nil, errors.ErrAuthentication
	}

	if token == "token" {
		return nil, errors.ErrAuthorization
	}

	return &policies.ThingID{Value: token}, nil
}

func (svc thingsServiceMock) CanAccessByID(context.Context, *policies.AccessByIDReq, ...grpc.CallOption) (*empty.Empty, error) {
	panic("not implemented")
}

func (svc thingsServiceMock) IsChannelOwner(ctx context.Context, in *policies.ChannelOwnerReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	if id, ok := svc.channels[in.GetOwner()]; ok {
		if id == in.ChanID {
			return nil, nil
		}
	}
	return nil, errors.ErrAuthorization
}

func (svc thingsServiceMock) Identify(context.Context, *policies.Token, ...grpc.CallOption) (*policies.ThingID, error) {
	panic("not implemented")
}

// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

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

func (svc thingsServiceMock) AuthorizeByKey(ctx context.Context, in *policies.TAuthorizeReq, opts ...grpc.CallOption) (*policies.ThingID, error) {
	token := in.GetSub()
	if token == "invalid" || token == "" {
		return nil, errors.ErrAuthentication
	}

	return &policies.ThingID{Value: token}, nil
}

func (svc thingsServiceMock) Authorize(context.Context, *policies.TAuthorizeReq, ...grpc.CallOption) (*policies.TAuthorizeRes, error) {
	panic("not implemented")
}

func (svc thingsServiceMock) Identify(context.Context, *policies.Key, ...grpc.CallOption) (*policies.ThingID, error) {
	panic("not implemented")
}

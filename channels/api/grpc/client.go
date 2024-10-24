// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"fmt"
	"time"

	grpcChannelsV1 "github.com/absmach/magistrala/internal/grpc/channels/v1"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const svcName = "channels.v1.ChannelsService"

var _ grpcChannelsV1.ChannelsServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	timeout                time.Duration
	removeThingConnections endpoint.Endpoint
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn, timeout time.Duration) grpcChannelsV1.ChannelsServiceClient {
	return &grpcClient{
		removeThingConnections: kitgrpc.NewClient(
			conn,
			svcName,
			"RemoveThingConnections",
			encodeRemoveThingConnectionsRequest,
			decodeRemoveThingConnectionsResponse,
			grpcChannelsV1.RemoveThingConnectionsRes{},
		).Endpoint(),
		timeout: timeout,
	}
}

func (client grpcClient) RemoveThingConnections(ctx context.Context, req *grpcChannelsV1.RemoveThingConnectionsReq, _ ...grpc.CallOption) (r *grpcChannelsV1.RemoveThingConnectionsRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	if _, err := client.removeThingConnections(ctx, req); err != nil {
		return &grpcChannelsV1.RemoveThingConnectionsRes{}, decodeError(err)
	}

	return &grpcChannelsV1.RemoveThingConnectionsRes{}, nil
}

func encodeRemoveThingConnectionsRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	return grpcReq.(*grpcChannelsV1.RemoveThingConnectionsReq), nil
}

func decodeRemoveThingConnectionsResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	return grpcRes.(*grpcChannelsV1.RemoveThingConnectionsRes), nil
}
func decodeError(err error) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.Unauthenticated:
			return errors.Wrap(svcerr.ErrAuthentication, errors.New(st.Message()))
		case codes.PermissionDenied:
			return errors.Wrap(svcerr.ErrAuthorization, errors.New(st.Message()))
		case codes.InvalidArgument:
			return errors.Wrap(errors.ErrMalformedEntity, errors.New(st.Message()))
		case codes.FailedPrecondition:
			return errors.Wrap(errors.ErrMalformedEntity, errors.New(st.Message()))
		case codes.NotFound:
			return errors.Wrap(svcerr.ErrNotFound, errors.New(st.Message()))
		case codes.AlreadyExists:
			return errors.Wrap(svcerr.ErrConflict, errors.New(st.Message()))
		case codes.OK:
			if msg := st.Message(); msg != "" {
				return errors.Wrap(errors.ErrUnidentified, errors.New(msg))
			}
			return nil
		default:
			return errors.Wrap(fmt.Errorf("unexpected gRPC status: %s (status code:%v)", st.Code().String(), st.Code()), errors.New(st.Message()))
		}
	}
	return err
}
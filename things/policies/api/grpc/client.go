package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mainflux/mainflux/things/policies"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
	"google.golang.org/grpc"
)

const svcName = "policies.AuthService"

var _ policies.ThingsServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	canAccessByKey endpoint.Endpoint
	canAccessByID  endpoint.Endpoint
	isChannelOwner endpoint.Endpoint
	identify       endpoint.Endpoint
	timeout        time.Duration
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn, timeout time.Duration) policies.ThingsServiceClient {
	return &grpcClient{
		canAccessByKey: otelkit.EndpointMiddleware(otelkit.WithOperation("can_access"))(kitgrpc.NewClient(
			conn,
			svcName,
			"CanAccessByKey",
			encodeCanAccessByKeyRequest,
			decodeIdentityResponse,
			policies.ThingID{},
		).Endpoint()),
		canAccessByID: otelkit.EndpointMiddleware(otelkit.WithOperation("can_access_by_id"))(kitgrpc.NewClient(
			conn,
			svcName,
			"CanAccessByID",
			encodeCanAccessByIDRequest,
			decodeEmptyResponse,
			empty.Empty{},
		).Endpoint()),
		isChannelOwner: otelkit.EndpointMiddleware(otelkit.WithOperation("is_channel_owner"))(kitgrpc.NewClient(
			conn,
			svcName,
			"IsChannelOwner",
			encodeIsChannelOwner,
			decodeEmptyResponse,
			empty.Empty{},
		).Endpoint()),
		identify: otelkit.EndpointMiddleware(otelkit.WithOperation("identify"))(kitgrpc.NewClient(
			conn,
			svcName,
			"Identify",
			encodeIdentifyRequest,
			decodeIdentityResponse,
			policies.ThingID{},
		).Endpoint()),

		timeout: timeout,
	}
}

func (client grpcClient) CanAccessByKey(ctx context.Context, req *policies.AccessByKeyReq, _ ...grpc.CallOption) (*policies.ThingID, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	ar := accessByKeyReq{
		thingKey: req.GetToken(),
		chanID:   req.GetChanID(),
	}
	res, err := client.canAccessByKey(ctx, ar)
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &policies.ThingID{Value: ir.id}, nil
}

func (client grpcClient) CanAccessByID(ctx context.Context, req *policies.AccessByIDReq, _ ...grpc.CallOption) (*empty.Empty, error) {
	ar := accessByIDReq{thingID: req.GetThingID(), chanID: req.GetChanID()}
	res, err := client.canAccessByID(ctx, ar)
	if err != nil {
		return nil, err
	}

	er := res.(emptyRes)
	return &empty.Empty{}, er.err
}

func (client grpcClient) IsChannelOwner(ctx context.Context, req *policies.ChannelOwnerReq, _ ...grpc.CallOption) (*empty.Empty, error) {
	ar := channelOwnerReq{owner: req.GetOwner(), chanID: req.GetChanID()}
	res, err := client.isChannelOwner(ctx, ar)
	if err != nil {
		return nil, err
	}

	er := res.(emptyRes)
	return &empty.Empty{}, er.err
}

func (client grpcClient) Identify(ctx context.Context, req *policies.Key, _ ...grpc.CallOption) (*policies.ThingID, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.identify(ctx, identifyReq{key: req.GetValue()})
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &policies.ThingID{Value: ir.id}, nil
}

func encodeCanAccessByKeyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(accessByKeyReq)
	return &policies.AccessByKeyReq{Token: req.thingKey, ChanID: req.chanID}, nil
}

func encodeCanAccessByIDRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(accessByIDReq)
	return &policies.AccessByIDReq{ThingID: req.thingID, ChanID: req.chanID}, nil
}

func encodeIsChannelOwner(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(channelOwnerReq)
	return &policies.ChannelOwnerReq{Owner: req.owner, ChanID: req.chanID}, nil
}

func encodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(identifyReq)
	return &policies.Key{Value: req.key}, nil
}

func decodeIdentityResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*policies.ThingID)
	return identityRes{id: res.GetValue()}, nil
}

func decodeEmptyResponse(_ context.Context, _ interface{}) (interface{}, error) {
	return emptyRes{}, nil
}

package grpc

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/groups"
	"github.com/mainflux/mainflux/things/policies"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ policies.ThingsServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	canAccessByKey kitgrpc.Handler
	canAccessByID  kitgrpc.Handler
	isChannelOwner kitgrpc.Handler
	identify       kitgrpc.Handler
	policies.UnimplementedThingsServiceServer
}

// NewServer returns new ThingsServiceServer instance.
func NewServer(csvc clients.Service, gsvc groups.Service, psvc policies.Service) policies.ThingsServiceServer {
	return &grpcServer{
		canAccessByKey: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("can_access_by_key"))(canAccessEndpoint(psvc)),
			decodeCanAccessByKeyRequest,
			encodeIdentityResponse,
		),
		canAccessByID: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("can_access_by_id"))(canAccessByIDEndpoint(psvc)),
			decodeCanAccessByIDRequest,
			encodeEmptyResponse,
		),
		isChannelOwner: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("is_channel_owner"))(isChannelOwnerEndpoint(gsvc)),
			decodeIsChannelOwnerRequest,
			encodeEmptyResponse,
		),
		identify: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("identify"))(identifyEndpoint(csvc)),
			decodeIdentifyRequest,
			encodeIdentityResponse,
		),
	}
}

func (gs *grpcServer) CanAccessByKey(ctx context.Context, req *policies.AccessByKeyReq) (*policies.ThingID, error) {
	_, res, err := gs.canAccessByKey.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}

	return res.(*policies.ThingID), nil
}

func (gs *grpcServer) CanAccessByID(ctx context.Context, req *policies.AccessByIDReq) (*empty.Empty, error) {
	_, res, err := gs.canAccessByID.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}

	return res.(*empty.Empty), nil
}

func (gs *grpcServer) IsChannelOwner(ctx context.Context, req *policies.ChannelOwnerReq) (*empty.Empty, error) {
	_, res, err := gs.isChannelOwner.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}

	return res.(*empty.Empty), nil
}

func (gs *grpcServer) Identify(ctx context.Context, req *policies.Key) (*policies.ThingID, error) {
	_, res, err := gs.identify.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}

	return res.(*policies.ThingID), nil
}

func decodeCanAccessByKeyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*policies.AccessByKeyReq)
	return accessByKeyReq{thingKey: req.GetToken(), chanID: req.GetChanID()}, nil
}

func decodeCanAccessByIDRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*policies.AccessByIDReq)
	return accessByIDReq{thingID: req.GetThingID(), chanID: req.GetChanID()}, nil
}

func decodeIsChannelOwnerRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*policies.ChannelOwnerReq)
	return channelOwnerReq{owner: req.GetOwner(), chanID: req.GetChanID()}, nil
}

func decodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*policies.Key)
	return identifyReq{key: req.GetValue()}, nil
}

func encodeIdentityResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &policies.ThingID{Value: res.id}, nil
}

func encodeEmptyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(emptyRes)
	return &empty.Empty{}, encodeError(res.err)
}

func encodeError(err error) error {
	switch {
	case errors.Contains(err, nil):
		return nil
	case errors.Contains(err, errors.ErrMalformedEntity),
		err == apiutil.ErrInvalidAuthKey,
		err == apiutil.ErrMissingID,
		err == apiutil.ErrMissingPolicySub,
		err == apiutil.ErrMissingPolicyObj,
		err == apiutil.ErrMissingPolicyAct,
		err == apiutil.ErrMalformedPolicy,
		err == apiutil.ErrMissingPolicyOwner,
		err == apiutil.ErrHigherPolicyRank:
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Contains(err, errors.ErrAuthentication),
		err == apiutil.ErrBearerToken:
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Contains(err, errors.ErrAuthorization):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

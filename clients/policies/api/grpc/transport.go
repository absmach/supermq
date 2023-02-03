package grpc

import (
	"context"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux/clients/policies"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ policies.AuthServiceServer = (*grpcServer)(nil)

type grpcServer struct {
	authorize kitgrpc.Handler
	policies.UnimplementedAuthServiceServer
}

// NewServer returns new AuthServiceServer instance.
func NewServer(svc policies.Service) policies.AuthServiceServer {
	return &grpcServer{
		authorize: kitgrpc.NewServer(
			otelkit.EndpointMiddleware(otelkit.WithOperation("authorize"))(authorizeEndpoint(svc)),
			decodeAuthorizeRequest,
			encodeAuthorizeResponse,
		),
	}
}

func (s *grpcServer) Authorize(ctx context.Context, req *policies.AuthorizeReq) (*policies.AuthorizeRes, error) {
	_, res, err := s.authorize.ServeGRPC(ctx, req)
	if err != nil {
		return nil, encodeError(err)
	}
	return res.(*policies.AuthorizeRes), nil
}

func decodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*policies.AuthorizeReq)
	return authReq{Act: req.GetAct(), Obj: req.GetObj(), Sub: req.GetSub(), EntityType: req.GetEntityType()}, nil
}

func encodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(authorizeRes)
	return &policies.AuthorizeRes{Authorized: res.authorized}, nil
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

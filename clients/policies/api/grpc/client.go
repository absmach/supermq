package grpc

import (
	"context"
	"time"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux/clients/policies"
	opentracing "github.com/opentracing/opentracing-go"
	"go.opentelemetry.io/contrib/instrumentation/github.com/go-kit/kit/otelkit"
	"google.golang.org/grpc"
)

const svcName = "policies.AuthService"

var _ policies.AuthServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	authorize endpoint.Endpoint
	timeout   time.Duration
}

// NewClient returns new gRPC client instance.
func NewClient(tracer opentracing.Tracer, conn *grpc.ClientConn, timeout time.Duration) policies.AuthServiceClient {
	return &grpcClient{
		authorize: otelkit.EndpointMiddleware(otelkit.WithOperation("authorize"))(kitgrpc.NewClient(
			conn,
			svcName,
			"Authorize",
			encodeAuthorizeRequest,
			decodeAuthorizeResponse,
			policies.AuthorizeRes{},
		).Endpoint()),

		timeout: timeout,
	}
}

func (client grpcClient) Authorize(ctx context.Context, req *policies.AuthorizeReq, _ ...grpc.CallOption) (r *policies.AuthorizeRes, err error) {
	ctx, close := context.WithTimeout(ctx, client.timeout)
	defer close()

	res, err := client.authorize(ctx, authReq{Act: req.GetAct(), Obj: req.GetObj(), Sub: req.GetSub(), EntityType: req.GetEntityType()})
	if err != nil {
		return &policies.AuthorizeRes{}, err
	}

	ar := res.(authorizeRes)
	return &policies.AuthorizeRes{Authorized: ar.authorized}, err
}

func decodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*policies.AuthorizeRes)
	return authorizeRes{authorized: res.Authorized}, nil
}

func encodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(authReq)
	return &policies.AuthorizeReq{
		Sub:        req.Sub,
		Obj:        req.Obj,
		Act:        req.Act,
		EntityType: req.EntityType,
	}, nil
}

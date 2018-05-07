package grpc

import (
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _ UsersServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	identify endpoint.Endpoint
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn) UsersServiceClient {
	endpoint := kitgrpc.NewClient(
		conn,
		"grpc.UsersService",
		"Identify",
		encodeIdentifyRequest,
		decodeIdentifyResponse,
		Identity{},
	).Endpoint()

	return grpcClient{endpoint}
}

func (client grpcClient) Identify(ctx context.Context, token *Token, _ ...grpc.CallOption) (*Identity, error) {
	res, err := client.identify(ctx, identityReq{token.GetValue()})
	if err != nil {
		return nil, err
	}

	ir := res.(identityRes)
	return &Identity{ir.id}, ir.err
}

func encodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(identityReq)
	return &Token{req.token}, nil
}

func decodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*Identity)
	return identityRes{res.GetValue(), nil}, nil
}

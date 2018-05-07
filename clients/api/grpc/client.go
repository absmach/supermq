package grpc

import (
	"fmt"

	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _ ClientsServiceClient = (*grpcClient)(nil)

type grpcClient struct {
	canAccess endpoint.Endpoint
}

// NewClient returns new gRPC client instance.
func NewClient(conn *grpc.ClientConn) ClientsServiceClient {
	endpoint := kitgrpc.NewClient(
		conn,
		"grpc.ClientsService",
		"CanAccess",
		encodeCanAccessRequest,
		decodeCanAccessResponse,
		Identity{},
	).Endpoint()

	return grpcClient{endpoint}
}

func (client grpcClient) CanAccess(ctx context.Context, req *AccessReq, _ ...grpc.CallOption) (*Identity, error) {
	res, err := client.canAccess(ctx, accessReq{req.GetClientKey(), req.GetChanID()})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	ar := res.(accessRes)
	return &Identity{ar.id}, ar.err
}

func encodeCanAccessRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(accessReq)
	return &AccessReq{req.clientKey, req.chanID}, nil
}

func decodeCanAccessResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*Identity)
	return accessRes{res.GetValue(), nil}, nil
}

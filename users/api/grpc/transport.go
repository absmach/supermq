package grpc

import (
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/mainflux/mainflux/users"
	context "golang.org/x/net/context"
)

type grpcServer struct {
	identify kitgrpc.Handler
}

// NewServer returns new UsersServiceServer instance.
func NewServer(svc users.Service) UsersServiceServer {
	return &grpcServer{
		kitgrpc.NewServer(
			identifyEndpoint(svc),
			decodeIdentifyRequest,
			encodeIdentifyResponse,
		),
	}
}

func (s *grpcServer) Identify(ctx context.Context, token *Token) (*Identity, error) {
	_, res, err := s.identify.ServeGRPC(ctx, token)
	if err != nil {
		return nil, err
	}
	return res.(*Identity), nil
}

func decodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*Token)
	return identityReq{req.GetValue()}, nil
}

func encodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(identityRes)
	return &Identity{res.id}, res.failed()
}
